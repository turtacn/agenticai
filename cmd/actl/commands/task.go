// cmd/actl/commands/task.go
package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/turtacn/agenticai/pkg/types"
	"github.com/turtacn/agenticai/pkg/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewTaskCmd(kubeCfg, defaultNS string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "task",
		Aliases: []string{"tasks"},
		Short:   "Manage agent tasks",
	}
	cmd.PersistentFlags().StringP("namespace", "n", defaultNS, "target namespace")

	cmd.AddCommand(
		taskSubmitCmd(kubeCfg),
		taskStatusCmd(kubeCfg),
		taskCancelCmd(kubeCfg),
		taskListCmd(kubeCfg),
	)
	return cmd
}

/* -------------------- submit -------------------- */
func taskSubmitCmd(kubeCfg string) *cobra.Command {
	var cpu, mem string
	var image, priority string
	var timeout time.Duration
	var commands []string

	cmd := &cobra.Command{
		Use:   "submit [COMMAND...]",
		Short: "Submit a new AI task",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			ns := cmdFlag(cmd, "namespace")

			spec := types.TaskSpec{
				Image:      image,
				Command:    append(args, commands...),
				Resources:  map[string]string{"cpu": cpu, "memory": mem},
				TimeBudget: metav1.Duration{Duration: timeout},
				Priority:   priority,
			}
			task := &types.Task{
				Spec: spec,
				Key:  "task-" + time.Now().UTC().Format("20060102030405"),
			}

			// 通过 configmap 转存任务 spec
			cm := toConfigMap(task, ns)

			cl, err := utils.ClientFromKubeConfig(kubeCfg)
			if err != nil {
				return err
			}
			if _, err := cl.CoreV1().ConfigMaps(ns).Create(ctx, cm, metav1.CreateOptions{}); err != nil {
				return fmt.Errorf("create task: %w", err)
			}
			fmt.Printf("✅ task/%s submitted\n", task.Key)
			return nil
		},
	}
	cmd.Flags().StringVar(&image, "image", "ghcr.io/turtacn/agentic/agent:latest", "runtime image")
	cmd.Flags().StringVar(&cpu, "cpu", "100m", "cpu resource")
	cmd.Flags().StringVar(&mem, "memory", "128Mi", "memory resource")
	cmd.Flags().DurationVar(&timeout, "timeout", 5*time.Minute, "max task duration")
	cmd.Flags().StringVar(&priority, "priority", "normal", "task priority")
	return cmd
}

/* -------------------- status -------------------- */
func taskStatusCmd(kubeCfg string) *cobra.Command {
	return &cobra.Command{
		Use:   "status TASK-ID",
		Short: "Get task status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]
			ns := cmdFlag(cmd, "namespace")

			cl, err := utils.ClientFromKubeConfig(kubeCfg)
			if err != nil {
				return err
			}
			cm, err := cl.CoreV1().ConfigMaps(ns).Get(cmd.Context(), taskKeyToCM(taskID), metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("task not found: %w", err)
			}
			task := fromConfigMap(cm)
			fmt.Printf("Task:   %s\n", task.Key)
			fmt.Printf("Status: %s\n", cm.Labels["status"])
			fmt.Printf("Age:    %v\n", time.Since(cm.CreationTimestamp.Time).Round(time.Second)))
			if msg := cm.Data["output"]; msg != "" {
				fmt.Printf("Output: %s\n", msg)
			}
			return nil
		},
	}
}

/* -------------------- cancel -------------------- */
func taskCancelCmd(kubeCfg string) *cobra.Command {
	return &cobra.Command{
		Use:   "cancel TASK-ID",
		Short: "Cancel running task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			taskID := args[0]
			ns := cmdFlag(cmd, "namespace")

			cl, err := utils.ClientFromKubeConfig(kubeCfg)
			if err != nil {
				return err
			}
			cm, err := cl.CoreV1().ConfigMaps(ns).Get(cmd.Context(), taskKeyToCM(taskID), metav1.GetOptions{})
			if err != nil {
				return err
			}
			cm.Labels["status"] = "cancelled"
			_, err = cl.CoreV1().ConfigMaps(ns).Update(cmd.Context(), cm, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("cancel task: %w", err)
			}
			fmt.Printf("🛑 task/%s cancelled\n", taskID)
			return nil
		},
	}
}

/* -------------------- list -------------------- */
func taskListCmd(kubeCfg string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			ns := cmdFlag(cmd, "namespace")
			cl, err := utils.ClientFromKubeConfig(kubeCfg)
			if err != nil {
				return err
			}
			cms, err := cl.CoreV1().ConfigMaps(ns).List(cmd.Context(), metav1.ListOptions{
				LabelSelector: "type=task",
			})
			if err != nil {
				return err
			}
			fmt.Printf("%-20s %-10s %-8s %-64s\n", "TASK-ID", "STATUS", "AGE", "COMMAND")
			for _, cm := range cms.Items {
				cmd := "N/A"
				if v := cm.Data["command"]; v != "" {
					cmd = v
				}
				fmt.Printf("%-20s %-10s %-8s %-64s\n",
					cm.Name, cm.Labels["status"], time.Since(cm.CreationTimestamp.Time).Round(time.Second), cmd)
			}
			return nil
		},
	}
}

// ---------- 帮助函数 ----------
func toConfigMap(task *types.Task, ns string) *corev1.ConfigMap {
	data, _ := json.MarshalIndent(task, "", "  ")
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      taskKeyToCM(task.Key),
			Namespace: ns,
			Labels: map[string]string{
				"type":   "task",
				"status": "pending",
			},
		},
		Data: map[string]string{
			"task":    string(data),
			"command": fmt.Sprintf("%v", task.Spec.Command),
		},
	}
}

func fromConfigMap(cm *corev1.ConfigMap) *types.Task {
	var t types.Task
	_ = json.Unmarshal([]byte(cm.Data["task"]), &t)
	return &t
}

func taskKeyToCM(k string) string { return "task-" + k }

func cmdFlag(cmd *cobra.Command, key string) string {
	v, _ := cmd.Flags().GetString(key)
	return v
}

func ClientFromKubeConfig(path string) (*kubernetes.Clientset, error) {
	return utils.ClientFromKubeConfig(path)
}
//Personal.AI order the ending
