// pkg/client/options.go
package client

type options struct {
	kubeconfigPath    string
	spiffeTrustDomain string
}

type Option func(*options)

// WithKubeConfigPath 指定 kubeconfig 路径，优先级高于环境变量
func WithKubeConfigPath(path string) Option {
	return func(o *options) { o.kubeconfigPath = path }
}

// WithTrustDomain 设置 SPIFFE TrustDomain，默认为 "agenticai.io"
func WithTrustDomain(domain string) Option {
	return func(o *options) { o.spiffeTrustDomain = domain }
}
//Personal.AI order the ending
