package cmd

import (
	"context"
	"github.com/crosstyan/dumb_downloader/api"
	"github.com/crosstyan/dumb_downloader/entity"
	"github.com/crosstyan/dumb_downloader/global/log"
	"github.com/crosstyan/dumb_downloader/utils"
	"github.com/imroc/req/v3"
	"github.com/joomcode/errorx"
	"github.com/panjf2000/ants/v2"
	"github.com/samber/mo"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/spf13/cobra"
	httpSwagger "github.com/swaggo/http-swagger/v2"
	"moul.io/chizap"

	// swagger embed files
	_ "github.com/crosstyan/dumb_downloader/docs"
)

const ChannelSize = 128

func makeSubDirectory(baseOutDir string, sub string) (string, error) {
	outDir := path.Join(baseOutDir, sub)
	stat, err := os.Stat(outDir)
	if !os.IsNotExist(err) {
		if !stat.IsDir() {
			return "", errorx.IllegalArgument.New("output directory %s is not a directory", outDir)
		}
	} else {
		err = os.MkdirAll(outDir, 0755)
		if err != nil {
			return "", errorx.Decorate(err, "failed to create directory %s", outDir)
		}
	}
	return outDir, nil
}

// @title Dumb Downloader API
// @version 1.0
// @license.name Do What the Fuck You Want to Public License
// @license.url http://www.wtfpl.net/
func serveRun(cmd *cobra.Command, args []string) {
	listenAddr, err := GetListenAddrFromViper()
	if err != nil {
		log.Sugar().Panicw("failed to get listen address", "error", err)
	}
	log.Sugar().Infow("listen", "addr", listenAddr)
	baseOutDir, err := GetOutDirFromViper()
	if err != nil {
		log.Sugar().Panicw("no valid output directory", "output_dir", baseOutDir)
	}
	log.Sugar().Infow("use base output directory", "output_dir", baseOutDir)
	poolSize, err := GetPoolSizeFromViper()
	if err != nil {
		log.Sugar().Panicw("bad pool size", "pool_size", poolSize)
	}
	log.Sugar().Infow("use pool size", "pool_size", poolSize)
	r := chi.NewRouter()
	// middleware
	chiZapM := chizap.New(log.Logger(), &chizap.Opts{})
	corsM := cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})
	swaggerH := httpSwagger.Handler(
		// The url pointing to API definition
		// this is a magic path...
		httpSwagger.URL("/swagger/doc.json"),
	)
	ch := make(chan entity.ReqResp, ChannelSize)
	ctx := context.Background()
	po, err := ants.NewPool(poolSize)
	if err != nil {
		log.Sugar().Panicw("failed to create pool", "error", err, "pool_size", poolSize)
	}

	client := req.C().ImpersonateChrome()
	_, f, err := GetHttpProxyFromViper()
	if err != nil {
		log.Sugar().Infow("no http proxy", "error", err.Error())
	} else {
		client = client.SetProxy(f)
	}
	for i := range make([]struct{}, poolSize) {
		err = po.Submit(func() {
			tryDownload(ctx, ch, client, baseOutDir)
		})
		if err != nil {
			log.Sugar().Panicw("failed to submit task", "error", err, "iteration", i)
		}
	}
	r.Use(chiZapM, corsM)
	one := time.Second
	// dumb swagger handler
	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/", http.StatusMovedPermanently)
	})
	r.Get("/swagger/*", swaggerH)
	r.Post("/download/sync", api.MakeSyncPushHandler(ch))
	r.Post("/download", api.MakeAsyncPushHandler(ch, one))
	err = http.ListenAndServe(listenAddr, r)
	if err != nil {
		log.Sugar().Panicw("listen", "err", err)
	}
}

func tryDownload(ctx context.Context, reqChan <-chan entity.ReqResp, client *req.Client, baseOutDir string) {
	for {
		select {
		case <-ctx.Done():
			return
		case reqResp := <-reqChan:
			{
				r := reqResp.Request
				if r == nil {
					log.Sugar().Errorw("nil request")
					continue
				}
				R := client.R()
				cookies := utils.Map(r.Cookies, func(c http.Cookie) *http.Cookie { return &c })
				R.SetCookies(cookies...)
				// don't break the impersonation
				for k, v := range r.Headers {
					R.SetHeader(k, v)
				}
				var resp *req.Response
				var err error
				// if it's async we could just use this goroutine to get the response
				if !reqResp.IsSync {
					resp, err = R.Get(r.Url)
				} else {
					// otherwise we have to poll the context
					type ResponseV = *req.Response
					c := make(chan mo.Result[ResponseV])
					go func() {
						resp, err := R.Get(r.Url)
						if err != nil {
							c <- mo.Err[ResponseV](err)
							return
						}
						c <- mo.Ok[ResponseV](resp)
					}()
					select {
					case <-ctx.Done():
						log.Sugar().Warnw("request context cancelled", "url", r.Url)
						continue
					case result := <-c:
						resp, err = result.Get()
					}
				}
				reCh, chOk := reqResp.ResponseChannel.Get()
				if err != nil {
					if chOk && reqResp.IsSync {
						reCh <- mo.Err[entity.RespV](err)
					}
					log.Sugar().Errorw("failed to download", "url", r.Url, "error", err)
					utils.PrintHeadersCookies(R)
					continue
				}
				if chOk && reqResp.IsSync {
					dlR := entity.DownloadResponse{}
					header := make(map[string]string)
					dlR.Headers = header
					if resp.Header != nil {
						for k, v := range resp.Header {
							vv := strings.Join(v, ",")
							dlR.Headers[k] = vv
						}
					}
					ct, ok := utils.TryGet(dlR.Headers, "Content-Type", "content-type", "Content-type", "content-Type", "Content-TYPE").Get()
					if ok {
						dlR.MIMEType = ct
					}
					dlR.StatusCode = resp.StatusCode
					dlR.Url = r.Url
					dlR.Body = resp.Bytes()

					reCh <- mo.Ok[entity.RespV](&dlR)
				}
				if r.OutPrefix == nil {
					log.Sugar().Infow("proxy", "url", r.Url)
					continue
				}
				isGoodStatusCode := func() bool {
					return resp.StatusCode >= 200 && resp.StatusCode < 300
				}()
				// only save image type
				ct := resp.Header.Get("Content-Type")
				// TODO: custom content type
				if !strings.Contains(ct, "image") || !isGoodStatusCode {
					log.Sugar().Errorw("bad response", "url", r.Url, "Content-Type", ct, "status", resp.StatusCode)
					utils.PrintHeadersCookies(R)
					log.Sugar().Debugw("response", "headers", resp.Header, "response", resp.String())
					return
				}
				var out string
				prefix := *r.OutPrefix
				if prefix == "" {
					out = baseOutDir
				} else {
					out, err = makeSubDirectory(baseOutDir, *r.OutPrefix)
					if err != nil {
						log.Sugar().Errorw("failed to create sub directory", "error", err, "url", r.Url, "prefix", prefix, "fallback", baseOutDir)
						out = baseOutDir
					}
				}
				err = os.WriteFile(out, resp.Bytes(), 0644)
				if err != nil {
					log.Sugar().Errorw("failed to save", "url", r.Url, "output", out, "error", err)
					continue
				} else {
					log.Sugar().Infow("downloaded", "url", r.Url, "output", out)
				}
			}
		}
	}
}

var serve = cobra.Command{
	Use:   "serve",
	Short: "serve a dumb downloader server",
	Args:  cobra.NoArgs,
	Run:   serveRun,
}
