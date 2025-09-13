// cmd/actl/commands/root.go
package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

// NewRootCmd 构建 actl 根命令
func NewRootCmd(version, buildDate, commitSHA string) *cobra.Command {
	var cfgFile string
	var kubeCfg, namespace string

	root := &cobra.Command{
		Use:           "actl",
		Short:         "AgenticAI CLI – manage your AI agents at scale",
		Long:          `actl is the official command-line tool for AgenticAI Platform.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       fmt.Sprintf("%s (built %s, sha:%s)", version, buildDate, commitSHA),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return nil
		},
	}
	// global
	flags := root.PersistentFlags()
	flags.StringVar(&cfgFile, "config", "", "config file (default $HOME/.actl.yaml)")
	flags.StringVar(&kubeCfg, "kubeconfig", defaultKubeConfig(), "path to kubeconfig")
	flags.StringVar(&namespace, "namespace", "default", "target namespace for k8s resources")

	// 子命令
	root.AddCommand(
		NewAgentCmd(kubeCfg, namespace),
		NewTaskCmd(kubeCfg, namespace),
		NewClusterCmd(kubeCfg),
		newVersionCmd(version),
		newCompletionCmd(),
	)

	return root
}

func defaultKubeConfig() string {
	if k := os.Getenv("KUBECONFIG"); k != "" {
		return k
	}
	return homedir.HomeDir() + "/.kube/config"
}

// newCompletionCmd 生成 completion 子命令
func newCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion script",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return fmt.Errorf("shell name required")
			}
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell %q", args[0])
			}
		},
	}
}

func newVersionCmd(v string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "show version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("actl %s\n", v)
		},
	}
}
//Personal.AI order the ending
