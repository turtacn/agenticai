# Testing Strategy

A robust, multi-layered testing strategy is essential for ensuring the quality, stability, and maintainability of the AgenticAI project. This document outlines the different types of tests used in the project and provides guidance on how to write them.

## 1. Unit Tests

Unit tests are the foundation of our testing strategy. They are fast, isolated, and verify the correctness of individual functions and components.

**Tooling:**
*   Go's standard `testing` package.
*   `github.com/stretchr/testify/assert` for assertions.
*   `sigs.k8s.io/controller-runtime/pkg/client/fake` for creating a mock Kubernetes client.

**Location:** Test files should be located in the same package as the code they are testing, with a `_test.go` suffix (e.g., `task_controller_test.go`).

**Example: Testing Controller Logic**

When testing controller logic, we use the fake client to simulate the Kubernetes API and verify that the reconciler makes the correct changes.

```go
// pkg/controller/task_controller_test.go

func TestCheckDependencies(t *testing.T) {
	// 1. Setup the scheme with all necessary types
	s := runtime.NewScheme()
	scheme.AddToScheme(s) // Add core Kubernetes types
	agenticaiov1.AddToScheme(s) // Add our custom types

	// 2. Create mock objects that will exist in the fake client
	depTask := &agenticaiov1.Task{
		ObjectMeta: metav1.ObjectMeta{Name: "dep1", Namespace: "default"},
		Status:     agenticaiov1.TaskStatus{Phase: agenticaiov1.TaskCompleted},
	}
	task := &agenticaiov1.Task{
		ObjectMeta: metav1.ObjectMeta{Name: "task2", Namespace: "default"},
		Spec: agenticaiov1.TaskSpec{
			Dependencies: []agenticaiov1.Dependency{
				{TaskID: "dep1", State: agenticaiov1.TaskCompleted},
			},
		},
	}

	// 3. Create the fake client and the reconciler
	reconciler := &TaskReconciler{
		Client: fake.NewClientBuilder().WithScheme(s).WithObjects(depTask, task).Build(),
		Scheme: s,
	}

	// 4. Call the function to be tested and assert the result
	ready, err := reconciler.checkDependencies(context.Background(), task)
	assert.NoError(t, err)
	assert.True(t, ready)
}
```

## 2. Integration Tests

Integration tests verify the interaction between different components of the system. For our controllers, this means testing against a real `kube-apiserver` and `etcd` instance to ensure that our reconciliation logic works correctly in a more realistic environment.

**Tooling:**
*   `sigs.k8s.io/controller-runtime/pkg/envtest` for setting up a temporary test environment.

**Location:** Integration tests should be placed in a separate `_test` package (e.g., `controller/suite_test.go`) to avoid circular dependencies.

**Example Skeleton:**

```go
// controller/suite_test.go

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	// ... register schemes ...

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

var _ = Describe("Task Controller", func() {
	It("Should create a Job when a Task is created", func() {
		// 1. Create a Task resource
		// 2. Verify that the controller creates a corresponding Job
		// 3. Update the Job status and verify the Task status is updated
	})
})
```

## 3. End-to-End (E2E) Tests

E2E tests simulate real user workflows from start to finish. They are the most comprehensive tests and provide the highest level of confidence that the system is working as expected.

**Tooling:**
*   `kind` (Kubernetes in Docker) for creating ephemeral Kubernetes clusters in CI.
*   A testing framework like Ginkgo or just the standard `testing` package to orchestrate the test steps.

**Workflow Example:**

1.  **Setup**: In a CI pipeline, use `kind` to create a new Kubernetes cluster.
2.  **Deploy**: Deploy the AgenticAI controller and all its required CRDs to the `kind` cluster.
3.  **Execute**: Use the compiled `actl` binary to run commands against the cluster.
    *   `actl task submit ...`
    *   `actl task status ...`
4.  **Verify**: Use `kubectl` to inspect the state of the cluster and verify that the resources (Tasks, Jobs, Pods) are in the expected state.
5.  **Teardown**: Delete the `kind` cluster.
