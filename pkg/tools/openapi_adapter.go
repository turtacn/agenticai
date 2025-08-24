// pkg/tools/openapi_adapter.go
package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"

	"github.com/turtacn/agenticai/internal/logger"
	types "github.com/turtacn/agenticai/pkg/types"
)

// OpenAPIAdapter 适配器
type OpenAPIAdapter interface {
	LoadSpec(ctx context.Context, raw []byte) error
	ListTools() []*types.Metadata
	Invoke(
		ctx context.Context,
		toolName string,
		input map[string]interface{},
	) (*types.ToolResult, error)
}

type openAPIAdapter struct {
	doc   *openapi3.T
	cache map[string]*types.ToolSpec // name->spec
	http  *http.Client
	trace trace.Tracer
}

func NewOpenAPIAdapter() OpenAPIAdapter {
	return &openAPIAdapter{
		cache: make(map[string]*types.ToolSpec),
		http:  &http.Client{Timeout: 10 * time.Second, Transport: otelhttp.NewTransport(http.DefaultTransport)},
		trace: trace.Tracer{},
	}
}

func (a *openAPIAdapter) LoadSpec(_ context.Context, raw []byte) error {
	doc, err := openapi3.NewLoader().LoadFromData(raw)
	if err != nil {
		return fmt.Errorf("invalid openapi: %w", err)
	}
	// 扁平化路径生成工具
	for path, pItem := range doc.Paths {
		for method, op := range pItem.Operations() {
			id := fmt.Sprintf("%s %s", method, path)
			a.cache[id] = &types.ToolSpec{
				ID:       id,
				Name:     op.OperationID,
				Version:  fmt.Sprintf("%d", 1),
				Metadata: op.Description,
				Category: "openapi",
				ArgsSchema: convertOpenAPIArgs(op.Parameters),
			}
		}
	}
	return nil
}

func (a *openAPIAdapter) ListTools() []*types.Metadata {
	out := make([]*types.Metadata, 0, len(a.cache))
	for _, v := range a.cache {
		out = append(out, &types.Metadata{
			ID:      v.ID,
			Name:    v.Name,
			Version: v.Version,
			Digest:  fmt.Sprintf("%x", hashSpec(v)),
		})
	}
	return out
}

func (a *openAPIAdapter) Invoke(
	ctx context.Context,
	toolName string,
	input map[string]interface{},
) (*types.ToolResult, error) {
	ctx, span := a.trace.Start(ctx, "OpenAPIAdapter.Invoke")
	defer span.End()
	spec, ok := a.cache[toolName]
	if !ok {
		return nil, fmt.Errorf("tool %s not found", toolName)
	}
	// 简易：用 GET 方式访问 path
	path := toolName
	req, _ := http.NewRequestWithContext(ctx, "GET", path, nil)
	q := req.URL.Query()
	for k, v := range input {
		q.Set(k, fmt.Sprintf("%v", v))
	}
	req.URL.RawQuery = q.Encode()
	resp, err := a.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	return &types.ToolResult{
		Output: string(body),
		Status: int32(resp.StatusCode),
	}, nil
}

/* -------------- helper --------------- */
func convertOpenAPIArgs(params openapi3.Parameters) types.AnyMap { return nil }
func hashSpec(s *types.ToolSpec) uint64                    { return 0 }
//Personal.AI order the ending
