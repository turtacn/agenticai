// internal/errors/errors.go
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Kind - 错误分类代码（可用于日志、告警、国际化）
type Kind string

const (
	KindInternal    Kind = "internal"     // 系统级
	KindValidation  Kind = "validation"   // 输入校验失败
	KindNotFound    Kind = "not_found"    // 资源不存在
	KindConflict    Kind = "conflict"     // 资源冲突
	KindPermission  Kind = "permission"   // 未授权
	KindTimeout     Kind = "timeout"      // 超时
	KindUnavailable Kind = "unavailable"  // 服务不可用
	KindCancelled   Kind = "cancelled"    // 任务取消
)

// Error - 自定义错误类型
type Error struct {
	Kind  Kind
	Op    string // 正在执行的操作
	Msg   string // 描述文本
	Err   error  // 原始错误
	Stack []byte // 调用栈快照
}

//
// constructors
//

func E(args ...interface{}) *Error {
	e := &Error{Kind: KindInternal}
	for _, arg := range args {
		switch v := arg.(type) {
		case Kind:
			e.Kind = v
		case error:
			e.Err = v
		case string:
			e.Msg = v
		default:
			panic(fmt.Errorf("errors.E: bad arg %T", arg))
		}
	}
	e.Stack = stack()
	return e
}

//
// helpers
//

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err == nil {
		return fmt.Sprintf("[%s] %s", e.Kind, e.Msg)
	}
	return fmt.Sprintf("[%s] %s: %v", e.Kind, e.Msg, e.Err)
}

func (e *Error) Unwrap() error   { return e.Err }
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Kind == t.Kind
}
func (e *Error) As(target interface{}) bool {
	return errors.As(e.Err, target)
}

//
// HTTP mapping
//

func (e *Error) HTTPStatus() int {
	switch e.Kind {
	case KindNotFound:
		return http.StatusNotFound
	case KindConflict:
		return http.StatusConflict
	case KindPermission:
		return http.StatusForbidden
	case KindTimeout:
		return http.StatusRequestTimeout
	case KindUnavailable:
		return http.StatusServiceUnavailable
	case KindValidation:
		return http.StatusBadRequest
	case KindCancelled:
		return http.StatusRequestTimeout
	default:
		return http.StatusInternalServerError
	}
}

//
// convenience constructors
//

func Internal(err error, args ...interface{}) *Error { return mk(KindInternal, err, args...) }
func Validation(err error, args ...interface{}) *Error { return mk(KindValidation, err, args...) }
func NotFound(err error, args ...interface{}) *Error { return mk(KindNotFound, err, args...) }
func Conflict(err error, args ...interface{}) *Error { return mk(KindConflict, err, args...) }
func Permission(err error, args ...interface{}) *Error { return mk(KindPermission, err, args...) }
func Timeout(err error, args ...interface{}) *Error { return mk(KindTimeout, err, args...) }
func Unavailable(err error, args ...interface{}) *Error { return mk(KindUnavailable, err, args...) }

func mk(k Kind, err error, a ...interface{}) *Error {
	msg := ""
	if len(a) > 0 {
		msg = a[0].(string)
	}
	return E(k, msg, err)
}

// stack - mini 调用栈快照（仅错误级别保留，生产采样）
func stack() []byte {
	return nil // placeholder for lightweight snapshots
}
// stack placeholder stub for brevity → real impl can embed runtime.Stack
//Personal.AI order the ending
