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

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	"github.com/ory/hydra/sdk"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rubenv/sql-migrate"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/studiously/classsvc/classsvc"
	"github.com/studiously/usersvc/ddl"
	"github.com/studiously/usersvc/usersvc"
)

var (
	addr      string
	debugAddr string
)

// hostCmd represents the host command
var hostCmd = &cobra.Command{
	Use:   "host",
	Short: "Start the service.",
	Long: `Starts the service on all transports and connects to a database backend.

This command exposes several environmental variables for controls. You can set environments using "export KEY=VALUE" (Linux/macOS) or "set KEY=VALUE" (Windows). On Linux, you can also set environments by prepending key value pairs: "KEY=VALUE KEY2=VALUE2 usersvc"

Core Controls
=============
- DATABASE_DRIVER: The driver to use with the database. Only 'postgres' is currently supported.
- DATABASE_CONFIG: A URL to a persistent backend.
- CLASSSVC_URL: A URL to an instance of classsvc.

Hydra Controls
==============
A Hydra server is required. Most endpoints (excepting health and unauthenticated ones) will fail without a valid Hydra server.
- HYDRA_CLIENT_ID: ID for Hydra client.
- HYDRA_CLIENT_SECRET: Secret for Hydra client.
- HYDRA_CLUSTER_URL: URL of Hydra cluster.
- HYDRA_TLS_VERIFY: Whether the client should verify Hydra's TLS.

Messaging Controls
==================
A NATS cluster is required for messaging across services. Without it, stale data pertaining to deleted resources may remain in the database, merely becoming inaccessible.
- NATS_CLUSTER_URL: URL of NATS cluster.
`,
	Run: func(cmd *cobra.Command, args []string) {
		// Set up logger
		var logger log.Logger
		{
			logger = log.NewLogfmtLogger(os.Stdout)
			logger = log.With(logger, "ts", log.DefaultTimestampUTC)
			logger = log.With(logger, "caller", log.DefaultCaller)
		}

		// Set up metrics
		var requestCount metrics.Counter
		{
			requestCount = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
				Namespace: "unitsvc",
				Name:      "request_count",
				Help:      "Total count of requests to all endpoints.",
			}, []string{})
		}
		var requestLatency metrics.Histogram
		{
			// Transport level metrics.
			requestLatency = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
				Namespace: "unitsvc",
				Name:      "request_duration_ns",
				Help:      "Request duration in nanoseconds.",
			}, []string{"method", "success"})
		}

		// Connect to Hydra
		var client *sdk.Client
		{
			cli, err := sdk.Connect(
				sdk.ClientID(viper.GetString("hydra.client.id")),
				sdk.ClientSecret(viper.GetString("hydra.client.secret")),
				sdk.ClusterURL(viper.GetString("hydra.cluster_url")),
				sdk.SkipTLSVerify(viper.GetBool("hydra.tls_verify")),
				sdk.Scopes("hydra"),
			)
			if err != nil {
				logger.Log("msg", "could not connect to Hydra cluster", "error", err, "cluster_url", viper.GetString("hydra.cluster_url"))
				os.Exit(-1)
			}
			client = cli
		}

		// Set up database
		var db *sql.DB
		{
			var driver = viper.GetString("database.driver")
			var config = viper.GetString("database.config")

			db, err := sql.Open(driver, config)
			if err != nil {
				logger.Log("msg", "database connection failed", "error", err)
				os.Exit(-1)
			}
			if err := pingDatabase(db); err != nil {
				logger.Log("msg", "database unresponsive")
				os.Exit(-1)
			}
			if err := setupDatabase(driver, db); err != nil {
				logger.Log("msg", "database migrations failed", "error", err)
				os.Exit(-1)
			}
		}

		var cs classsvc.Service
		{
			var err error
			cs, err = classsvc.MakeClientEndpoints(viper.GetString("classsvc.addr"))
			if err != nil {
				logger.Log("msg", "could not make classsvc client", "error", err)
				os.Exit(-1)
			}
		}

		// Initialize service and middleware
		var service usersvc.Service
		{
			service = usersvc.New(db, cs)
			service = usersvc.LoggingMiddleware(logger)(service)
			service = usersvc.InstrumentingMiddleware(requestCount, requestLatency)(service)
		}

		errs := make(chan error)

		// Start debug server
		go func() {
			logger := log.With(logger, "transport", "debug")
			m := http.NewServeMux()
			m.Handle("/metrics", promhttp.Handler())
			logger.Log("addr", debugAddr)
			errs <- http.ListenAndServe(debugAddr, m)
		}()

		// Handle keyboard interrupts
		go func() {
			c := make(chan os.Signal)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			errs <- fmt.Errorf("%s", <-c)
		}()

		// Start HTTP server for main service
		var h = usersvc.MakeHTTPHandler(service, client, logger)
		go func(address string) {
			logger.Log("transport", "HTTP", "addr", addr)
			errs <- http.ListenAndServe(address, h)
		}(addr)

	},
}

func init() {
	RootCmd.AddCommand(hostCmd)

	hostCmd.Flags().StringVarP(&addr, "addr", "a", ":8080", "HTTP listen address")
	hostCmd.Flags().StringVarP(&debugAddr, "debug-addr", "d", ":8081", "Debug and metrics listen address")

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
		time.Sleep(time.Second)
	}
	return
}
