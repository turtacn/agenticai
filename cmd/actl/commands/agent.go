// cmd/actl/commands/agent.go
package commands

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/turtacn/agenticai/pkg/types"
	"github.com/turtacn/agenticai/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// NewAgentCmd 聚合所有与 agent 相关的子命令
func NewAgentCmd(kubeCfg, defaultNS string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "agent",
		Aliases: []string{"agents"},
		Short:   "Manage AI agents",
	}
	cmd.PersistentFlags().StringP("namespace", "n", defaultNS, "target namespace")

	cmd.AddCommand(
		agentCreateCmd(kubeCfg),
		agentListCmd(kubeCfg),
		agentDeleteCmd(kubeCfg),
		agentLogsCmd(kubeCfg),
	)
	return cmd
}

/* -------------------- create -------------------- */
func agentCreateCmd(kubeCfg string) *cobra.Command {
	var cpu, mem string
	var replicas int
	cmd := &cobra.Command{
		Use:   "create NAME [IMAGE]",
		Short: "Deploy a new agent",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ns, _ := cmd.Flags().GetString("namespace")
			name := args[0]
			image := "ghcr.io/turtacn/agentic/agent:latest"
			if len(args) == 2 {
				image = args[1]
			}

			cs, err := utils.ClientFromKubeConfig(kubeCfg)
			if err != nil {
				return fmt.Errorf("k8s client: %w", err)
			}

			pod := &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name: name,
					Labels: map[string]string{
						"app": "agentic-agent",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:            "agent",
						Image:           image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(cpu),
								corev1.ResourceMemory: resource.MustParse(mem),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(cpu),
								corev1.ResourceMemory: resource.MustParse(mem),
							},
						},
						Env: []corev1.EnvVar{{
							Name: "AGENT_NAME", Value: name,
						}},
					}},
				},
			}

			if err := utils.CreateOrUpdate(cmd.Context(), cs, ns, pod); err != nil {
				return fmt.Errorf("deploy agent: %w", err)
			}
			fmt.Printf("🚀 agent/%s created in namespace %s\n", name, ns)
			return nil
		},
	}
	cmd.Flags().StringVar(&cpu, "cpu", "100m", "cpu request/limit")
	cmd.Flags().StringVar(&mem, "mem", "128Mi", "memory request/limit")
	return cmd
}

/* -------------------- list -------------------- */
func agentListCmd(kubeCfg string) *cobra.Command {
	return &cobra.Command{
		Use:   "list [MATCH]",
		Short: "List running agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			ns, _ := cmd.Flags().GetString("namespace")
			cs, err := utils.ClientFromKubeConfig(kubeCfg)
			if err != nil {
				return err
			}
			sel := "app=agentic-agent"
			if len(args) > 0 {
				sel = fmt.Sprintf("app=agentic-agent,name=%s", args[0])
			}
			list, err := utils.GetPodByLabel(ctx, cs, ns, sel)
			if err != nil {
				return err
			}
			fmt.Printf("%-20s %-12s %-8s %-20s\n", "NAME", "STATUS", "AGE", "IMAGE")
			for _, p := range list.Items {
				img := "<none>"
				if len(p.Spec.Containers) > 0 {
					img = p.Spec.Containers[0].Image
				}
				age := time.Since(p.CreationTimestamp.Time).Round(time.Second)
				fmt.Printf("%-20s %-12s %-8s %-20s\n", p.Name, string(p.Status.Phase), age, img)
			}
			return nil
		},
	}
}

/* -------------------- delete -------------------- */
func agentDeleteCmd(kubeCfg string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Remove agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			ns, _ := cmd.Flags().GetString("namespace")
			cs, err := utils.ClientFromKubeConfig(kubeCfg)
			if err != nil {
				return err
			}
			if err := utils.DeletePod(ctx, cs, ns, args[0]); err != nil {
				return err
			}
			fmt.Printf("🗑 agent/%s deleted\n", args[0])
			return nil
		},
	}
	return cmd
}

/* -------------------- logs -------------------- */
func agentLogsCmd(kubeCfg string) *cobra.Command {
	var tail int64 = 100
	var follow bool
	cmd := &cobra.Command{
		Use:   "logs NAME",
		Short: "View agent logs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			ns, _ := cmd.Flags().GetString("namespace")
			name := args[0]

			cs, err := utils.ClientFromKubeConfig(kubeCfg)
			if err != nil {
				return err
			}
			req := cs.CoreV1().Pods(ns).GetLogs(name, &corev1.PodLogOptions{
				Follow: follow,
				TailLines: &tail,
			})
			stream, err := req.Stream(ctx)
			if err != nil {
				return err
			}
			defer stream.Close()

			sig := make(chan os.Signal, 1)
			signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				<-sig
				stream.Close()
			}()
			_, _ = io.Copy(os.Stdout, stream)
			return nil
		},
	}
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "follow logs")
	cmd.Flags().Int64Var(&tail, "tail", 100, "lines to display")
	return cmd
}

// utils.ClientFromKubeConfig 内部用
func ClientFromKubeConfig(cfgPath string) (*kubernetes.Clientset, error) {
	return utils.ClientFromKubeConfig(cfgPath)
}
//Personal.AI order the ending
