package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"connectrpc.com/connect"
	connectcors "connectrpc.com/cors"
	"github.com/andrieee44/countmein/api"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/cors"
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

func withCORS(h http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins: strings.Split(os.Getenv("HOSTS"), ","),
		AllowedMethods: connectcors.AllowedMethods(),
		AllowedHeaders: append(connectcors.AllowedHeaders(), "Authorization"),
		ExposedHeaders: connectcors.ExposedHeaders(),
	}).Handler(h)
}

func main() {
	var (
		mux     *http.ServeMux
		handler http.Handler
		err     error
	)

	mux = http.NewServeMux()
	mux.Handle("/", http.StripPrefix(
		"/",
		http.FileServer(http.Dir(os.Getenv("PUBLIC_PATH"))),
	))

	mux.Handle(api.NewUserHandler(db, opts...))
	mux.Handle(api.NewCalendarHandler(db, opts...))
	handler = withCORS(mux)

	logger.Info(
		"Server Starting",
		"HOSTS", os.Getenv("HOSTS"),
		"PORT", os.Getenv("PORT"),
		"PUBLIC_PATH", os.Getenv("PUBLIC_PATH"),
	)

	err = http.ListenAndServe(
		fmt.Sprintf(":%s", os.Getenv("PORT")),
		h2c.NewHandler(handler, &http2.Server{}),
	)
	if err != nil {
		panic(err)
	}
}
