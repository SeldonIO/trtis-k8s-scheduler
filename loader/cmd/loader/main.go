package main

import (
	"flag"
	"github.com/go-logr/logr"
	"github.com/otiai10/copy"
	http2 "github.com/seldonio/trtis-scheduler/loader/http"
	"os"
	"path"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var (
	trtisHost      = flag.String("trtis-host", "0.0.0.0", "TRTIS host")
	trtisHttpPort  = flag.Int("trtis-http-port", 8000, "TRTIS http port")
	modelSrc       = flag.String("model-src", "", "Src folder for model")
	trtisModelRepo = flag.String("trtis-model-repo", "/mnt/trtis/models", "TRTIS Model Repository")
)

// Copy model to dst folder
// Assumes last pasrt of model is the model name and appends this to dst
func copyModel(src, dst, modelName string, log logr.Logger) {
	dstPath := path.Join(dst, modelName)
	err := copy.Copy(src, dstPath)
	if err != nil {
		log.Error(err, "failed to copy model")
		os.Exit(-1)
	}
}

func main() {
	flag.Parse()

	logf.SetLogger(logf.ZapLogger(false))
	log := logf.Log.WithName("proxy")

	if *modelSrc == "" {
		log.Info("model-src must be provided")
		os.Exit(-1)
	}

	log.Info("Started")

	_, modelName := path.Split(*modelSrc)
	log.Info("Copy model from ", "src", *modelSrc, "dst", *trtisModelRepo, "model-name", modelName)
	copyModel(*modelSrc, *trtisModelRepo, modelName, log)

	//TODO wait for TRTIS to show model is loaded and change annotation on this pod to show allocated
	// so memory will not be added to that shown in TRTIS by scheduler
	modelStatus := http2.NewModelStatus(*trtisHost, *trtisHttpPort, modelName, log)
	err := modelStatus.WaitForModelLoaded()
	if err != nil {
		os.Exit(-1)
	}
}
