package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/crosstyan/dumb_downloader/entity"
	"github.com/crosstyan/dumb_downloader/global/log"
	"github.com/samber/mo"
	"net/http"
	"time"
)

type Resp = mo.Either[error, *entity.DownloadResponse]

type ReqResp struct {
	Request *entity.DownloadRequest
	// if this channel exists then we use sync API
	ResponseChannel *<-chan Resp
	isSync          bool
	context         context.Context
}

// Context returns the context of the request.
// it's only useful when the request is sync.
func (rr *ReqResp) Context() context.Context {
	return rr.context
}

func (rr *ReqResp) IsSync() bool {
	return rr.isSync
}

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
	b, _ := json.Marshal(eR)
	_, err = resp.Write(b)
	if err != nil {
		log.Sugar().Errorw("failed to write response", "error", err)
		return
	}
}

// MakeAsyncPushHandler creates a handler that pushes the request to the channel.
func MakeAsyncPushHandler(
	reqChan chan<- ReqResp,
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
		if dlReq.OutPrefix == nil {
			writeError(resp, errors.New("async API only accepts save output"), http.StatusBadRequest)
			return
		}
		select {
		case reqChan <- ReqResp{Request: dlReq, ResponseChannel: nil, context: ctx, isSync: false}:
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
	reqChan chan<- ReqResp,
) http.HandlerFunc {
	pushQueue := func(resp http.ResponseWriter, req *http.Request) {
		var ctx = req.Context()
		dlReq, err := GetDownloadRequest(req)
		if err != nil {
			writeError(resp, err, http.StatusBadRequest)
			return
		}
		respChan := make(chan Resp)
		oneWay := (<-chan Resp)(respChan)
		select {
		case reqChan <- ReqResp{Request: dlReq, ResponseChannel: &oneWay, context: ctx, isSync: true}:
		case response := <-respChan:
			{
				err, r := response.Unpack()
				if err != nil {
					writeError(resp, err, http.StatusInternalServerError)
					return
				}
				b, err := json.Marshal(r)
				if err != nil {
					writeError(resp, err, http.StatusInternalServerError)
					return
				}
				resp.WriteHeader(http.StatusOK)
				_, err = resp.Write(b)
				if err != nil {
					log.Sugar().Errorw("failed to write response", "error", err)
					return
				}
				return
			}
		case <-ctx.Done():
			writeError(resp, errors.New("timeout"), http.StatusGatewayTimeout)
			return
		}
	}
	return pushQueue
}
