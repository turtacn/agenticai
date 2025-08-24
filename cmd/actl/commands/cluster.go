// cmd/actl/commands/cluster.go
package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/turtacn/agenticai/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewClusterCmd(kubeCfg string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cluster",
		Aliases: []string{"c"},
		Short:   "Manage AgenticAI cluster",
	}

	cmd.AddCommand(
		clusterInitCmd(kubeCfg),
		clusterStatusCmd(kubeCfg),
		clusterDiagnoseCmd(kubeCfg),
	)
	return cmd
}

/* -------------------- init -------------------- */
func clusterInitCmd(kubeCfg string) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize cluster prerequisites",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("🔧 Verifying cluster prerequisites...")
			cs, err := utils.ClientFromKubeConfig(kubeCfg)
			if err != nil {
				return err
			}
			if err := utils.ClusterHealth(cmd.Context(), cs); err != nil {
				return fmt.Errorf("cluster NOT ready: %w", err)
			}
			fmt.Println("✅ K8s API reachable")

			// 1. 检查命名空间
			nsCli := cs.CoreV1().Namespaces()
			if _, err := nsCli.Get(cmd.Context(), "agentic", metav1.GetOptions{}); err == nil {
				fmt.Println("✅ namespace 'agentic' already exists")
			} else {
				if _, err := nsCli.Create(cmd.Context(), &corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{Name: "agentic"},
				}, metav1.CreateOptions{}); err != nil {
					return fmt.Errorf("create namespace: %w", err)
				}
				fmt.Println("✅ created namespace 'agentic'")
			}

			// 2. 检查权限（简化：尝试创建 Job 模板）
			if _, err := cs.BatchV1().CronJobs("agentic").
				List(cmd.Context(), metav1.ListOptions{Limit: 1}); err != nil {
				return fmt.Errorf("need cluster-admin: %w", err)
			}
			fmt.Println("✅ RBAC permission granted")
			fmt.Println("🎉 Cluster ready for AgenticAI!")
			return nil
		},
	}
}

/* -------------------- status -------------------- */
func clusterStatusCmd(kubeCfg string) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Current cluster overview",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cs, err := utils.ClientFromKubeConfig(kubeCfg)
			if err != nil {
				return err
			}

			// Nodes 摘要
			nodes, err := cs.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
			if err != nil {
				return err
			}
			gpu := int32(0)
			for _, n := range nodes.Items {
				if v, ok := n.Status.Capacity["nvidia.com/gpu"]; ok {
					gpu += v.Value()
				}
			}
			fmt.Printf("Nodes    : %d [GPU=%d]\n", len(nodes.Items), gpu)

			// Running Pods
			pods, err := cs.CoreV1().Pods("").List(ctx, metav1.ListOptions{
				LabelSelector: "app=agentic-agent",
			})
			if err != nil {
				return err
			}
			fmt.Printf("Agents   : %d running\n", len(pods.Items))

			// Storage
			_, err = cs.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{Limit: 1})
			if err == nil {
				fmt.Println("Storage  : ✅ provisioned")
			} else {
				fmt.Println("Storage  : ❌ missing")
			}
			return nil
		},
	}
}

/* -------------------- diagnose -------------------- */
func clusterDiagnoseCmd(kubeCfg string) *cobra.Command {
	return &cobra.Command{
		Use:   "diagnose",
		Short: "Run health checks and print report",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cs, err := utils.ClientFromKubeConfig(kubeCfg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ failed to talk to cluster: %v\n", err)
				return nil
			}

			ok := true
			report := func(name string, err error) {
				if err == nil {
					fmt.Printf("✅ %s\n", name)
				} else {
					fmt.Printf("❌ %s: %v\n", name, err)
					ok = false
				}
			}

			// 1. Connectivity
			report("Cluster Reachable", utils.ClusterHealth(ctx, cs))

			// 2. Namespace & RBAC
			_, err = cs.CoreV1().Namespaces().Get(ctx, "agentic", metav1.GetOptions{})
			report("Namespace 'agentic'", err)

			// 3. Nodes Ready
			nodes, err := cs.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
			if err != nil {
				report("Fetch Nodes", err)
			} else {
				for _, n := range nodes.Items {
					ready := false
					for _, c := range n.Status.Conditions {
						if c.Type == "Ready" && c.Status == "True" {
							ready = true
							break
						}
					}
					if !ready {
						fmt.Printf("❌ node %s NotReady\n", n.Name)
						ok = false
					}
				}
			}

			if ok {
				fmt.Println("🔍 Diagnose complete – Cluster is healthy.")
			} else {
				fmt.Println("📋 Please address the issues above.")
			}
			return nil
		},
	}
}
//Personal.AI order the ending
