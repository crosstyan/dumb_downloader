package cmd

import (
	"github.com/crosstyan/dumb_downloader/global/log"
	"net/http"
	"net/url"
	"os"

	"github.com/joomcode/errorx"
	"github.com/spf13/viper"
)

var NotFound = errorx.CommonErrors.NewType("not_found", errorx.NotFound())

func GetListenAddrFromViper() (string, error) {
	listenAddr := viper.GetString("listen")
	if listenAddr == "" {
		return "", NotFound.New("no listen address")
	}
	return listenAddr, nil
}

func GetHttpProxyFromViper() (*url.URL, ProxyFunc, error) {
	httpProxy := viper.GetString("http_proxy")
	if httpProxy != "" {
		log.Sugar().Infof("using %s as http/https proxy", httpProxy)
		u, err := url.Parse(httpProxy)
		if err != nil {
			return nil, nil, err
		}
		f := http.ProxyURL(u)
		return u, f, nil
	} else {
		return nil, nil, NotFound.New("no http/https proxy found")
	}
}

func GetOutDirFromViper() (string, error) {
	outDir := viper.GetString("output_dir")
	if outDir == "" {
		return "", NotFound.New("no output directory")
	}
	info, err := os.Stat(outDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(outDir, 0755)
		if err != nil {
			return "", errorx.Decorate(err, "failed to create output directory %s", outDir)
		}
	} else if !info.IsDir() {
		return "", errorx.IllegalArgument.New("output directory %s is not a directory", outDir)
	}
	return outDir, nil
}

func GetPoolSizeFromViper() (int, error) {
	poolSize := viper.GetInt("pool_size")
	if poolSize <= 0 {
		return 0, errorx.IllegalArgument.New("pool size should be positive")
	}
	return poolSize, nil
}
