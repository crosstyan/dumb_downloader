package cmd

import (
	"encoding/json"
	"os"

	"github.com/crosstyan/dumb_downloader/description"
	"github.com/crosstyan/dumb_downloader/log"
	"github.com/kr/pretty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var root = cobra.Command{
	Use:   "dumbdl",
	Short: "a dumb downloader that would download with a certain structure",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := args[0]
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
		log.Sugar().Infof("config file content: %s", pretty.Sprint(d))
	},
}

var cfgFile string

func Execute() error {
	return root.Execute()
}

func init() {
	root.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
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
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		log.Sugar().Infof("Using config file: %s", viper.ConfigFileUsed())
	}
}
