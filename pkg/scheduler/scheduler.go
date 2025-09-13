// pkg/scheduler/scheduler.go
package scheduler

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/turtacn/agenticai/internal/logger"
	agenticaiov1 "github.com/turtacn/agenticai/pkg/apis/agenticai.io/v1"
)

//
// Interface
//
type Scheduler interface {
	// Schedule schedules one task once.
	Schedule(context.Context, *agenticaiov1.Task) (*ScheduleResult, error)
}

//
// ScheduleResult
//
type ScheduleResult struct {
	TargetAgent string
}

//
// defaultScheduler
//
type defaultScheduler struct {
	kube client.Client
	rm   ResourceManager // 共享的缓存
	// scorer  ScoreFunc // TODO: Implement scoring
}

type ScoreFunc func(task *agenticaiov1.Task, agent *AgentSnapshot) int

// Provide default impl
func NewDefault(kube client.Client) Scheduler {
	return New(kube, nil) // TODO: Implement scoring
}

func New(kube client.Client, scorer ScoreFunc) Scheduler {
	return &defaultScheduler{
		kube: kube,
		rm:   NewResourceManager(kube),
		// scorer:  scorer,
	}
}

//
// Schedule entry
//
func (s *defaultScheduler) Schedule(ctx context.Context, task *agenticaiov1.Task) (*ScheduleResult, error) {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().
		Tracer("pkg/scheduler").
		Start(ctx, "Schedule")
	defer span.End()
	log := logger.WithCtx(ctx)

	// 快速过滤候选 Agents
	candidates, err := s.rm.PredicateAgents(ctx, task)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no schedulable agents")
	}

	// 打分排序
	// sort.Slice(candidates, func(i, j int) bool {
	// 	return s.scorer(task, candidates[i]) > s.scorer(task, candidates[j])
	// })
	best := candidates[0]

	// 写回 reserved capacity
	if err := s.rm.Reserve(ctx, best.Name, task); err != nil {
		return nil, fmt.Errorf("reserve failed: %w", err)
	}

	log.Info("scheduled task", zap.String("task", task.Namespace+"/"+task.Name),
		zap.String("agent", best.Name))
	return &ScheduleResult{TargetAgent: best.Name}, nil
}
//Personal.AI order the ending
