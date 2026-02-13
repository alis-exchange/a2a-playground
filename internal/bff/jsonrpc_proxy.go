package bff

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sync"

	"connectrpc.com/connect"
	"github.com/a2aproject/a2a-go/a2a"
	"github.com/a2aproject/a2a-go/a2aclient"
	"github.com/a2aproject/a2a-go/a2apb"
	"github.com/a2aproject/a2a-go/a2apb/pbconv"
	"github.com/alis-exchange/a2a-playground/gen/go/a2apbconnect"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
)

// jsonrpcProxy implements A2AServiceHandler by forwarding requests to a JSON-RPC agent via a2aclient.
type jsonrpcProxy struct {
	agentURL string
	mu       sync.Mutex
	client   *a2aclient.Client
}

// NewJSONRPCProxy creates a proxy that forwards to the JSON-RPC agent at agentURL.
// agentURL must be a full URL e.g. http://localhost:8080/jsonrpc.
func NewJSONRPCProxy(agentURL string) *jsonrpcProxy {
	return &jsonrpcProxy{agentURL: agentURL}
}

// getClient returns or creates the a2aclient, connecting lazily on first use.
func (p *jsonrpcProxy) getClient(ctx context.Context) (*a2aclient.Client, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.client != nil {
		return p.client, nil
	}
	endpoints := []a2a.AgentInterface{
		{URL: p.agentURL, Transport: a2a.TransportProtocolJSONRPC},
	}
	client, err := a2aclient.NewFromEndpoints(ctx, endpoints,
		a2aclient.WithDefaultsDisabled(),
		a2aclient.WithJSONRPCTransport(nil),
		a2aclient.WithInterceptors(&agentHeadersInterceptor{}),
	)
	if err != nil {
		return nil, err
	}
	p.client = client
	return p.client, nil
}

// ConnectOptions returns Connect-RPC handler options for the proxy.
func (p *jsonrpcProxy) ConnectOptions() []connect.HandlerOption {
	return nil
}

// Handler returns the path and HTTP handler for mounting the A2A Connect service.
func (p *jsonrpcProxy) Handler() (string, http.Handler) {
	return a2apbconnect.NewA2AServiceHandler(p, p.ConnectOptions()...)
}

// Ensure jsonrpcProxy implements a2apbconnect.A2AServiceHandler.
var _ a2apbconnect.A2AServiceHandler = (*jsonrpcProxy)(nil)

// agentHeadersInterceptor reads agent headers from context and injects them into req.Meta.
type agentHeadersInterceptor struct {
	a2aclient.PassthroughInterceptor
}

func (agentHeadersInterceptor) Before(ctx context.Context, req *a2aclient.Request) (context.Context, error) {
	headers := AgentHeadersFromContext(ctx)
	for k, v := range headers {
		req.Meta.Append(k, v)
	}
	return ctx, nil
}

// taskIDFromName extracts task ID from a "tasks/{id}" name field.
func taskIDFromName(name string) (a2a.TaskID, error) {
	return pbconv.ExtractTaskID(name)
}

// SendMessage forwards the request to the JSON-RPC agent.
func (p *jsonrpcProxy) SendMessage(ctx context.Context, req *connect.Request[a2apb.SendMessageRequest]) (*connect.Response[a2apb.SendMessageResponse], error) {
	client, err := p.getClient(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	params, err := pbconv.FromProtoSendMessageRequest(req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	result, err := client.SendMessage(ctx, params)
	if err != nil {
		return nil, toConnectError(err)
	}
	resp, err := pbconv.ToProtoSendMessageResponse(result)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(resp), nil
}

// SendStreamingMessage forwards the streaming request to the JSON-RPC agent.
func (p *jsonrpcProxy) SendStreamingMessage(ctx context.Context, req *connect.Request[a2apb.SendMessageRequest], stream *connect.ServerStream[a2apb.StreamResponse]) error {
	client, err := p.getClient(ctx)
	if err != nil {
		return connect.NewError(connect.CodeUnavailable, err)
	}
	params, err := pbconv.FromProtoSendMessageRequest(req.Msg)
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	for event, err := range client.SendStreamingMessage(ctx, params) {
		if err != nil {
			return toConnectError(err)
		}
		protoEvent, err := pbconv.ToProtoStreamResponse(event)
		if err != nil {
			return connect.NewError(connect.CodeInternal, err)
		}
		if err := stream.Send(protoEvent); err != nil {
			return err
		}
	}
	return nil
}

// GetTask forwards the request to the JSON-RPC agent.
func (p *jsonrpcProxy) GetTask(ctx context.Context, req *connect.Request[a2apb.GetTaskRequest]) (*connect.Response[a2apb.Task], error) {
	client, err := p.getClient(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	params, err := pbconv.FromProtoGetTaskRequest(req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	task, err := client.GetTask(ctx, params)
	if err != nil {
		return nil, toConnectError(err)
	}
	protoTask, err := pbconv.ToProtoTask(task)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(protoTask), nil
}

// ListTasks returns CodeUnimplemented; JSON-RPC has no tasks/list per A2A spec.
func (p *jsonrpcProxy) ListTasks(context.Context, *connect.Request[a2apb.ListTasksRequest]) (*connect.Response[a2apb.ListTasksResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("ListTasks is not supported for JSON-RPC agents"))
}

// CancelTask forwards the request to the JSON-RPC agent.
func (p *jsonrpcProxy) CancelTask(ctx context.Context, req *connect.Request[a2apb.CancelTaskRequest]) (*connect.Response[a2apb.Task], error) {
	client, err := p.getClient(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	taskID, err := taskIDFromName(req.Msg.GetName())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	task, err := client.CancelTask(ctx, &a2a.TaskIDParams{ID: taskID})
	if err != nil {
		return nil, toConnectError(err)
	}
	protoTask, err := pbconv.ToProtoTask(task)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(protoTask), nil
}

// TaskSubscription forwards the streaming request to the JSON-RPC agent.
func (p *jsonrpcProxy) TaskSubscription(ctx context.Context, req *connect.Request[a2apb.TaskSubscriptionRequest], stream *connect.ServerStream[a2apb.StreamResponse]) error {
	client, err := p.getClient(ctx)
	if err != nil {
		return connect.NewError(connect.CodeUnavailable, err)
	}
	taskID, err := taskIDFromName(req.Msg.GetName())
	if err != nil {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}
	for event, err := range client.ResubscribeToTask(ctx, &a2a.TaskIDParams{ID: taskID}) {
		if err != nil {
			return toConnectError(err)
		}
		protoEvent, err := pbconv.ToProtoStreamResponse(event)
		if err != nil {
			return connect.NewError(connect.CodeInternal, err)
		}
		if err := stream.Send(protoEvent); err != nil {
			return err
		}
	}
	return nil
}

// CreateTaskPushNotificationConfig forwards the request to the JSON-RPC agent.
func (p *jsonrpcProxy) CreateTaskPushNotificationConfig(ctx context.Context, req *connect.Request[a2apb.CreateTaskPushNotificationConfigRequest]) (*connect.Response[a2apb.TaskPushNotificationConfig], error) {
	client, err := p.getClient(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	config, err := pbconv.FromProtoCreateTaskPushConfigRequest(req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	result, err := client.SetTaskPushConfig(ctx, config)
	if err != nil {
		return nil, toConnectError(err)
	}
	protoConfig, err := pbconv.ToProtoTaskPushConfig(result)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(protoConfig), nil
}

// GetTaskPushNotificationConfig forwards the request to the JSON-RPC agent.
func (p *jsonrpcProxy) GetTaskPushNotificationConfig(ctx context.Context, req *connect.Request[a2apb.GetTaskPushNotificationConfigRequest]) (*connect.Response[a2apb.TaskPushNotificationConfig], error) {
	client, err := p.getClient(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	params, err := pbconv.FromProtoGetTaskPushConfigRequest(req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	config, err := client.GetTaskPushConfig(ctx, params)
	if err != nil {
		return nil, toConnectError(err)
	}
	protoConfig, err := pbconv.ToProtoTaskPushConfig(config)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(protoConfig), nil
}

// ListTaskPushNotificationConfig forwards the request to the JSON-RPC agent.
func (p *jsonrpcProxy) ListTaskPushNotificationConfig(ctx context.Context, req *connect.Request[a2apb.ListTaskPushNotificationConfigRequest]) (*connect.Response[a2apb.ListTaskPushNotificationConfigResponse], error) {
	client, err := p.getClient(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	taskID, err := pbconv.ExtractTaskID(req.Msg.GetParent())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	params := &a2a.ListTaskPushConfigParams{TaskID: taskID}
	configs, err := client.ListTaskPushConfig(ctx, params)
	if err != nil {
		return nil, toConnectError(err)
	}
	protoResp, err := pbconv.ToProtoListTaskPushConfig(configs)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(protoResp), nil
}

// DeleteTaskPushNotificationConfig forwards the request to the JSON-RPC agent.
func (p *jsonrpcProxy) DeleteTaskPushNotificationConfig(ctx context.Context, req *connect.Request[a2apb.DeleteTaskPushNotificationConfigRequest]) (*connect.Response[emptypb.Empty], error) {
	client, err := p.getClient(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	params, err := pbconv.FromProtoDeleteTaskPushConfigRequest(req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}
	if err := client.DeleteTaskPushConfig(ctx, params); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

// toConnectError maps a2a errors to Connect codes.
func toConnectError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, io.EOF) {
		return nil
	}
	var a2aErr *a2a.Error
	if errors.As(err, &a2aErr) {
		switch {
		case errors.Is(a2aErr.Err, a2a.ErrTaskNotFound):
			return connect.NewError(connect.CodeNotFound, err)
		case errors.Is(a2aErr.Err, a2a.ErrInvalidRequest), errors.Is(a2aErr.Err, a2a.ErrInvalidParams):
			return connect.NewError(connect.CodeInvalidArgument, err)
		case errors.Is(a2aErr.Err, a2a.ErrMethodNotFound):
			return connect.NewError(connect.CodeUnimplemented, err)
		case errors.Is(a2aErr.Err, a2a.ErrUnauthenticated):
			return connect.NewError(connect.CodeUnauthenticated, err)
		case errors.Is(a2aErr.Err, a2a.ErrUnauthorized):
			return connect.NewError(connect.CodePermissionDenied, err)
		}
	}
	return connect.NewError(connect.CodeUnknown, err)
}

// GetAgentCard forwards the request to the JSON-RPC agent.
func (p *jsonrpcProxy) GetAgentCard(ctx context.Context, req *connect.Request[a2apb.GetAgentCardRequest]) (*connect.Response[a2apb.AgentCard], error) {
	client, err := p.getClient(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, err)
	}
	card, err := client.GetAgentCard(ctx)
	if err != nil {
		return nil, toConnectError(err)
	}
	protoCard, err := pbconv.ToProtoAgentCard(card)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(protoCard), nil
}

