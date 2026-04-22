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
	apiv1 "github.com/andrieee44/countmein/api/v1"
	apiv2 "github.com/andrieee44/countmein/api/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/cors"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var (
	logger        *slog.Logger
	db            *sql.DB
	rpcHandlersv1 []apiv2.RPCHandlerFn
	rpcHandlersv2 []apiv2.RPCHandlerFn
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

func handlerv1tov2(path string, handler http.Handler) apiv2.RPCHandlerFn {
	return func() (string, http.Handler) {
		return path, handler
	}
}

func initAPI() {
	var (
		sessionStore *apiv2.UserSessionService
		optsv1       []connect.HandlerOption
		optsv2       []connect.HandlerOption
	)

	sessionStore = apiv2.NewUserSessionService(db)

	optsv1 = []connect.HandlerOption{
		connect.WithInterceptors(
			apiv1.NewErrorInterceptor(logger),
			apiv1.NewAuthInterceptor(db),
		),
	}

	rpcHandlersv1 = []apiv2.RPCHandlerFn{
		handlerv1tov2(apiv1.NewUserHandler(db, optsv1...)),
		handlerv1tov2(apiv1.NewCalendarHandler(db, optsv1...)),
		handlerv1tov2(apiv1.NewAIHandler(logger, optsv1...)),
	}

	optsv2 = []connect.HandlerOption{
		connect.WithInterceptors(
			apiv2.ErrorInterceptor(logger),
			sessionStore.AuthInterceptor(),
		),
	}

	rpcHandlersv2 = []apiv2.RPCHandlerFn{
		apiv2.UserHandler(db, sessionStore, optsv2...),
		apiv2.UserSessionHandler(db, optsv2...),
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
		mux        *http.ServeMux
		rpcHandler apiv2.RPCHandlerFn
		handler    http.Handler
		err        error
	)

	mux = http.NewServeMux()

	mux.Handle("/", http.StripPrefix(
		"/",
		http.FileServer(http.Dir(os.Getenv("PUBLIC_PATH"))),
	))

	for _, rpcHandler = range rpcHandlersv1 {
		mux.Handle(rpcHandler())
	}

	for _, rpcHandler = range rpcHandlersv2 {
		mux.Handle(rpcHandler())
	}

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
