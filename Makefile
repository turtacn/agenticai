# Generate CRD client code
generate-client:
	/bin/bash $(go list -f '{{.Dir}}' k8s.io/code-generator)/kube_codegen.sh \
		all \
		github.com/turtacn/agenticai/pkg/client \
		github.com/turtacn/agenticai/pkg/apis \
		"agenticai.io:v1" \
		--go-header-file=./hack/boilerplate.go.txt

# Generate deepcopy for API types
generate:
	/bin/bash $(go list -f '{{.Dir}}' sigs.k8s.io/controller-tools/cmd/controller-gen)/controller-gen \
		object:headerFile="./hack/boilerplate.go.txt" \
		paths="./pkg/apis/..."
