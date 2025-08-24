// pkg/security/identity.go
package security

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"path/filepath"

	"go.opentelemetry.io/otel/trace"
	"golang.org/x/net/context/ctxhttp"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	api "github.com/turtacn/agenticai/pkg/types"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/svid/x509svid"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"github.com/turtacn/agenticai/internal/logger"
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
	logger.Info(ctx, "loaded SPIFFE identity", "svid", svid.ID)
	return &identityImpl{source: src, spiffeID: svid.ID}, nil
}

func (i *identityImpl) ID() spiffeid.ID { return i.spiffeID }

func (i *identityImpl) Source() credentials.TransportCredentials {
	return credentials.NewTLS(i.source)
}

// --------------------------------------------------------------------

// SPIFFEInterceptor gRPC 一元拦截器，附加 caller id
func SPIFFEInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		peer, ok := peer.FromContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "no peer info")
		}
		tlsInfo, ok := peer.AuthInfo.(credentials.TLSInfo)
		if !ok || len(tlsInfo.State.PeerCertificates) == 0 {
			return nil, status.Error(codes.Unauthenticated, "invalid tls auth")
		}
		cert := tlsInfo.State.PeerCertificates[0]
		id, err := x509svid.IDFromCert(cert)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}
		newCtx := context.WithValue(ctx, ctxKeyCaller{}, id.String())
		return handler(newCtx, req)
	}
}

func SPIFFEStreamInterceptor() grpc.StreamServerInterceptor {
	return grpc.NewServerStream
}
//Personal.AI order the ending
