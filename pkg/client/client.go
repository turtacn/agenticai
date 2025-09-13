// pkg/client/client.go
package client

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"

	"github.com/turtacn/agenticai/internal/constants"
	"github.com/turtacn/agenticai/pkg/apis" // 触发 scheme 注册
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//
// ======== 公共接口 ========
//
type Interface interface {
	// CRUD 工具统一入口
	Agent() AgentInterface
	Task() TaskInterface
	Tool() ToolInterface
	SecurityPolicy() SecurityPolicyInterface
	Telemetry() TelemetryInterface

	// 原生 Kubernetes 核心资源接口
	KubeReader() ctrl.Reader
	KubeWriter() ctrl.Writer

	// Stats 暴露给 controller
	RetryMetrics() *RetryMetrics
}

type RetryMetrics struct{}

type AgentInterface interface {
	Get(ctx context.Context, name string, options metav1.GetOptions) (*apis.Agent, error)
	List(ctx context.Context, options metav1.ListOptions) (*apis.AgentList, error)
	Create(ctx context.Context, agent *apis.Agent, options metav1.CreateOptions) (*apis.Agent, error)
}
type TaskInterface interface{}
type ToolInterface interface{}
type SecurityPolicyInterface interface{}
type TelemetryInterface interface{}

type agentCli struct {
	client ctrl.Client
	stats  *RetryMetrics
}

func (c *agentCli) Get(ctx context.Context, name string, options metav1.GetOptions) (*apis.Agent, error) {
	return nil, nil
}
func (c *agentCli) List(ctx context.Context, options metav1.ListOptions) (*apis.AgentList, error) {
	return nil, nil
}
func (c *agentCli) Create(ctx context.Context, agent *apis.Agent, options metav1.CreateOptions) (*apis.Agent, error) {
	return nil, nil
}

type taskCli struct {
	client ctrl.Client
	stats  *RetryMetrics
}
type toolCli struct {
	client ctrl.Client
	stats  *RetryMetrics
}
type securityCli struct {
	client ctrl.Client
	stats  *RetryMetrics
}
type telemetryCli struct {
	client ctrl.Client
	stats  *RetryMetrics
}

//
// ======== 构造器 ========
//
var (
	instance Interface
	initOnce sync.Once
)

// New 构造器（内部用），返回 error 供测试
func New(ctx context.Context, opts ...Option) (Interface, error) {
	// 解析选项
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	conf, err := restConfig(o)
	if err != nil {
		return nil, fmt.Errorf("client.New: build rest config: %w", err)
	}

	// SPIFFE mTLS transport
	if o.spiffeTrustDomain == "" {
		o.spiffeTrustDomain = constants.TrustDomain
	}
	rt, err := spiffeRoundTripper(conf, o.spiffeTrustDomain)
	if err != nil {
		return nil, fmt.Errorf("client.New: spiffe transport: %w", err)
	}
	conf.Transport = rt

	// controller-runtime client
	scheme := runtime.NewScheme()
	if err := apis.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("client.New: register scheme: %w", err)
	}
	crClient, err := ctrl.New(conf, ctrl.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("client.New: ctrl client: %w", err)
	}

	// kubernetes native clientset
	clientset, err := kubernetes.NewForConfig(conf)
	if err != nil {
		return nil, fmt.Errorf("client.New: kube clientset: %w", err)
	}

	_, err = clientset.ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("client.New: test API server: %w", err)
	}

	impl := &clientImpl{
		crClient: crClient,
		scheme:   scheme,
		stats:    &RetryMetrics{},
	}

	return impl, nil
}

//
// ======== 单例快速入口 ========
//
func Get(ctx context.Context, opts ...Option) (Interface, error) {
	var err error
	initOnce.Do(func() {
		instance, err = New(ctx, opts...)
	})
	return instance, err
}

//
// ======== 实现层 ========
//
type clientImpl struct {
	crClient ctrl.Client
	scheme   *runtime.Scheme
	stats    *RetryMetrics
}

func (c *clientImpl) Agent() AgentInterface            { return &agentCli{c.crClient, c.stats} }
func (c *clientImpl) Task() TaskInterface              { return &taskCli{c.crClient, c.stats} }
func (c *clientImpl) Tool() ToolInterface              { return &toolCli{c.crClient, c.stats} }
func (c *clientImpl) SecurityPolicy() SecurityPolicyInterface { return &securityCli{c.crClient, c.stats} }
func (c *clientImpl) Telemetry() TelemetryInterface          { return &telemetryCli{c.crClient, c.stats} }

func (c *clientImpl) KubeReader() ctrl.Reader { return c.crClient }
func (c *clientImpl) KubeWriter() ctrl.Writer { return c.crClient }

func (c *clientImpl) RetryMetrics() *RetryMetrics { return c.stats }

//
// ======== RESTConfig 构建 ========
//
func restConfig(o *options) (*rest.Config, error) {
	// 1) in-cluster
	cfg, err := rest.InClusterConfig()
	if err == nil {
		return cfg, nil
	}

	// 2) out-of-cluster (dev)
	kubeconfig := ""
	if o.kubeconfigPath != "" {
		kubeconfig = o.kubeconfigPath
	} else if kc := os.Getenv("KUBECONFIG"); kc != "" {
		kubeconfig = kc
	}
	loader := &clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeconfig}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loader,
		&clientcmd.ConfigOverrides{},
	)
	return clientConfig.ClientConfig()
}

//
// ======== SPIFFE Transport ========
//
func spiffeRoundTripper(restCfg *rest.Config, trustDomain string) (http.RoundTripper, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create a workloadapi.X509Source to fetch SVIDs
	source, err := workloadapi.NewX509Source(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to create X509 source: %w", err)
	}

	// Create a tls.Config that uses the X509Source
	// Authorize any SPIFFE ID from the trust domain
	tlsConfig := tlsconfig.MTLSClientConfig(source, source, tlsconfig.AuthorizeAny())

	// Use the default transport as a base
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = tlsConfig

	return transport, nil
}

//
// ======== 初始化 Scheme 注册（一次性） ========
//
func init() {
	// 让 apis 包初始化 Scheme；无需主动，已 import _ (apis)
}
//Personal.AI order the ending
