package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"connectrpc.com/connect"
	"github.com/andrieee44/countmein/api"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var (
	logger *slog.Logger
	db     *sql.DB
	opts   []connect.HandlerOption
)

func init() {
	logger = slog.Default()

	initDB()
	initAPI()
}

func initDB() {
	var err error

	db, err = sql.Open(
		"mysql",
		fmt.Sprintf(
			"%s:%s@tcp(%s:3306)/%s?parseTime=true",
			os.Getenv("DB_USERNAME"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_HOST"),
			os.Getenv("DB_NAME"),
		),
	)
	if err != nil {
		panic(err)
	}

	logger.Info(
		"Pinging Database",
		"DB_HOST", os.Getenv("DB_HOST"),
		"DB_NAME", os.Getenv("DB_NAME"),
	)

	err = db.Ping()
	if err != nil {
		panic(err)
	}
}

func initAPI() {
	opts = []connect.HandlerOption{
		connect.WithInterceptors(
			api.NewErrorInterceptor(logger),
			api.NewAuthInterceptor(db),
		),
	}
}

func main() {
	var (
		mux *http.ServeMux
		err error
	)

	mux = http.NewServeMux()
	mux.Handle(api.NewUserHandler(db, opts...))

	logger.Info(
		"Server Starting",
		"host", "localhost",
		"port", os.Getenv("PORT"),
	)

	err = http.ListenAndServe(
		fmt.Sprintf(":%s", os.Getenv("PORT")),
		h2c.NewHandler(mux, &http2.Server{}),
	)
	if err != nil {
		panic(err)
	}
}
