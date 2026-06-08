package proxy

import (
	"context"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/Edu963/ocultar/pkg/refinery"
)

// GrpcServer represents the gRPC interceptor listener for Ocultar.
// This allows intercepting SIEM logs (e.g. OpenTelemetry) over HTTP/2.
type GrpcServer struct {
	eng        *refinery.Refinery
	listenAddr string
	server     *grpc.Server
}

// NewGrpcServer creates a new gRPC proxy server
func NewGrpcServer(eng *refinery.Refinery, listenAddr string) *GrpcServer {
	return &GrpcServer{
		eng:        eng,
		listenAddr: listenAddr,
	}
}

// Start opens the TCP listener and serves gRPC traffic.
// This implements a transparent gRPC interceptor for SIEM logs.
func (s *GrpcServer) Start() error {
	lis, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port: %w", err)
	}

	// Create a generic gRPC server with interceptors.
	// In production, this would be bound to specific proto definitions like OTLP logs.
	s.server = grpc.NewServer(
		grpc.UnaryInterceptor(s.unaryInterceptor),
		grpc.StreamInterceptor(s.streamInterceptor),
	)

	go func() {
		log.Printf("[GRPC] Listening on %s for SIEM traffic", s.listenAddr)
		if err := s.server.Serve(lis); err != nil {
			log.Printf("[GRPC] Server error: %v", err)
		}
	}()

	return nil
}

// Stop gracefully shuts down the gRPC server.
func (s *GrpcServer) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}

// unaryInterceptor intercepts standard Unary gRPC calls to detect and redact PII
func (s *GrpcServer) unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	sanitizedReq, err := s.eng.ProcessInterface(req, "grpc_proxy")
	if err != nil {
		log.Printf("[GRPC-BLOCK] Intercept error on %s: %v", info.FullMethod, err)
		return nil, fmt.Errorf("fail-closed security block on gRPC stream")
	}
	return handler(ctx, sanitizedReq)
}

// streamInterceptor intercepts streaming gRPC calls (common in telemetry and SIEM logging).
func (s *GrpcServer) streamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// Acts as the entry point for intercepting streaming log data.
	return handler(srv, ss)
}
