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
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/Studiously/usersvc/ddl"
	"github.com/Studiously/usersvc/service"
	"github.com/go-kit/kit/log"
	"github.com/ory/common/env"
	"github.com/ory/hydra/sdk"
	"github.com/rubenv/sql-migrate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var addr string
var dry bool = true

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "usersvc",
	Short: "Create a new instance of the user service.",
	Long: `usersvc is a microservice for user management and local authentication for Studiously.

	While usersvc handles migrations, it is recommended that the database be manually inspected to ensure proper functionality.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		var logger log.Logger
		{
			logger = log.NewLogfmtLogger(logrus.StandardLogger().Out)
			logger = log.With(logger, "ts", log.DefaultTimestampUTC)
			logger = log.With(logger, "caller", log.DefaultCaller)
		}
		// Set up Hydra
		client, err := sdk.Connect(
			sdk.ClientID(env.Getenv("HYDRA_CLIENT_ID", "consent")),
			sdk.ClientSecret(env.Getenv("HYDRA_CLIENT_SECRET", "demovo")),
			sdk.ClusterURL(env.Getenv("HYDRA_CLUSTER_URL", "http://localhost:4444")),
		)
		if err != nil {
			logrus.WithError(err).Fatal("could not connect to Hydra")
		}
		var s service.Service
		{
			switch dry {
			case false:
				// Set up database
				var driver = os.Getenv("DATABASE_DRIVER")
				var config = os.Getenv("DATABASE_CONFIG")

				db, err := sql.Open(driver, config)
				if err != nil {
					logrus.Fatalln("database connection failed", err)
				}
				if err := pingDatabase(db); err != nil {
					logrus.Errorln(err)
					logrus.Fatalln("database ping attempts failed")
				}
				if err := setupDatabase(driver, db); err != nil {
					logrus.Errorln(err)
					logrus.Fatalln("migration failed")
				}
				s = service.NewPersistentService(db, client)
			case true:
				s = service.NewInmemService()
			}
		}
		var h = service.MakeHTTPHandler(s, client, logger)

		errs := make(chan error)
		go func() {
			c := make(chan os.Signal, 100)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			errs <- fmt.Errorf("%s", <-c)
		}()

		go func() {
			logger.Log("transport", "HTTP", "addr", addr)
			errs <- http.ListenAndServe(addr, h)
		}()

		logger.Log("exit", <-errs)
	},
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

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.usersvc.yaml)")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	RootCmd.PersistentFlags().StringVarP(&addr, "addr", "a", ":8080", "HTTP bind address")
	RootCmd.PersistentFlags().BoolVarP(&dry, "dry", "d", true, "Uses an in-memory testing store for a dry run.")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".usersvc") // name of config file (without extension)
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME") // adding home directory as first search path
	viper.AutomaticEnv()         // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func setupDatabase(driver string, db *sql.DB) error {
	var migrations = &migrate.AssetMigrationSource{
		Asset:    ddl.Asset,
		AssetDir: ddl.AssetDir,
		Dir:      driver,
	}
	_, err := migrate.Exec(db, driver, migrations, migrate.Up)
	return err
}

func pingDatabase(db *sql.DB) (err error) {
	for i := 0; i < 30; i++ {
		err = db.Ping()
		if err == nil {
			return
		}
		logrus.Infof("database ping failed. retry in 1s")
		time.Sleep(time.Second)
	}
	return
}
