package cmd

import (
	"os"

	"github.com/crosstyan/dumb_downloader/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

func init() {
	root.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")

	root.PersistentFlags().StringP("http-proxy", "P", "", "HTTP proxy")
	viper.BindPFlag("http_proxy", root.PersistentFlags().Lookup("http-proxy"))

	viper.SetEnvPrefix("DUMB")
	viper.AutomaticEnv()
	viper.BindEnv("http_proxy", "http_proxy", "HTTP_PROXY", "https_proxy", "HTTPS_PROXY")

	from.PersistentFlags().StringP("output-dir", "o", "out", "output directory")
	viper.BindPFlag("output_dir", root.PersistentFlags().Lookup("output-dir"))

	serve.PersistentFlags().StringP("listen", "l", "127.0.0.1:8888", "listen address")
	viper.BindPFlag("listen", root.PersistentFlags().Lookup("listen"))
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
