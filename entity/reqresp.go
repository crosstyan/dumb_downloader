package entity

import (
	"context"
	"github.com/samber/mo"
)

type Resp = mo.Either[error, *DownloadResponse]

type ReqResp struct {
	Request *DownloadRequest
	// if this channel exists then we use sync API
	ResponseChannel *<-chan Resp
	IsSync          bool
	Context         context.Context
}
