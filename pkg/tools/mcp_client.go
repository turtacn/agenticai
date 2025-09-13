// pkg/tools/mcp_client.go
package tools

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/turtacn/agenticai/internal/logger"
	"github.com/turtacn/agenticai/pkg/apis"
	// pb "githu" // 省略自动生成pb前缀包
)

// MCPClient 协议客户端
type MCPClient interface {
	Connect(ctx context.Context, addr string) error
	Close() error
	ListTools(ctx context.Context) ([]*apis.Metadata, error)
	CallTool(ctx context.Context, name string, args map[string]interface{}) (*apis.ToolResult, error)
	ServerInfo(ctx context.Context) (*apis.MCPServerInfo, error)
}

type mcpClient struct {
	conn *grpc.ClientConn
	// client   pb.MCPServerClient
	trace trace.Tracer
}

func NewMCPClient() MCPClient {
	return &mcpClient{
		trace: otel.Tracer("mcp"),
	}
}

func (c *mcpClient) Connect(ctx context.Context, addr string) error {
	ctx, span := c.trace.Start(ctx, "MCPClient.Connect")
	defer span.End()

	dial, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second))
	if err != nil {
		return fmt.Errorf("mcp connect fail %w", err)
	}
	c.conn = dial
	// c.client = pb.NewMCPServerClient(dial)
	logger.Info(ctx, "mcp connected", zap.String("addr", addr))
	return nil
}

func (c *mcpClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *mcpClient) ListTools(ctx context.Context) ([]*apis.Metadata, error) {
	// ctx, span := c.trace.Start(ctx, "MCPClient.ListTools")
	// defer span.End()
	// resp, err := c.client.ListTools(ctx, &pb.ListToolsRequest{})
	// if err != nil {
	// 	return nil, err
	// }
	// out := make([]*types.Metadata, 0, len(resp.Tools))
	// for _, t := range resp.Tools {
	// 	out = append(out, &types.Metadata{
	// 		ID:      t.Id,
	// 		Name:    t.Name,
	// 		Version: t.Version,
	// 		Digest:  t.Digest,
	// 	})
	// }
	// return out, nil
	return nil, errors.New("not implemented")
}

func (c *mcpClient) CallTool(ctx context.Context, name string, args map[string]interface{}) (*apis.ToolResult, error) {
	// ctx, span := c.trace.Start(ctx, "MCPClient.CallTool")
	// defer span.End()
	// resp, err := c.client.CallTool(ctx, &pb.CallToolRequest{
	// 	ToolId:    name,
	// 	Arguments: args,
	// })
	// if err != nil {
	// 	return nil, err
	// }
	// return &types.ToolResult{
	// 	Output: resp.Output,
	// 	Error:  resp.Error,
	// 	Status: resp.Status,
	// }, nil
	return nil, errors.New("not implemented")
}

func (c *mcpClient) ServerInfo(ctx context.Context) (*apis.MCPServerInfo, error) {
	// resp, err := c.client.GetServerInfo(ctx, &pb.GetServerInfoRequest{})
	// if err != nil {
	// 	return nil, err
	// }
	// return &types.MCPServerInfo{
	// 	Name:    resp.Name,
	// 	Version: resp.Version,
	// }, nil
	return nil, errors.New("not implemented")
}
//Personal.AI order the ending
