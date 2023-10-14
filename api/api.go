package api

import (
	"context"
	"encoding/json"
	"github.com/crosstyan/dumb_downloader/entity"
	"net/http"
	"time"
)

func GetDownloadRequest(req *http.Request) (*entity.DownloadRequest, error) {
	dlReq := entity.DownloadRequest{}
	buf := make([]byte, 1024)
	_, err := req.Body.Read(buf)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(buf, &dlReq)
	if err != nil {
		return nil, err
	}
	return &dlReq, nil
}

// MakeAsyncPushHandler creates a handler that pushes the request to the channel.
func MakeAsyncPushHandler(
	reqChan chan<- *entity.DownloadRequest,
	timeout time.Duration,
) http.HandlerFunc {
	pushQueue := func(resp http.ResponseWriter, req *http.Request) {
		var ctx, cancel = context.WithTimeout(req.Context(), timeout)
		defer cancel()
		dlReq, err := GetDownloadRequest(req)
		if err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			return
		}
		select {
		case reqChan <- dlReq:
			resp.WriteHeader(http.StatusAccepted)
			return
		case <-ctx.Done():
			resp.WriteHeader(http.StatusServiceUnavailable)
			return
		}
	}
	return pushQueue
}

func MakeSyncPushHandler(
	reqChan chan<- *entity.DownloadRequest,
	respChan <-chan *entity.DownloadResponse,
	timeout time.Duration,
) http.HandlerFunc {
	return nil
}
