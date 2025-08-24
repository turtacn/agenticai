// pkg/security/rbac.go
package security

import (
	"context"
	"fmt"
	"regexp"
	"sync"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/turtacn/agenticai/internal/errors"
	"github.com/turtacn/agenticai/internal/logger"
	api "github.com/turtacn/agenticai/pkg/types"
)

const ctxKeyCaller = callerKey("caller")

type callerKey string

// RBAC 管理器
type RBAC interface {
	UpdatePolicy(pol *api.SecurityPolicy)
	Authorize(ctx context.Context, subject, action, resource string) error
}

type rule struct {
	sub, act, res *regexp.Regexp
}

type rbac struct {
	mu    sync.RWMutex
	rules []rule
}

func NewRBAC() RBAC { return &rbac{} }

func (r *rbac) UpdatePolicy(pol *api.SecurityPolicy) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rules := make([]rule, 0, len(pol.Rules))
	for _, ro := range pol.Rules {
		rules = append(rules, rule{
			sub: regexp.MustCompile(wild2regex(ro.Subject)),
			act: regexp.MustCompile(wild2regex(ro.Action)),
			res: regexp.MustCompile(wild2regex(ro.Resource)),
		})
	}
	r.rules = rules
}

func (r *rbac) Authorize(ctx context.Context, subject, action, resource string) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, ru := range r.rules {
		if ru.sub.MatchString(subject) && ru.act.MatchString(action) && ru.res.MatchString(resource) {
			return nil
		}
	}
	return status.Error(codes.PermissionDenied, "unauthorized")
}

func wild2regex(s string) string {
	return "^" + s + "$"
}
//Personal.AI order the ending
