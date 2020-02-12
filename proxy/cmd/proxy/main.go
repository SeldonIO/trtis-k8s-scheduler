package main

import (
	"flag"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/seldonio/trtis-scheduler/proxy/grpc"
	trtis "github.com/seldonio/trtis-scheduler/proxy/proto/trtis"
	"net"
	"net/http"
	"net/http/httputil"
	url2 "net/url"
	"os"
	"os/signal"
	"path"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"syscall"
	"time"
)

var (
	grpcPort       = flag.Int("grpcPort", 9001, "grpc port")
	httpPort       = flag.Int("httpPort", 9000, "http port")
	trtisHost      = flag.String("trtis-host", "0.0.0.0", "TRTIS host")
	trtisGrpcPort  = flag.Int("trtis-grpc-port", 8001, "TRTIS grpc port")
	trtisHttpPort  = flag.Int("trtis-http-port", 8000, "TRTIS http port")
	modelName       = flag.String("model-name", "", "Model name")
	trtisModelRepo = flag.String("trtis-model-repo", "/mnt/trtis/models", "TRTIS Model Repository")
)

func startGrpcServer(log logr.Logger) {
	client, err := grpc.NewTrtisClient(*trtisHost, *trtisGrpcPort)
	if err != nil {
		log.Error(err, "Failed to create TRTIS client")
		os.Exit(-1)
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
	if err != nil {
		log.Error(err, "Failed to listen")
		os.Exit(-1)
	}
	server := grpc.CreateGrpcServer()
	proxy := grpc.NewTrtisProxy(client)
	trtis.RegisterGRPCServiceServer(server, proxy)

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
		_ = <-sigs
		log.Info("Received signal")
		server.GracefulStop()
	}()

	log.Info("grpc listening on ", "grpcPort", *grpcPort)
	err = server.Serve(lis)
	if err != nil {
		log.Error(err, "grpc server error")
	}

	log.Info("Stopping")
}

func startHttpProxy(log logr.Logger) {
	url, err := url2.Parse(fmt.Sprintf("http://%s:%d", *trtisHost, *trtisHttpPort))
	if err != nil {
		log.Error(err, "Failed to parse urr from ", "trtisHost", *trtisHost)
	}
	handler := httputil.NewSingleHostReverseProxy(url)

	address := fmt.Sprintf("0.0.0.0:%d", *httpPort)
	log.Info("Http Listening", "Address", address)

	srv := &http.Server{
		Handler: handler,
		Addr:    address,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error(err, "Server error")
		}
	}()
}

func removeModel(modelRepo, modelName string, log logr.Logger) {
	dstPath := path.Join(modelRepo, modelName)
	err := os.RemoveAll(dstPath)
	if err != nil {
		log.Error(err, "Failed to remove model")
	}
}


func main() {
	flag.Parse()

	logf.SetLogger(logf.ZapLogger(false))
	log := logf.Log.WithName("proxy")

	log.Info("Started")

	startHttpProxy(log)
	startGrpcServer(log)

	if *modelName != "" {
		log.Info("Cleaning model from ", "dst", *trtisModelRepo, "modelName", modelName)
		removeModel(*trtisModelRepo, *modelName, log)
	}
}
