package cmd

import (
	"github.com/crosstyan/dumb_downloader/global/log"
	"github.com/spf13/pflag"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	ListenFlagName    = "listen"
	HttpProxyFlagName = "http_proxy"
	PoolSizeFlagName  = "pool_size"
	OutputDirFlagName = "output_dir"
)

var root = cobra.Command{
	Use:   "dumbdl",
	Short: "a dumb downloader that would download with a certain structure",
}

var cfgFile string

func Execute() error {
	root.AddCommand(&serve, &from)
	return root.Execute()
}

// https://github.com/spf13/viper/discussions/1054
// https://github.com/spf13/cobra/blob/95d8a1e45d7719c56dc017e075d3e6099deba85d/command_test.go#L1645-L1652
func init() {
	cobra.OnInitialize(initConfig)
	root.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")
	normFnNew := func(f *pflag.FlagSet, name string) pflag.NormalizedName {
		return pflag.NormalizedName(strings.ReplaceAll(name, "_", "-"))
	}
	root.SetGlobalNormalizationFunc(normFnNew)
	root.PersistentFlags().StringP(HttpProxyFlagName, "P", "", "HTTP proxy")
	err := viper.BindPFlag(HttpProxyFlagName, root.PersistentFlags().Lookup(HttpProxyFlagName))
	if err != nil {
		log.Sugar().Panicw("failed to bind flag", "flag", HttpProxyFlagName, "error", err)
	}

	viper.SetEnvPrefix("DUMB")
	viper.AutomaticEnv()
	err = viper.BindEnv(HttpProxyFlagName, "http_proxy", "HTTP_PROXY", "https_proxy", "HTTPS_PROXY")
	if err != nil {
		log.Sugar().Panicw("failed to bind env", "env", HttpProxyFlagName, "error", err)
	}

	root.PersistentFlags().IntP(PoolSizeFlagName, "p", 16, "pool size")
	err = viper.BindPFlag(PoolSizeFlagName, root.PersistentFlags().Lookup(PoolSizeFlagName))
	if err != nil {
		log.Sugar().Panicw("failed to bind flag", "flag", PoolSizeFlagName, "error", err)
	}

	root.PersistentFlags().StringP(OutputDirFlagName, "o", "out", "output directory")
	err = viper.BindPFlag(OutputDirFlagName, root.PersistentFlags().Lookup(OutputDirFlagName))
	if err != nil {
		log.Sugar().Panicw("failed to bind flag", "flag", OutputDirFlagName, "error", err)
	}

	serve.PersistentFlags().StringP(ListenFlagName, "l", "127.0.0.1:8888", "listen address")
	err = viper.BindPFlag(ListenFlagName, serve.PersistentFlags().Lookup(ListenFlagName))
	if err != nil {
		log.Sugar().Panicw("failed to bind flag", "flag", ListenFlagName, "error", err)
	}
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
		log.Sugar().Infow("using config file", "file", viper.ConfigFileUsed())
	} else {
		log.Sugar().Infof("Not using config file because %s", err.Error())
	}
}
