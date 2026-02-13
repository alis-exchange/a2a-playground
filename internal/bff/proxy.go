package bff

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"

	"connectrpc.com/connect"
	"github.com/a2aproject/a2a-go/a2apb"
	"github.com/alis-exchange/a2a-playground/gen/go/a2apbconnect"
	"go.alis.build/client/v2"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/grpc/metadata"
)

// A2AServiceHandler is the interface implemented by both grpcProxy and jsonrpcProxy.
type A2AServiceHandler interface {
	a2apbconnect.A2AServiceHandler
	Handler() (string, http.Handler)
}

// grpcProxy implements A2AServiceHandler by forwarding requests to a gRPC agent.
type grpcProxy struct {
	agentURL string
	mu       sync.Mutex
	client   a2apb.A2AServiceClient
}

// NewGrpcProxy creates a proxy that forwards to the gRPC agent at agentURL.
// agentURL may be "localhost:8080" or "http://localhost:8080"; the scheme is stripped for gRPC.
// Connection is established lazily on first request so startup does not fail if the agent is unreachable.
func NewGrpcProxy(agentURL string) *grpcProxy {
	return &grpcProxy{agentURL: agentURL}
}

// getClient returns or creates the gRPC client, connecting lazily on first use.
func (p *grpcProxy) getClient(ctx context.Context) (a2apb.A2AServiceClient, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.client != nil {
		return p.client, nil
	}
	addr := p.agentURL
	isLocalhost := strings.Contains(addr, "localhost:")

	conn, err := client.NewConn(ctx, addr, isLocalhost, client.WithoutAuth())
	if err != nil {
		return nil, err
	}
	p.client = a2apb.NewA2AServiceClient(conn)
	return p.client, nil
}

// NewGrpcProxyFromConn creates a proxy from an existing gRPC connection.
func NewGrpcProxyFromConn(client a2apb.A2AServiceClient) *grpcProxy {
	return &grpcProxy{client: client}
}

// ConnectOptions returns Connect-RPC handler options for the proxy.
func (p *grpcProxy) ConnectOptions() []connect.HandlerOption {
	return nil
}

// Handler returns the path and HTTP handler for mounting the A2A Connect service.
func (p *grpcProxy) Handler() (string, http.Handler) {
	return a2apbconnect.NewA2AServiceHandler(p, p.ConnectOptions()...)
}

// Ensure grpcProxy implements a2apbconnect.A2AServiceHandler.
var _ a2apbconnect.A2AServiceHandler = (*grpcProxy)(nil)

// withAgentHeaders appends agent headers from context to outgoing gRPC metadata.
func withAgentHeaders(ctx context.Context) context.Context {
	headers := AgentHeadersFromContext(ctx)
	if len(headers) == 0 {
		return ctx
	}
	md := metadata.New(nil)
	for k, v := range headers {
		md.Append(k, v)
	}
	return metadata.NewOutgoingContext(ctx, md)
}

// SendMessage forwards the request to the gRPC agent.
func (p *grpcProxy) SendMessage(ctx context.Context, req *connect.Request[a2apb.SendMessageRequest]) (*connect.Response[a2apb.SendMessageResponse], error) {
	ctx = withAgentHeaders(ctx)
	client, err := p.getClient(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	resp, err := client.SendMessage(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// SendStreamingMessage forwards the streaming request to the gRPC agent.
func (p *grpcProxy) SendStreamingMessage(ctx context.Context, req *connect.Request[a2apb.SendMessageRequest], stream *connect.ServerStream[a2apb.StreamResponse]) error {
	ctx = withAgentHeaders(ctx)
	client, err := p.getClient(ctx)
	if err != nil {
		return connect.NewError(connect.CodeUnavailable, err)
	}
	grpcStream, err := client.SendStreamingMessage(ctx, req.Msg)
	if err != nil {
		return err
	}
	for {
		msg, err := grpcStream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := stream.Send(msg); err != nil {
			return err
		}
	}
}

// GetTask forwards the request to the gRPC agent.
func (p *grpcProxy) GetTask(ctx context.Context, req *connect.Request[a2apb.GetTaskRequest]) (*connect.Response[a2apb.Task], error) {
	ctx = withAgentHeaders(ctx)
	client, err := p.getClient(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	resp, err := client.GetTask(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// ListTasks forwards the request to the gRPC agent.
func (p *grpcProxy) ListTasks(ctx context.Context, req *connect.Request[a2apb.ListTasksRequest]) (*connect.Response[a2apb.ListTasksResponse], error) {
	ctx = withAgentHeaders(ctx)
	client, err := p.getClient(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	resp, err := client.ListTasks(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// CancelTask forwards the request to the gRPC agent.
func (p *grpcProxy) CancelTask(ctx context.Context, req *connect.Request[a2apb.CancelTaskRequest]) (*connect.Response[a2apb.Task], error) {
	ctx = withAgentHeaders(ctx)
	client, err := p.getClient(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	resp, err := client.CancelTask(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

// TaskSubscription forwards the streaming request to the gRPC agent.
func (p *grpcProxy) TaskSubscription(ctx context.Context, req *connect.Request[a2apb.TaskSubscriptionRequest], stream *connect.ServerStream[a2apb.StreamResponse]) error {
	ctx = withAgentHeaders(ctx)
	client, err := p.getClient(ctx)
	if err != nil {
		return connect.NewError(connect.CodeUnavailable, err)
	}
	grpcStream, err := client.TaskSubscription(ctx, req.Msg)
	if err != nil {
		return err
	}
	for {
		msg, err := grpcStream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := stream.Send(msg); err != nil {
			return err
		}
	}
}

func (p *grpcProxy) CreateTaskPushNotificationConfig(ctx context.Context, req *connect.Request[a2apb.CreateTaskPushNotificationConfigRequest]) (*connect.Response[a2apb.TaskPushNotificationConfig], error) {
	ctx = withAgentHeaders(ctx)
	client, err := p.getClient(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	resp, err := client.CreateTaskPushNotificationConfig(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (p *grpcProxy) GetTaskPushNotificationConfig(ctx context.Context, req *connect.Request[a2apb.GetTaskPushNotificationConfigRequest]) (*connect.Response[a2apb.TaskPushNotificationConfig], error) {
	ctx = withAgentHeaders(ctx)
	client, err := p.getClient(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	resp, err := client.GetTaskPushNotificationConfig(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (p *grpcProxy) ListTaskPushNotificationConfig(ctx context.Context, req *connect.Request[a2apb.ListTaskPushNotificationConfigRequest]) (*connect.Response[a2apb.ListTaskPushNotificationConfigResponse], error) {
	ctx = withAgentHeaders(ctx)
	client, err := p.getClient(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	resp, err := client.ListTaskPushNotificationConfig(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (p *grpcProxy) GetAgentCard(ctx context.Context, req *connect.Request[a2apb.GetAgentCardRequest]) (*connect.Response[a2apb.AgentCard], error) {
	ctx = withAgentHeaders(ctx)
	client, err := p.getClient(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	resp, err := client.GetAgentCard(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (p *grpcProxy) DeleteTaskPushNotificationConfig(ctx context.Context, req *connect.Request[a2apb.DeleteTaskPushNotificationConfigRequest]) (*connect.Response[emptypb.Empty], error) {
	ctx = withAgentHeaders(ctx)
	client, err := p.getClient(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	resp, err := client.DeleteTaskPushNotificationConfig(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}
