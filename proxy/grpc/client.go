package grpc

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	nvidia_inferenceserver "github.com/seldonio/trtis-scheduler/proxy/proto/trtis"
	"google.golang.org/grpc"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type TrtisClient struct {
	Log  logr.Logger
	conn *grpc.ClientConn
}

func NewTrtisClient(host string, port int) (*TrtisClient, error) {

	conn, err := getConnection(host, port)
	if err != nil {
		return nil, err
	}
	client := TrtisClient{
		Log:  logf.Log.WithName("TrtisClient"),
		conn: conn,
	}
	return &client, nil
}

func getConnection(host string, port int) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, port), opts...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (t TrtisClient) Status(ctx context.Context, in *nvidia_inferenceserver.StatusRequest, opts ...grpc.CallOption) (*nvidia_inferenceserver.StatusResponse, error) {
	t.Log.Info("Status called")
	client := nvidia_inferenceserver.NewGRPCServiceClient(t.conn)
	return client.Status(ctx, in, opts...)
}

func (t TrtisClient) Health(ctx context.Context, in *nvidia_inferenceserver.HealthRequest, opts ...grpc.CallOption) (*nvidia_inferenceserver.HealthResponse, error) {
	panic("implement me")
}

func (t TrtisClient) Infer(ctx context.Context, in *nvidia_inferenceserver.InferRequest, opts ...grpc.CallOption) (*nvidia_inferenceserver.InferResponse, error) {
	t.Log.Info("Infer called")
	client := nvidia_inferenceserver.NewGRPCServiceClient(t.conn)
	return client.Infer(ctx, in, opts...)
}

func (t TrtisClient) StreamInfer(ctx context.Context, opts ...grpc.CallOption) (nvidia_inferenceserver.GRPCService_StreamInferClient, error) {
	panic("implement me")
}

func (t TrtisClient) ModelControl(ctx context.Context, in *nvidia_inferenceserver.ModelControlRequest, opts ...grpc.CallOption) (*nvidia_inferenceserver.ModelControlResponse, error) {
	panic("implement me")
}

func (t TrtisClient) SharedMemoryControl(ctx context.Context, in *nvidia_inferenceserver.SharedMemoryControlRequest, opts ...grpc.CallOption) (*nvidia_inferenceserver.SharedMemoryControlResponse, error) {
	panic("implement me")
}

func (t TrtisClient) Repository(ctx context.Context, in *nvidia_inferenceserver.RepositoryRequest, opts ...grpc.CallOption) (*nvidia_inferenceserver.RepositoryResponse, error) {
	panic("implement me")
}
