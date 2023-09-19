package cmd

import (
	"fmt"
	"github.com/bearboy/jetson_prometheus_exporter/build"
	"github.com/bearboy/jetson_prometheus_exporter/exporter"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"runtime"
)

var (
	cfgFile string
)

var rootCmd = &cobra.Command{
	Use:   "jetson_exporter",
	Short: "Prometheus exporter for detailed jetson info",
	Long:  `Prometheus exporter analysing jetson gpu info`,
	Run: func(cmd *cobra.Command, args []string) {
		printHeader()
		interval := viper.GetInt("tegrastats-interval")
		filePath := viper.GetString("tegrastats-log-file")
		cleanFileInterval := viper.GetInt("logfile-cleanup-interval-hours")
		//启动tegrastats
		tegrastats := exporter.Tegrastats{Interval: interval, LogPath: filePath}
		tegrastats.Start(interval, filePath)
		e := exporter.NewExporter(
			interval,
			filePath,
			cleanFileInterval,
			&tegrastats,
		)
		e.InitPrometheus()
		e.RunServer(viper.GetString("jetson-bind-address"))
	},
}

// Execute runs the command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
func init() {
	cobra.OnInitialize(initConfig)
	flags := rootCmd.PersistentFlags()
	pwd, _ := os.Getwd()
	flags.StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.jetson_exporter.yaml)")
	flags.StringP("jetson-bind-address", "b", "0.0.0.0:9995", "Address to bind to")
	flags.StringP("tegrastats-log-file", "p", pwd, "Dumps the output of tegrastats to <filename>.")
	flags.IntP("tegrastats-interval", "i", 1000, "Samples the information in <milliseconds>")
	flags.IntP("logfile-cleanup-interval-hours", "l", 1, "After how many hours we want to clean up tegrastats_logfile(argument above).")
	viper.BindPFlags(flags)

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		// Search config in home directory with name ".jetson_exporter" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".jetson_exporter")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func printHeader() {
	log.Printf("Jetson Prometheus Exporter %s build date: %s	 Go: %s	GOOS: %s	GOARCH: %s",
		build.BuildVersion,
		build.BuildDate,
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
	)
}
