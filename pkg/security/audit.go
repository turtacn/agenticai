// pkg/security/audit.go
package security

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"

	"github.com/turtacn/agenticai/internal/logger"
	api "github.com/turtacn/agenticai/pkg/types"
)

var (
	auditCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "agenticai_audit_events_total",
		Help: "audit events",
	}, []string{"result", "resource"})
)

// Audit 单例
var (
	fp       *os.File
	encoder *json.Encoder
)

func init() {
	var err error
	fp, err = os.OpenFile("/var/log/audit.jsonl", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		panic(err)
	}
	encoder = json.NewEncoder(fp)
}

func AuditLog(ctx context.Context, ev *api.AuditEvent) {
	ev.Timestamp = time.Now().UTC()
	if err := encoder.Encode(ev); err != nil {
		logger.Error(ctx, "audit write fail", zap.Error(err))
	}
	result := "success"
	if ev.Error != "" {
		result = "fail"
	}
	auditCounter.WithLabelValues(result, ev.Resource).Inc()
}
//Personal.AI order the ending
