package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var config Config

var RootCmd = &cobra.Command{
	Use:   "",
	Short: "A simple reverse proxy",
	Run:   func(cmd *cobra.Command, args []string) {},
}

func init() {
	RootCmd.PersistentFlags().String("config", "", "config file (default is ./balance.yml)")
	viper.BindPFlag("config", RootCmd.PersistentFlags().Lookup("config"))
	viper.AutomaticEnv()
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	configFile := viper.GetString("config")

	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.AddConfigPath("./")
		viper.SetConfigName("balance")
		viper.SetConfigType("yml")
	}

	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config file: %v\n", err)
		os.Exit(1)
	}

	if err := viper.Unmarshal(&config); err != nil {
		fmt.Fprintf(os.Stderr, "Error unmarshalling config: %v\n", err)
		os.Exit(1)
	}
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		panic(err)
	}

	proxyHandler := NewProxyHandler(config.BackendConfigs)
	defer proxyHandler.Cleanup()

	balancingAlgorithm := getBalancingAlgorithm(proxyHandler)

	log.Fatal(http.ListenAndServe(config.Port, balancingAlgorithm()))
}
