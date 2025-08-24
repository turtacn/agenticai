// pkg/types/tool.go
package types

import (
	"encoding/json"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ToolSpec struct {
	// 基本元数据（面向平台）
	Name        string            `json:"name"`        // 全局唯一
	Version     string            `json:"version"`
	DisplayName string            `json:"displayName,omitempty"`
	Description string            `json:"description,omitempty"`
	Author      string            `json:"author,omitempty"`
	Categories  []string          `json:"categories,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`

	// 运行方式（三者 1:N）
	MCP       *MCPBinding       `json:"mcp,omitempty"`
	OpenAPI   *OpenAPIBinding   `json:"openapi,omitempty"`
	CustomExec *CustomBinding   `json:"custom,omitempty"`

	// 权限&安全
	RequiredPermissions []Permission `json:"requiredPermissions,omitempty"`
	NetworkPolicy       *NetworkPolicy `json:"networkPolicy,omitempty"`

	// 质量保障
	CORS      *CORSPolicy `json:"cors,omitempty"`
	RateLimit *RateLimit  `json:"rateLimit,omitempty"`

	// 可观测
	DefaultTimeout metav1.Duration `json:"defaultTimeout,omitempty"`
}

// Tool 顶层 CRD 对象
type Tool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ToolSpec `json:"spec"`
}

// MCPBinding 描述连接到 MCP 服务器
type MCPBinding struct {
	ServerURL   string            `json:"serverUrl"`
	Headers     map[string]string `json:"headers,omitempty"`
	RequestBody *json.RawMessage  `json:"requestBody,omitempty"` // 支持模板化
}

// OpenAPIBinding 描述 REST/OpenAPI 端点
type OpenAPIBinding struct {
	SpecURL     string            `json:"specUrl"`     // OpenAPI/Swagger
	BaseURL     string            `json:"baseUrl,omitempty"`
	AuthMethods map[string]string `json:"authMethods,omitempty"` // securitySchemeKey->Value
}

// CustomBinding 容器化脚本/可执行
type CustomBinding struct {
	Image string   `json:"image"`
	Cmd   []string `json:"cmd,omitempty"`
	Args  []string `json:"args,omitempty"`
}

// Permission 细粒度权限，用于 RBAC
type Permission struct {
	Resource string `json:"resource"`
	Actions  []string `json:"actions"`
}

// NetworkPolicy 限制沙箱内网访问
type NetworkPolicy struct {
	AllowOutbound []string `json:"allowOutbound,omitempty"`
}

// CORSPolicy 供 HTTP API/Gateway
type CORSPolicy struct {
	Origins        []string `json:"origins,omitempty"`
	AllowedHeaders []string `json:"allowedHeaders,omitempty"`
}

// RateLimit 网关级限制
type RateLimit struct {
	Max      int32    `json:"max"`
	Window   *time.Duration `json:"window,omitempty"`
}
//Personal.AI order the ending
