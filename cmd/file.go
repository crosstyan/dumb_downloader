package cmd

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"

	"github.com/crosstyan/dumb_downloader/description"
	"github.com/crosstyan/dumb_downloader/log"
	"github.com/crosstyan/dumb_downloader/utils"
	"github.com/imroc/req/v3"
	"github.com/joomcode/errorx"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"
)

type ProxyFunc = func(*http.Request) (*url.URL, error)

func GetDescription(descriptionPath string) (*description.Description, error) {
	if _, err := os.Stat(descriptionPath); os.IsNotExist(err) {
		return nil, errorx.Decorate(err, "config file %s not found", descriptionPath)
	}
	content, err := os.ReadFile(descriptionPath)
	if err != nil {
		return nil, errorx.Decorate(err, "failed to read config file %s", descriptionPath)
	}
	d := description.Description{}
	err = json.Unmarshal(content, &d)
	if err != nil {
		return nil, errorx.Decorate(err, "failed to parse config file %s", descriptionPath)
	}
	return &d, nil
}

func GetLatestRequest(description *description.Description) (*description.RequestRecord, error) {
	// https://github.com/panjf2000/ants
	reqs := maps.Values(description.Requests)
	if len(reqs) == 0 {
		return nil, errorx.IllegalArgument.New("no requests found in description")
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
		log.Sugar().Panicw("failed to get description", "error", err)
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

	for _, v := range d.Links {
		// get the last part of the path
		p := path.Base(v.Path)
		out := path.Join(outDir, p)
		stat, err := os.Stat(out)
		if !os.IsNotExist(err) {
			if stat.IsDir() {
				log.Sugar().Errorw("output file is a directory. skip.", "url", v.String(), "output", out)
				continue
			}
			log.Sugar().Infow("output file already exists. skip.", "url", v.String(), "output", out)
			continue
		}
		// convert to array of pointers...
		cookies := utils.Map(c, func(c http.Cookie) *http.Cookie { return &c })
		r := client.R().SetCookies(cookies...)
		referer, ok := refererO.Get()
		if ok {
			r.SetHeader("Referer", referer)
		}
		r.SetHeader("Sec-Fetch-Dest", "image")
		r.SetHeader("Sec-Fetch-Mode", "no-cors")
		r.SetHeader("Sec-Fetch-Site", "same-site")
		res, err := r.Get(v.String())
		printHeadersCookies := func() {
			log.Sugar().Debugw("request headers", "headers", res.Request.Headers)
			cs := res.Request.Cookies
			for _, c := range cs {
				log.Sugar().Debugw("cookie", "name", c.Name, "value", c.Value)
			}
		}
		if err != nil {
			log.Sugar().Errorw("failed to download image", "url", v.String(), "error", err)
			printHeadersCookies()
			continue
		}
		if res.Header.Get("Content-Type") != "image/jpeg" {
			log.Sugar().Errorw("not a jpeg image", "url", v.String(), "content-type", res.Header.Get("Content-Type"))
			log.Sugar().Debugw("response", "headers", res.Header, "response", res.String())
			printHeadersCookies()
			// TODO: retry and max Error to break
			continue
		}
		err = os.WriteFile(out, res.Bytes(), 0644)
		if err != nil {
			log.Sugar().Errorw("failed to save image", "url", v.String(), "error", err)
			continue
		} else {
			log.Sugar().Infow("downloaded", "url", v.String(), "output", out)
		}
	}
}

var from = cobra.Command{
	Use:   "from",
	Short: "download from a description file",
	Args:  cobra.ExactArgs(1),
	Run:   runDescription,
}
