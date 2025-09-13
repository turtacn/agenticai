// internal/config/config.go
package config

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/turtacn/agenticai/internal/constants"
	"github.com/turtacn/agenticai/internal/errors"
	"github.com/turtacn/agenticai/internal/logger"
)

// Config 是所有组件统一读取的结构
type Config struct {
	// 顶层
	ProjectName string `mapstructure:"project_name" json:"project_name"` // 仅输出用
	Version     string // 运行时填充
	Mode        string `mapstructure:"mode"` // development|production|test

	// Server
	Server Server `mapstructure:"server"`
	// Logging
	Log Log `mapstructure:"log"`
	// Kubernetes Controller
	K8sController K8sController `mapstructure:"k8s_controller"`
	// Observability
	Observability Observability `mapstructure:"observability"`
	// Security
	Security Security `mapstructure:"security"`
	// Sandbox
	Sandbox Sandbox `mapstructure:"sandbox"`
	// Storage
	Storage Storage `mapstructure:"storage"`
}

// ---------- 子结构 ----------
type Server struct {
	HTTPPort        int           `mapstructure:"http_port" json:"http_port"`
	MetricsPort     int           `mapstructure:"metrics_port" json:"metrics_port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout" json:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout" json:"write_timeout"`
	GracefulTimeout time.Duration `mapstructure:"graceful_timeout" json:"graceful_timeout"`
}

type Log struct {
	Level  string `mapstructure:"level"`  // debug/info/warn/error
	Output string `mapstructure:"output"` // stdout/file
}

type K8sController struct {
	WorkerThreads    int    `mapstructure:"worker_threads"` // 默认 4
	LeaderElectionID string `mapstructure:"leader_election_id"`
	WatchNamespace   string `mapstructure:"watch_namespace"` // ""  = all
}

type Observability struct {
	JaegerURL      string  `mapstructure:"jaeger_url"`
	MetricsEnabled bool    `mapstructure:"metrics_enabled"`
	MetricsPath    string  `mapstructure:"metrics_path"`   // 默认 /metrics
	TraceSampling  float64 `mapstructure:"trace_sampling"` // 0-1
}

type Security struct {
	TrustDomain     string `mapstructure:"trust_domain"`     // SPIFFE
	KeyStoreBackend string `mapstructure:"keystore_backend"` // k8s/vault
}

type Sandbox struct {
	Type         string            `mapstructure:"type"` // gvisor/kata/firecracker
	CPULimit     string            `mapstructure:"cpu_limit"`
	MemoryLimit  string            `mapstructure:"memory_limit"`
	ExtraSysctls map[string]string `mapstructure:"extra_sysctls"` // 高级可调
}

type Storage struct {
	Type   string                 `mapstructure:"type"`   // local/s3/gcp/abs/minio/etc
	Config map[string]interface{} `mapstructure:"config"` // 具体 backend map
}

type GatewayConfig struct {
	Listen string `mapstructure:"listen"`
}

// ---------- 单例 + 锁 ----------
var (
	conf Config
	mu   sync.RWMutex
	v    *viper.Viper
)

// Init 必须且仅能在程序启动时调用一次（main()->init→Init）
func Init(modeOpt ...string) error {
	mu.Lock()
	defer mu.Unlock()

	// 默认覆盖顺序：default → env → mode 文件
	v = viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	// 路径优先级：./configs/default -> ./configs/<mode>
	v.AddConfigPath("./configs/default")
	if len(modeOpt) > 0 && modeOpt[0] != "" {
		v.AddConfigPath("./configs/" + modeOpt[0])
	}
	v.AutomaticEnv()
	v.SetEnvPrefix("AIAI") // AIAI_SERVER_HTTP_PORT=xxxx
	// 替换句点为下划线
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 默认值
	setDefaults()

	if err := v.ReadInConfig(); err != nil {
		logger.Error(nil, "failed to read default config", zap.Error(err))
		return errors.E(errors.KindInternal, err, "config load failure")
	}

	if err := v.Unmarshal(&conf); err != nil {
		logger.Error(nil, "failed to unmarshal config", zap.Error(err))
		return errors.E(errors.KindInternal, err, "config parse failure")
	}

	// 补字段
	if conf.Version == "" {
		conf.Version = constants.Version
	}
	if conf.ProjectName == "" {
		conf.ProjectName = constants.ProjectName
	}

	// 校验
	if err := validate(&conf); err != nil {
		return err
	}

	return nil
}

// setDefaults 与 constants 保持一致
func setDefaults() {
	v.SetDefault("mode", "development")
	v.SetDefault("server.http_port", constants.DefaultHTTPPort)
	v.SetDefault("server.metrics_port", constants.DefaultMetricsPort)
	v.SetDefault("server.read_timeout", constants.DefaultTimeout)
	v.SetDefault("server.write_timeout", constants.DefaultTimeout*2)
	v.SetDefault("server.graceful_timeout", time.Second*15)

	v.SetDefault("log.level", constants.DefaultLogLevel.String())
	v.SetDefault("log.output", "stdout")

	v.SetDefault("k8s_controller.worker_threads", constants.ControllerWorkerThreads)
	v.SetDefault("k8s_controller.leader_election_id", fmt.Sprintf("%s-leader", constants.ProjectName))
	v.SetDefault("k8s_controller.watch_namespace", metav1.NamespaceAll)

	v.SetDefault("observability.metrics_enabled", true)
	v.SetDefault("observability.metrics_path", "/metrics")
	v.SetDefault("observability.trace_sampling", 0.1)

	v.SetDefault("security.trust_domain", constants.TrustDomain)
	v.SetDefault("security.keystore_backend", "k8s")

	v.SetDefault("sandbox.type", "gvisor")
	v.SetDefault("sandbox.cpu_limit", constants.DefaultSandboxCPU)
	v.SetDefault("sandbox.memory_limit", constants.DefaultSandboxMemory)
}

// validate 校验逻辑写死在此处，后期可抽接口
func validate(c *Config) error {
	if _, ok := map[string]struct{}{
		"development": {},
		"production":  {},
		"test":        {},
	}[c.Mode]; !ok {
		return errors.E(errors.KindValidation, fmt.Sprintf("invalid mode %q", c.Mode))
	}
	if c.Log.Level != "" {
		if _, ok := map[string]struct{}{
			"debug": {}, "info": {}, "warn": {}, "error": {},
		}[string(c.Log.Level)]; !ok {
			return errors.E(errors.KindValidation, fmt.Sprintf("invalid log level %q", c.Log.Level))
		}
	}
	return nil
}

// Get 读配置线程安全
func Get() Config {
	mu.RLock()
	defer mu.RUnlock()
	return conf
}

// Watch 注册运行时热重载钩子
// callback(新Config) ，返回 true 则接受，false 丢弃并重载文件
func Watch(callback func(Config) bool) {
	go func() {
		for {
			v.WatchConfig()
			v.OnConfigChange(func(e fsnotify.Event) {
				var newConf Config
				if err := v.Unmarshal(&newConf); err != nil {
					logger.Warn(nil, "config changed but unmarshal failed", zap.Error(err))
					return
				}
				mu.Lock()
				defer mu.Unlock()
				if callback(newConf) {
					conf = newConf
					logger.Info(nil, "config reloaded successfully")
				}
			})
			// 阻塞等待，防止快速循环
			time.Sleep(5 * time.Second)
		}
	}()
}

//Personal.AI order the ending
