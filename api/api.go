package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/crosstyan/dumb_downloader/entity"
	"github.com/crosstyan/dumb_downloader/global/log"
	"github.com/samber/mo"
	"io"
	"net/http"
	"strconv"
	"time"
)

func getDownloadRequest(req *http.Request) (*entity.DownloadRequest, error) {
	dlReq := entity.DownloadRequest{}
	buf, err := io.ReadAll(req.Body)
	defer req.Body.Close()
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
	eR := entity.ErrorResponse{Error: err.Error()}
	b, _ := json.Marshal(eR)
	_, err = resp.Write(b)
	if err != nil {
		log.Sugar().Errorw("failed to write response", "error", err)
		return
	}
}

// MakeAsyncPushHandler creates a handler that pushes the request to the channel.
// @Summary Async Download
// @Description Push a download request to the queue
// @Tag download
// @Accept json
// @Produce json
// @Param request body entity.DownloadRequest true "download request"
// @Success 202
// @Failure 400 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /download [post]
func MakeAsyncPushHandler(
	reqChan chan<- entity.ReqResp,
	timeout time.Duration,
) http.HandlerFunc {
	pushQueue := func(resp http.ResponseWriter, req *http.Request) {
		var ctx, cancel = context.WithTimeout(req.Context(), timeout)
		defer cancel()
		dlReq, err := getDownloadRequest(req)
		log.Sugar().Infow("request", "url", dlReq.Url)
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
		case reqChan <- entity.ReqResp{Request: dlReq,
			ResponseChannel: mo.None[entity.ResponseChannelV](), Context: ctx, IsSync: false}:
			resp.WriteHeader(http.StatusAccepted)
			return
		case <-ctx.Done():
			writeError(resp, errors.New("timeout"), http.StatusGatewayTimeout)
			return
		}
	}
	return pushQueue
}

// MakeSyncPushHandler creates a sync handler that pushes the request to the channel.
// @Summary Sync Download
// @Description Push a download request to the queue and wait for the response
// @Tag download
// @Accept json
// @Produce json
// @Param transparent query bool false "If the response is transparent. See also strconv.ParseBool"
// @Param request body entity.DownloadRequest true "download request"
// @Success 200 {object} entity.DownloadResponse
// @Failure 400 {object} entity.ErrorResponse
// @Failure 500 {object} entity.ErrorResponse
// @Router /download/sync [post]
func MakeSyncPushHandler(
	reqChan chan<- entity.ReqResp,
) http.HandlerFunc {
	pushQueue := func(resp http.ResponseWriter, req *http.Request) {
		var ctx = req.Context()
		query := req.URL.Query()
		isTransparent := func() bool {
			t := query.Get("transparent")
			// log.Sugar().Debugw("query", "transparent", t)
			b, err := strconv.ParseBool(t)
			if err != nil {
				return false
			}
			return b
		}()
		dlReq, err := getDownloadRequest(req)
		if err != nil {
			log.Sugar().Errorw("request", "error", err)
			writeError(resp, err, http.StatusBadRequest)
			return
		}
		log.Sugar().Infow("request", "url", dlReq.Url, "isTransparent", isTransparent)
		respChan := make(chan entity.RespT)
		reqChan <- entity.ReqResp{Request: dlReq,
			ResponseChannel: mo.Some[entity.ResponseChannelV](respChan), Context: ctx, IsSync: true}
		select {
		case response := <-respChan:
			{
				r, err := response.Get()
				if err != nil {
					writeError(resp, err, http.StatusInternalServerError)
					return
				}
				if r == nil {
					writeError(resp, errors.New("nil response"), http.StatusInternalServerError)
					return
				}
				// https://pkg.go.dev/encoding/json#Marshal
				// https://www.alexedwards.net/blog/json-surprises-and-gotchas
				if !isTransparent {
					resp.Header().Add("Content-Type", "application/json")
					buf := bytes.NewBuffer([]byte{})
					enc := json.NewEncoder(buf)
					enc.SetEscapeHTML(false)
					err = enc.Encode(r)
					if err != nil {
						writeError(resp, err, http.StatusInternalServerError)
						return
					}
					_, err = resp.Write(buf.Bytes())
					resp.WriteHeader(http.StatusOK)
					if err != nil {
						log.Sugar().Errorw("failed to write response", "error", err)
						resp.WriteHeader(http.StatusInternalServerError)
						return
					}
				} else {
					for k, v := range r.Headers {
						resp.Header().Add(k, v)
					}
					resp.WriteHeader(r.StatusCode)
					_, err = resp.Write(r.Body)
					if err != nil {
						log.Sugar().Errorw("failed to write response", "error", err)
						resp.WriteHeader(http.StatusInternalServerError)
						return
					}
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
