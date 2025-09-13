// pkg/security/rbac.go
package security

import (
	"context"
	"regexp"
	"sync"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/turtacn/agenticai/pkg/apis"
)

const ctxKeyCaller = callerKey("caller")

type callerKey string

// RBAC 管理器
type RBAC interface {
	UpdatePolicy(pol *apis.SecurityPolicy)
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

func (r *rbac) UpdatePolicy(pol *apis.SecurityPolicy) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rules := make([]rule, 0, len(pol.Spec.Rules))
	for _, ro := range pol.Spec.Rules {
		// TODO: This is a temporary simplification to allow compilation.
		// The RBAC authorizer needs to be updated to properly handle multiple verbs and resources per rule.
		if len(ro.Verbs) > 0 && len(ro.Resources) > 0 {
			rules = append(rules, rule{
				sub: regexp.MustCompile(wild2regex(ro.Role)),
				act: regexp.MustCompile(wild2regex(ro.Verbs[0])),
				res: regexp.MustCompile(wild2regex(ro.Resources[0])),
			})
		}
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
