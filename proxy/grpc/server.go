package grpc

import (
	"context"
	"github.com/go-logr/logr"
	trtis "github.com/seldonio/trtis-scheduler/proxy/proto/trtis"
	"google.golang.org/grpc"
	"math"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

func CreateGrpcServer() *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(math.MaxInt32),
		grpc.MaxSendMsgSize(math.MaxInt32),
	}
	grpcServer := grpc.NewServer(opts...)
	return grpcServer
}

type TrtisProxy struct {
	Log         logr.Logger
	client      *TrtisClient
	callOptions []grpc.CallOption
}

func NewTrtisProxy(client *TrtisClient) *TrtisProxy {
	opts := []grpc.CallOption{
		grpc.MaxCallSendMsgSize(math.MaxInt32),
		grpc.MaxCallRecvMsgSize(math.MaxInt32),
	}
	return &TrtisProxy{
		Log:         logf.Log.WithName("TrtisServer"),
		client:      client,
		callOptions: opts,
	}
}
func (t *TrtisProxy) Infer(ctx context.Context, req *trtis.InferRequest) (*trtis.InferResponse, error) {
	t.Log.Info("Infer called")
	return t.client.Infer(ctx, req, t.callOptions...)
}

func (t *TrtisProxy) Status(ctx context.Context, req *trtis.StatusRequest) (*trtis.StatusResponse, error) {
	t.Log.Info("Status called")
	return t.client.Status(ctx, req, t.callOptions...)
}

func (t *TrtisProxy) Health(context.Context, *trtis.HealthRequest) (*trtis.HealthResponse, error) {
	panic("implement me")
}

func (t *TrtisProxy) StreamInfer(trtis.GRPCService_StreamInferServer) error {
	panic("implement me")
}

func (t *TrtisProxy) ModelControl(context.Context, *trtis.ModelControlRequest) (*trtis.ModelControlResponse, error) {
	panic("implement me")
}

func (t *TrtisProxy) SharedMemoryControl(context.Context, *trtis.SharedMemoryControlRequest) (*trtis.SharedMemoryControlResponse, error) {
	panic("implement me")
}

func (t *TrtisProxy) Repository(context.Context, *trtis.RepositoryRequest) (*trtis.RepositoryResponse, error) {
	panic("implement me")
}
