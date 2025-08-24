// pkg/scheduler/resource_manager.go
package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/uber-go/zap"
	"sync"
	"time"

	gocmp "github.com/google/go-cmp/cmp"
	"go.opentelemetry.io/otel"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/turtacn/agenticai/internal/constants"
	e "github.com/turtacn/agenticai/internal/errors"
	"github.com/turtacn/agenticai/internal/logger"
	apiv1 "github.com/turtacn/agenticai/pkg/apis"
)

// ------- 内存账本结构 -------
type resourceRecord struct {
	Allocatable corev1.ResourceList
	Reserved    corev1.ResourceList
	LastSeen    time.Time
}

// ------- ResourceManager 接口 -------
type ResourceManager interface {
	// 查询
	PredicateAgents(ctx context.Context, task *apiv1.Task) ([]*AgentSnapshot, error)

	// 预留/释放
	Reserve(ctx context.Context, agent string, task *apiv1.Task) error
	Release(ctx context.Context, agent string, task *apiv1.Task) error

	// 常驻同步协程
	SyncLoop(stop <-chan struct{})
}

// ------- 实现 -------
type manager struct {
	kube client.Client

	mu sync.RWMutex
	// key = agent.name
	table map[string]*resourceRecord

	notifyChan chan struct{} // trigger immediate sync
}

// ------- AgentSnapshot ------
type AgentSnapshot struct {
	Name        string
	Allocatable corev1.ResourceList
	Reserved    corev1.ResourceList
}

// ------- New -------
func NewResourceManager(kube client.Client) ResourceManager {
	return &manager{
		kube:       kube,
		table:      make(map[string]*resourceRecord),
		notifyChan: make(chan struct{}, 1),
	}
}

// ------- 查询过滤 -------
func (m *manager) PredicateAgents(ctx context.Context, task *apiv1.Task) ([]*AgentSnapshot, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var out []*AgentSnapshot
	for k, rec := range m.table {
		// CPU Mem GPU 检查
		avail := subResource(rec.Allocatable, rec.Reserved)
		if needSatisfied(avail, task.Spec.Resources.Limits) {
			out = append(out, &AgentSnapshot{
				Name:        k,
				Allocatable: rec.Allocatable,
				Reserved:    rec.Reserved,
			})
		}
	}
	return out, nil
}

// ------- Reserve -------
func (m *manager) Reserve(ctx context.Context, agent string, task *apiv1.Task) error {
	return m.mutate(ctx, agent, func(rec *resourceRecord) error {
		newReserved := addResource(rec.Reserved, task.Spec.Resources.Limits)
		if !needSatisfied(rec.Allocatable, newReserved) {
			return e.E(e.KindConflict, fmt.Errorf("agent %s capacity exceeded", agent))
		}
		rec.Reserved = newReserved
		return nil
	})
}

// ------- Release -------
func (m *manager) Release(ctx context.Context, agent string, task *apiv1.Task) error {
	return m.mutate(ctx, agent, func(rec *resourceRecord) error {
		rec.Reserved = subResource(rec.Reserved, task.Spec.Resources.Limits)
		if rec.Reserved == nil {
			rec.Reserved = corev1.ResourceList{}
		}
		return nil
	})
}

// ------- 内部 mutate 流程 -------
func (m *manager) mutate(ctx context.Context, agent string, fn func(*resourceRecord) error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	rec, ok := m.table[agent]
	if !ok {
		return e.E(e.KindNotFound, fmt.Sprintf("agent %s not found in cache", agent))
	}
	if err := fn(rec); err != nil {
		return err
	}
	// trigger async persistence
	select {
	case m.notifyChan <- struct{}{}:
	default:
	}
	return nil
}

// ------- 账本同步协程 -------
func (m *manager) SyncLoop(stop <-chan struct{}) {
	ctx := context.Background()
	tick := time.NewTicker(30 * time.Second)
	defer tick.Stop()

	// 首次立即加载
	m.syncFromApiserver(ctx)

	for {
		select {
		case <-stop:
			return
		case <-tick.C:
		case <-m.notifyChan:
		}
		m.syncFromApiserver(ctx)
	}
}

func (m *manager) syncFromApiserver(ctx context.Context) {
	tracer := otel.Tracer("manager")
	_, span := tracer.Start(ctx, "syncFromApiserver")
	defer span.End()
	log := logger.WithCtx(ctx)

	agentList := &apiv1.AgentList{}
	if err := m.kube.List(ctx, agentList, client.Limit(512)); err != nil {
		log.Error("failed to list agents for cache sync", zap.Error(err))
		return
	}

	now := time.Now()
	m.mu.Lock()
	defer m.mu.Unlock()

	newTable := make(map[string]*resourceRecord)

	for _, a := range agentList.Items {
		var allocatable corev1.ResourceList
		if a.Status.DesiredReplicas > 0 && len(a.Status.Conditions) > 0 && readyCondTrue(a.Status.Conditions) {
			allocatable = corev1.ResourceList{
				corev1.ResourceCPU:    *resource.NewScaledQuantity(1000, resource.Milli),
				corev1.ResourceMemory: *resource.NewScaledQuantity(512, resource.Mega),
			}
		} else {
			continue
		}

		key := fmt.Sprintf("%s/%s", a.Namespace, a.Name)
		// keep reservation if already exists
		var reserved corev1.ResourceList
		if old, ok := m.table[key]; ok {
			reserved = old.Reserved.DeepCopy()
		} else {
			reserved = corev1.ResourceList{}
		}

		newTable[key] = &resourceRecord{
			Allocatable: allocatable,
			Reserved:    reserved,
			LastSeen:    now,
		}
	}
	// atomic swap
	m.table = newTable
}

// ------- utils -------
func needSatisfied(allocatable, required corev1.ResourceList) bool {
	for k, q := range required {
		av, ok := allocatable[k]
		if !ok || av.Cmp(q) < 0 {
			return false
		}
	}
	return true
}

func addResource(a, b corev1.ResourceList) corev1.ResourceList {
	out := a.DeepCopy()
	for k, v := range b {
		out[k].Add(v)
	}
	return out
}

func subResource(a, b corev1.ResourceList) corev1.ResourceList {
	out := a.DeepCopy()
	for k, v := range b {
		sum := out[k].DeepCopy()
		sum.Sub(v)
		if sum.Sign() < 0 {
			sum.Set(resource.MustParse("0"))
		}
		out[k] = sum
	}
	return out
}

func readyCondTrue(conds []apiv1.AgentCondition) bool {
	for _, c := range conds {
		if c.Type == "Available" && c.Status == "True" {
			return true
		}
	}
	return false
}
//Personal.AI order the ending
