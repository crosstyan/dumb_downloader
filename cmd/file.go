package cmd

import (
	"encoding/json"
	"github.com/crosstyan/dumb_downloader/global/log"
	"github.com/panjf2000/ants/v2"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/crosstyan/dumb_downloader/entity"
	"github.com/crosstyan/dumb_downloader/utils"
	"github.com/imroc/req/v3"
	"github.com/joomcode/errorx"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
)

type ProxyFunc = func(*http.Request) (*url.URL, error)

func GetDescription(descriptionPath string) (*entity.Description, error) {
	if _, err := os.Stat(descriptionPath); os.IsNotExist(err) {
		return nil, errorx.Decorate(err, "config file %s not found", descriptionPath)
	}
	content, err := os.ReadFile(descriptionPath)
	if err != nil {
		return nil, errorx.Decorate(err, "failed to read config file %s", descriptionPath)
	}
	d := entity.Description{}
	err = json.Unmarshal(content, &d)
	if err != nil {
		return nil, errorx.Decorate(err, "failed to parse config file %s", descriptionPath)
	}
	return &d, nil
}

func GetLatestRequest(description *entity.Description) (*entity.RequestRecord, error) {
	// https://github.com/panjf2000/ants
	reqs := maps.Values(description.Requests)
	if len(reqs) == 0 {
		return nil, errorx.IllegalArgument.New("no requests found in entity")
	}
	sort.Slice(reqs, func(i, j int) bool {
		return reqs[i].Timestamp.After(reqs[j].Timestamp)
	})
	latest := reqs[0]
	return &latest, nil
}

func runDescription(cmd *cobra.Command, args []string) {
	target := args[0]
	d, err := GetDescription(target)
	if err != nil {
		log.Sugar().Panicw("failed to get entity", "error", err)
	}
	latest, err := GetLatestRequest(d)
	if err != nil {
		log.Sugar().Panicw("failed to get latest request", "error", err)
	}
	c := latest.Cookies
	h := latest.Headers
	refererO := utils.TryGet[string](h, "Referer", "referer")
	outDir, err := GetOutDirFromViper()
	if err != nil {
		log.Sugar().Panicw("failed to get output directory", "error", err)
	}

	// https://req.cool/zh/docs/tutorial/http-fingerprint/
	// https://req.cool/zh/docs/tutorial/tls-fingerprint/
	// [ImpersonateChrome] would also set TLS fingerprint
	// https://req.cool/zh/docs/tutorial/proxy/
	client := req.C().ImpersonateChrome()
	_, f, err := GetHttpProxyFromViper()
	if err != nil {
		log.Sugar().Errorw("failed to get proxy", "error", err)
	} else {
		client = client.SetProxy(f)
	}

	referer, ok := refererO.Get()
	if ok {
		log.Sugar().Debugw("using referer", "referer", referer)
	} else {
		log.Sugar().Warnf("no referer set")
	}

	sz, err := GetPoolSizeFromViper()
	if err != nil {
		log.Sugar().Panicw("failed to get pool size", "error", err)
	}
	p, err := ants.NewPool(sz)
	if err != nil {
		log.Sugar().Panicw("failed to create pool", "error", err)
	}
	defer p.Release()
	var wg sync.WaitGroup
	for _, link := range d.Links {
		wg.Add(1)
		dlFn := func() {
			// get the last part of the path
			p := path.Base(link.Path)
			out := path.Join(outDir, p)
			stat, err := os.Stat(out)
			if !os.IsNotExist(err) {
				if stat.IsDir() {
					log.Sugar().Errorw("output file is a directory. skip.", "url", link.String(), "output", out)
					return
				}
				log.Sugar().Infow("output file already exists. skip.", "url", link.String(), "output", out)
				return
			}
			// convert to array of pointers...
			cookies := utils.Map(c, func(c http.Cookie) *http.Cookie { return &c })
			R := client.R().SetCookies(cookies...)
			referer, ok := refererO.Get()
			if ok {
				R.SetHeader("Referer", referer)
			}
			R.SetHeader("Sec-Fetch-Dest", "image")
			R.SetHeader("Sec-Fetch-Mode", "no-cors")
			R.SetHeader("Sec-Fetch-Site", "same-site")
			res, err := R.Get(link.String())
			if err != nil {
				log.Sugar().Errorw("failed to download image", "url", link.String(), "error", err)
				utils.PrintHeadersCookies(R)
				return
			}

			// TODO: custom content type. For now I only need image
			if ct := res.Header.Get("Content-Type"); !strings.Contains(ct, "image") {
				log.Sugar().Errorw("not a image", "url", link.String(), "Content-Type", res.Header.Get("Content-Type"))
				log.Sugar().Debugw("response", "headers", res.Header, "response", res.String())
				utils.PrintHeadersCookies(R)
				// TODO: retry and max Error to break
				return
			}
			err = os.WriteFile(out, res.Bytes(), 0644)
			if err != nil {
				log.Sugar().Errorw("failed to save image", "url", link.String(), "error", err)
				return
			} else {
				log.Sugar().Infow("downloaded", "url", link.String(), "output", out)
			}
		}
		err = p.Submit(func() {
			dlFn()
			wg.Done()
		})
	}
	wg.Wait()
}

var from = cobra.Command{
	Use:   "from",
	Short: "download from a entity file",
	Args:  cobra.ExactArgs(1),
	Run:   runDescription,
}
