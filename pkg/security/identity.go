// pkg/security/identity.go
package security

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/credentials"

	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/turtacn/agenticai/internal/logger"
	"go.uber.org/zap"
)

type Identity interface {
	ID() spiffeid.ID
	Source() credentials.TransportCredentials
}

type identityImpl struct {
	source   *workloadapi.X509Source
	spiffeID spiffeid.ID
}

// GetIdentity 获取工作负载身份；若无则阻塞重试直至成功
func GetIdentity(ctx context.Context) (Identity, error) {
	ctx, span := trace.SpanFromContext(ctx).TracerProvider().Tracer("security").Start(ctx, "GetIdentity")
	defer span.End()

	w, err := workloadapi.New(ctx)
	if err != nil {
		return nil, err
	}
	src, err := workloadapi.NewX509Source(ctx, workloadapi.WithClient(w))
	if err != nil {
		return nil, err
	}
	svid, err := src.GetX509SVID()
	if err != nil {
		return nil, err
	}
	logger.Info(ctx, "loaded SPIFFE identity", zap.Stringer("svid", svid.ID))
	return &identityImpl{source: src, spiffeID: svid.ID}, nil
}

func (i *identityImpl) ID() spiffeid.ID { return i.spiffeID }

func (i *identityImpl) Source() credentials.TransportCredentials {
	tlsConfig := tlsconfig.MTLSServerConfig(i.source, i.source, tlsconfig.AuthorizeAny())
	return credentials.NewTLS(tlsConfig)
}

// --------------------------------------------------------------------

// SPIFFEInterceptor gRPC 一元拦截器，附加 caller id
func SPIFFEInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		id, err := authorize(ctx)
		if err != nil {
			return nil, err
		}
		newCtx := context.WithValue(ctx, ctxKeyCaller, id.String())
		return handler(newCtx, req)
	}
}

func SPIFFEStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		id, err := authorize(ss.Context())
		if err != nil {
			return err
		}
		newCtx := context.WithValue(ss.Context(), ctxKeyCaller, id.String())
		return handler(srv, &wrappedServerStream{ss, newCtx})
	}
}

// authorize is a helper to extract and validate a peer's SVID.
func authorize(ctx context.Context) (spiffeid.ID, error) {
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return spiffeid.ID{}, status.Error(codes.Unauthenticated, "no peer info")
	}
	tlsInfo, ok := peer.AuthInfo.(credentials.TLSInfo)
	if !ok || len(tlsInfo.State.PeerCertificates) == 0 {
		return spiffeid.ID{}, status.Error(codes.Unauthenticated, "invalid tls auth")
	}
	id, err := x509svid.IDFromCert(tlsInfo.State.PeerCertificates[0])
	if err != nil {
		return spiffeid.ID{}, status.Error(codes.Unauthenticated, err.Error())
	}
	return id, nil
}

// wrappedServerStream is a thin wrapper around grpc.ServerStream that allows modifying the context.
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
//Personal.AI order the ending
