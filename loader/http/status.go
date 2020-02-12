package http

import (
	"fmt"
	"github.com/go-logr/logr"
	"net/http"
	"time"
)

type ModelStatus struct {
	log logr.Logger
	url string
}

func NewModelStatus(host string, port int, modelName string, log logr.Logger) *ModelStatus {
	url := fmt.Sprintf("http://%s:%d/api/status/%s", host, port, modelName)
	return &ModelStatus{
		log: log,
		url: url,
	}
}

func (m *ModelStatus) isModelLoaded() (bool, error) {
	request, err := http.NewRequest("GET", m.url, nil)
	if err != nil {
		m.log.Error(err, "Failed to create request")
		return false, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		m.log.Error(err, "Status call failed")
		return false, err
	}
	if response.StatusCode == http.StatusOK {
		return true, nil
	} else {
		m.log.Info("Model not loaded", "status", response.StatusCode)
		return false, nil
	}
}

func (m *ModelStatus) WaitForModelLoaded() error {
	ok := false
	var err error
	for !ok {
		ok, err = m.isModelLoaded()
		if err != nil {
			m.log.Error(err, "Failed to get model status")
			return err
		}
		time.Sleep(time.Second * 2)
	}
	m.log.Info("Model loaded")
	return nil
}
