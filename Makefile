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

# Generate protobuf code
generate-proto:
	go get google.golang.org/protobuf/cmd/protoc-gen-go@v1.33.0
	go get google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0
	./protoc/bin/protoc --proto_path=./api/proto/mcp --proto_path=./protoc/include \
		--plugin=protoc-gen-go=$(go list -f '{{.Target}}' google.golang.org/protobuf/cmd/protoc-gen-go) \
		--plugin=protoc-gen-go-grpc=$(go list -f '{{.Target}}' google.golang.org/grpc/cmd/protoc-gen-go-grpc) \
		--go_out=./pkg/gen --go_opt=paths=source_relative \
		--go-grpc_out=./pkg/gen --go-grpc_opt=paths=source_relative \
		model_context.proto

# Build all binaries
build:
	go build -o ./bin/actl ./cmd/actl
	go build -o ./bin/controller ./cmd/controller
	go build -o ./bin/agent-runtime ./cmd/agent-runtime
	go build -o ./bin/tool-gateway ./cmd/tool-gateway

# Run all tests
test:
	go test ./...
