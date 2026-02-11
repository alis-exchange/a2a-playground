package bff

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"

	"connectrpc.com/connect"
	"github.com/alis-exchange/a2a-playground/gen/go/a2apb"
	"github.com/alis-exchange/a2a-playground/gen/go/a2apbconnect"
	"go.alis.build/client/v2"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// A2AProxy implements A2AServiceHandler by forwarding requests to a gRPC agent.
type A2AProxy struct {
	agentURL string
	mu       sync.Mutex
	client   a2apb.A2AServiceClient
}

// NewA2AProxy creates a proxy that forwards to the gRPC agent at agentURL.
// agentURL may be "localhost:8080" or "http://localhost:8080"; the scheme is stripped for gRPC.
// Connection is established lazily on first request so startup does not fail if the agent is unreachable.
func NewA2AProxy(agentURL string) *A2AProxy {
	return &A2AProxy{agentURL: agentURL}
}

// getClient returns or creates the gRPC client, connecting lazily on first use.
func (p *A2AProxy) getClient(ctx context.Context) (a2apb.A2AServiceClient, error) {
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

// NewA2AProxyFromConn creates a proxy from an existing gRPC connection.
func NewA2AProxyFromConn(client a2apb.A2AServiceClient) *A2AProxy {
	return &A2AProxy{client: client}
}

// ConnectOptions returns Connect-RPC handler options for the proxy.
func (p *A2AProxy) ConnectOptions() []connect.HandlerOption {
	return nil
}

// Handler returns the path and HTTP handler for mounting the A2A Connect service.
func (p *A2AProxy) Handler() (string, http.Handler) {
	return a2apbconnect.NewA2AServiceHandler(p, p.ConnectOptions()...)
}

// Ensure A2AProxy implements a2apbconnect.A2AServiceHandler.
var _ a2apbconnect.A2AServiceHandler = (*A2AProxy)(nil)

// SendMessage forwards the request to the gRPC agent.
func (p *A2AProxy) SendMessage(ctx context.Context, req *connect.Request[a2apb.SendMessageRequest]) (*connect.Response[a2apb.SendMessageResponse], error) {
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
func (p *A2AProxy) SendStreamingMessage(ctx context.Context, req *connect.Request[a2apb.SendMessageRequest], stream *connect.ServerStream[a2apb.StreamResponse]) error {
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
func (p *A2AProxy) GetTask(ctx context.Context, req *connect.Request[a2apb.GetTaskRequest]) (*connect.Response[a2apb.Task], error) {
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
func (p *A2AProxy) ListTasks(ctx context.Context, req *connect.Request[a2apb.ListTasksRequest]) (*connect.Response[a2apb.ListTasksResponse], error) {
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
func (p *A2AProxy) CancelTask(ctx context.Context, req *connect.Request[a2apb.CancelTaskRequest]) (*connect.Response[a2apb.Task], error) {
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

// SubscribeToTask forwards the streaming request to the gRPC agent.
func (p *A2AProxy) SubscribeToTask(ctx context.Context, req *connect.Request[a2apb.SubscribeToTaskRequest], stream *connect.ServerStream[a2apb.StreamResponse]) error {
	client, err := p.getClient(ctx)
	if err != nil {
		return connect.NewError(connect.CodeUnavailable, err)
	}
	grpcStream, err := client.SubscribeToTask(ctx, req.Msg)
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

func (p *A2AProxy) CreateTaskPushNotificationConfig(ctx context.Context, req *connect.Request[a2apb.CreateTaskPushNotificationConfigRequest]) (*connect.Response[a2apb.TaskPushNotificationConfig], error) {
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

func (p *A2AProxy) GetTaskPushNotificationConfig(ctx context.Context, req *connect.Request[a2apb.GetTaskPushNotificationConfigRequest]) (*connect.Response[a2apb.TaskPushNotificationConfig], error) {
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

func (p *A2AProxy) ListTaskPushNotificationConfig(ctx context.Context, req *connect.Request[a2apb.ListTaskPushNotificationConfigRequest]) (*connect.Response[a2apb.ListTaskPushNotificationConfigResponse], error) {
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

func (p *A2AProxy) GetExtendedAgentCard(ctx context.Context, req *connect.Request[a2apb.GetExtendedAgentCardRequest]) (*connect.Response[a2apb.AgentCard], error) {
	client, err := p.getClient(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	resp, err := client.GetExtendedAgentCard(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (p *A2AProxy) DeleteTaskPushNotificationConfig(ctx context.Context, req *connect.Request[a2apb.DeleteTaskPushNotificationConfigRequest]) (*connect.Response[emptypb.Empty], error) {
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
