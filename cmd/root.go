// Copyright © 2017 Meyer Zinn <meyerzinn@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "usersvc",
	Short: "User service for Studiously.",
	Long:  `usersvc is an identity provider microservice for user management and local authentication in Studiously.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//Run: func(cmd *cobra.Command, args []string) {
	//var logger log.Logger
	//{
	//	logger = log.NewLogfmtLogger(logrus.StandardLogger().Out)
	//	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	//	logger = log.With(logger, "caller", log.DefaultCaller)
	//}
	//// Set up Hydra
	//client, err := sdk.Connect(
	//	sdk.ClientID(os.Getenv("HYDRA_CLIENT_ID")),
	//	sdk.ClientSecret(os.Getenv("HYDRA_CLIENT_SECRET")),
	//	sdk.ClusterURL(os.Getenv("HYDRA_CLUSTER_URL")),
	//)
	//if err != nil {
	//	logrus.WithError(err).Fatal("could not connect to Hydra")
	//}
	//var s usersvc.Service
	//{
	//	// Set up database
	//	var driver = os.Getenv("DATABASE_DRIVER")
	//	var config = os.Getenv("DATABASE_CONFIG")
	//
	//	db, err := sql.Open(driver, config)
	//	if err != nil {
	//		logrus.Fatalln("database connection failed", err)
	//	}
	//	if err := pingDatabase(db); err != nil {
	//		logrus.Errorln(err)
	//		logrus.Fatalln("database ping attempts failed")
	//	}
	//	if err := setupDatabase(driver, db); err != nil {
	//		logrus.Errorln(err)
	//		logrus.Fatalln("migration failed")
	//	}
	//	s = usersvc.New(db, client)
	//}
	//var h = usersvc.MakeHTTPHandler(s, client, logger)
	//
	//errs := make(chan error)
	//go func() {
	//	c := make(chan os.Signal, 100)
	//	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	//	errs <- fmt.Errorf("%s", <-c)
	//}()
	//
	//go func() {
	//	logger.Log("transport", "HTTP", "addr", addr)
	//	errs <- http.ListenAndServe(addr, h)
	//}()
	//
	//logger.Log("exit", <-errs)
	//},
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.unitsvc.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".usersvc") // name of config file (without extension)
	viper.AddConfigPath("$HOME")    // adding home directory as first search path
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
