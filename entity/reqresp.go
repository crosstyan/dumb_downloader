package entity

import (
	"context"
	"github.com/samber/mo"
)

// RespV is the inner type of the RespT (V for Value)
type RespV = *DownloadResponse

// RespT is the response type of the download request.
// See also ReqResp.ResponseChannel
type RespT = mo.Result[RespV]

type ResponseChannelV = chan<- RespT
type ResponseChannelT = mo.Option[ResponseChannelV]

type ReqResp struct {
	Request *DownloadRequest
	// if this channel exists then we use sync API
	ResponseChannel ResponseChannelT
	IsSync          bool
	Context         context.Context
}
