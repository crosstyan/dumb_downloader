package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/crosstyan/dumb_downloader/entity"
	"github.com/crosstyan/dumb_downloader/log"
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

func writeError(resp http.ResponseWriter, err error, code int) {
	resp.WriteHeader(code)
	eR := entity.ErrorResponse{Error: err}
	b, _ := eR.MarshalJSON()
	_, err = resp.Write(b)
	if err != nil {
		log.Sugar().Errorw("failed to write response", "error", err)
		return
	}
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
			writeError(resp, err, http.StatusBadRequest)
			return
		}
		// if save output is not set, it's meaningless to use async API
		if !dlReq.IsSaveOutput {
			writeError(resp, errors.New("async API only accepts save output"), http.StatusBadRequest)
			return
		}
		select {
		case reqChan <- dlReq:
			resp.WriteHeader(http.StatusAccepted)
			return
		case <-ctx.Done():
			writeError(resp, errors.New("timeout"), http.StatusGatewayTimeout)
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
