package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/Studiously/usersvc"
	"github.com/Studiously/usersvc/ddl"
	"github.com/go-kit/kit/log"
	_"github.com/lib/pq"
	"github.com/rubenv/sql-migrate"
)

const (
	DatabaseKey = "database"
)

func main() {

	var (
		httpAddr = flag.String("http.addr", ":8080", "HTTP listen address")
		mode     = flag.String("mode", "inmem", "Storage mode (inmem or persist)")
	)
	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var s usersvc.Service
	{
		switch (*mode) {

		case "persist":
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

			s = usersvc.NewPersistentService(db)
		case "inmem":
			fallthrough
		default:
			s = usersvc.NewInmemService()
			break
		}
	}

	var h http.Handler
	{
		h = usersvc.MakeHTTPHandler(s, logger)
	}

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	go func() {
		logger.Log("transport", "HTTP", "addr", *httpAddr)
		errs <- http.ListenAndServe(*httpAddr, h)
	}()

	logger.Log("exit", <-errs)
}

// helper function to setup the database by performing
// automated database migration steps.
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

func db(c context.Context) *sql.DB {
	return c.Value(DatabaseKey).(*sql.DB)
}
