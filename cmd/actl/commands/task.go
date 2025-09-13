// cmd/actl/commands/task.go
package commands

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	agenticaiov1 "github.com/turtacn/agenticai/pkg/apis/agenticai.io/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
	var cpu, mem, image string
	var priority int32
	var timeout time.Duration

	cmd := &cobra.Command{
		Use:   "submit [COMMAND...]",
		Short: "Submit a new AI task",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ns := cmdFlag(cmd, "namespace")

			res := corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(cpu),
					corev1.ResourceMemory: resource.MustParse(mem),
				},
			}

			spec := agenticaiov1.TaskSpec{
				ImageRef:  image,
				Command:   args,
				Resources: res,
				Timeout:   metav1.Duration{Duration: timeout},
				Priority:  priority,
			}
			task := &agenticaiov1.Task{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "task-" + time.Now().UTC().Format("20060102-150405"),
					Namespace: ns,
				},
				Spec: spec,
			}

			// TODO: This command should create a Task CRD, not a ConfigMap.
			// This requires a generated clientset, which is failing in the sandbox.
			// Commenting out for now to allow compilation.
			/*
				cm := toConfigMap(task, ns)

				cl, err := ClientFromKubeConfig(kubeCfg)
				if err != nil {
					return err
				}
				if _, err := cl.CoreV1().ConfigMaps(ns).Create(ctx, cm, metav1.CreateOptions{}); err != nil {
					return fmt.Errorf("create task configmap: %w", err)
				}
			*/
			fmt.Printf("✅ task/%s would be submitted (creation logic commented out)\n", task.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&image, "image", "ghcr.io/turtacn/agentic/agent:latest", "runtime image")
	cmd.Flags().StringVar(&cpu, "cpu", "100m", "cpu resource request")
	cmd.Flags().StringVar(&mem, "memory", "128Mi", "memory resource request")
	cmd.Flags().DurationVar(&timeout, "timeout", 5*time.Minute, "max task duration")
	cmd.Flags().Int32Var(&priority, "priority", 0, "task priority (higher value is higher priority)")
	return cmd
}

/* -------------------- status -------------------- */
func taskStatusCmd(kubeCfg string) *cobra.Command {
	return &cobra.Command{
		Use:   "status TASK-ID",
		Short: "Get task status (currently disabled)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("This command is temporarily disabled and needs to be refactored to use Task CRDs.")
			return nil
		},
	}
}

/* -------------------- cancel -------------------- */
func taskCancelCmd(kubeCfg string) *cobra.Command {
	return &cobra.Command{
		Use:   "cancel TASK-ID",
		Short: "Cancel running task (currently disabled)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("This command is temporarily disabled and needs to be refactored to use Task CRDs.")
			return nil
		},
	}
}

/* -------------------- list -------------------- */
func taskListCmd(kubeCfg string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List tasks (currently disabled)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("This command is temporarily disabled and needs to be refactored to use Task CRDs.")
			return nil
		},
	}
}

// TODO: Refactor these helpers to work with Task CRDs
/*
// ---------- 帮助函数 ----------
func toConfigMap(task *types.Task, ns string) *corev1.ConfigMap {
	data, _ := json.MarshalIndent(task.Spec, "", "  ")
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      taskKeyToCM(task.Name),
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
*/
func cmdFlag(cmd *cobra.Command, key string) string {
	v, _ := cmd.Flags().GetString(key)
	return v
}

//Personal.AI order the ending
