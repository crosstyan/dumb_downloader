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
	"github.com/kr/pretty"
	"github.com/samber/mo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/maps"
)

var root = cobra.Command{
	Use:   "dumbdl",
	Short: "a dumb downloader that would download with a certain structure",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := args[0]
		httpProxy := viper.GetString("http_proxy")
		type ProxyFunc = func(*http.Request) (*url.URL, error)
		var maybe_proxy mo.Option[ProxyFunc]
		if httpProxy != "" {
			log.Sugar().Infof("using %s as http/https proxy", httpProxy)
			u, err := url.Parse(httpProxy)
			if err != nil {
				log.Sugar().Errorf("failed to parse http/https proxy: %s", httpProxy)
				return
			}
			maybe_proxy = mo.Some(http.ProxyURL(u))
		} else {
			log.Sugar().Infof("no http/https proxy")
		}
		outDir := viper.GetString("output_dir")
		if outDir == "" {
			log.Sugar().Errorf("no output directory")
			return
		}
		log.Sugar().Infof("using %s as output directory", outDir)
		info, err := os.Stat(outDir)
		if os.IsNotExist(err) {
			log.Sugar().Infof("creating output directory %s", outDir)
			err = os.MkdirAll(outDir, 0755)
			if err != nil {
				log.Sugar().Errorf("failed to create output directory %s", outDir)
				return
			}
		} else if !info.IsDir() {
			log.Sugar().Errorf("output directory %s is not a directory", outDir)
			return
		}
		log.Sugar().Infof("try to read config file: %s", target)
		if _, err := os.Stat(target); os.IsNotExist(err) {
			log.Sugar().Errorf("config file %s not found", target)
			return
		}
		content, err := os.ReadFile(target)
		if err != nil {
			log.Sugar().Errorf("failed to read config file %s", target)
			return
		}
		d := description.Description{}
		err = json.Unmarshal(content, &d)
		if err != nil {
			log.Sugar().Errorf("failed to parse config file %s", target)
			return
		}
		// log.Sugar().Debugf("config file content: %s", pretty.Sprint(d))
		// https://github.com/panjf2000/ants
		reqs := maps.Values(d.Requests)
		if len(reqs) == 0 {
			log.Sugar().Errorf("no request found in config file %s", target)
			return
		}
		sort.Slice(reqs, func(i, j int) bool {
			return reqs[i].Timestamp.After(reqs[j].Timestamp)
		})
		latest := reqs[0]
		c := latest.Cookies
		h := latest.Headers
		maybe_referer := utils.TryGet[string](h, "Referer", "referer")
		// https://req.cool/zh/docs/tutorial/http-fingerprint/
		// https://req.cool/zh/docs/tutorial/tls-fingerprint/
		// [ImpersonateChrome] would also set TLS fingerprint
		// https://req.cool/zh/docs/tutorial/proxy/

		client := req.C().ImpersonateChrome()
		proxy, ok := maybe_proxy.Get()
		if ok {
			client = client.SetProxy(proxy)
		}

		for _, v := range d.Images {
			if err != nil {
				log.Sugar().Errorf("failed to create request for %s", v.String())
				return
			}
			// convert to array of pointers...
			cookies := utils.Map(c, func(c http.Cookie) *http.Cookie { return &c })
			r := client.R().SetCookies(cookies...)
			referer, ok := maybe_referer.Get()
			if ok {
				log.Sugar().Debugf("using referer %s", referer)
				r.Headers.Add("Referer", referer)
			} else {
				log.Sugar().Warnf("no referer found")
			}
			res, err := r.Get(v.String())
			printHeadersCookies := func() {
				log.Sugar().Debugf("request headers: %s", pretty.Sprint(res.Request.Headers))
				cs := res.Request.Cookies
				for _, c := range cs {
					log.Sugar().Debugf("%s: %s", c.Name, c.Value)
				}
			}
			if err != nil {
				log.Sugar().Errorf("failed to download %s", v.String())
				printHeadersCookies()
				return
				// continue
			}
			if res.Header.Get("Content-Type") != "image/jpeg" {
				log.Sugar().Errorf("not a jpeg image: %s", v.String())
				log.Sugar().Errorf("response: %s", res.String())
				printHeadersCookies()
				return
				// continue
			}
			log.Sugar().Infof("downloaded %s", v.String())
			p := path.Base(v.Path)
			out := path.Join(outDir, p)
			err = os.WriteFile(out, res.Bytes(), 0644)
			if err != nil {
				log.Sugar().Errorf("failed to write file %s", out)
				continue
			} else {
				log.Sugar().Infof("saved to %s", out)
			}
		}
	},
}

var cfgFile string

func Execute() error {
	return root.Execute()
}

func init() {
	root.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")
	root.PersistentFlags().StringP("http-proxy", "P", "", "HTTP proxy")
	viper.BindPFlag("http_proxy", root.PersistentFlags().Lookup("http-proxy"))
	root.PersistentFlags().StringP("output-dir", "o", "out", "output directory")
	viper.BindPFlag("output_dir", root.PersistentFlags().Lookup("output-dir"))
	viper.SetEnvPrefix("dumb")
	viper.AutomaticEnv()
	viper.BindEnv("http_proxy", "http_proxy", "HTTP_PROXY", "https_proxy", "HTTPS_PROXY")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("toml")
		viper.SetConfigName("dumb.toml")
	}
	if err := viper.ReadInConfig(); err == nil {
		log.Sugar().Infof("Using config file: %s", viper.ConfigFileUsed())
	}
}
