package entity

import (
	"context"
	"github.com/samber/mo"
)

// RespIn is the inner type of the Resp
type RespIn = *DownloadResponse

// Resp is the response type of the download request.
// See also ReqResp.ResponseChannel
type Resp = mo.Result[RespIn]

type ReqResp struct {
	Request *DownloadRequest
	// if this channel exists then we use sync API
	ResponseChannel *chan<- Resp
	IsSync          bool
	Context         context.Context
}
