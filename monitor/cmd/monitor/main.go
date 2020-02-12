package main

import (
	"flag"
	"github.com/go-logr/logr"
	"github.com/seldonio/trtis-scheduler/monitor/k8s"
	"github.com/seldonio/trtis-scheduler/monitor/metric"
	"os"
	"os/signal"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"syscall"
	"time"
)

var (
	nodeName           = flag.String("node-name", "", "The node name")
	trtisHost        = flag.String("trtis-host", "0.0.0.0", "TRTIS host")
	trtisMetricsPort = flag.Int("trtis-http-port", 8002, "TRTIS http port")
)

func getTrtisHost(envVar, host string, log logr.Logger) string {
	envHost := os.Getenv(envVar)
	if envHost == "" {
		log.Info("Using TRTIS host from command line")
		return host
	} else {
		log.Info("Using TRTIS host from environment variable")
		return envHost
	}
}

func main() {
	flag.Parse()

	logf.SetLogger(logf.ZapLogger(false))
	log := logf.Log.WithName("proxy")
	log.Info("Started")

	nodeAnnotator, err := k8s.NewNodeAnnotator(*nodeName, log)
	if err != nil {
		log.Error(err, "Failed to get node annotator")
	}

	trtisMetrics := metric.NewTrtisMetrics(*trtisHost, *trtisMetricsPort, log)

	err = trtisMetrics.UpdateMetrics()
	if err != nil {
		log.Error(err, "Failed to get gpu metrics")
	} else {
		trtisMetrics.ShowMetrics()
		nodeAnnotator.PatchNodeAnnotation(trtisMetrics.GpuMetrics)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

	ticker := time.NewTicker(5 * time.Second)

	for {
		select {
		case _ = <-sigs:
			log.Info("Stopping")
			ticker.Stop()
			return
		case <-ticker.C:
			err = trtisMetrics.UpdateMetrics()
			if err != nil {
				log.Error(err, "Failed to get gpu metrics")
			} else {
				trtisMetrics.ShowMetrics()
				nodeAnnotator.PatchNodeAnnotation(trtisMetrics.GpuMetrics)
			}
		}
	}
}
