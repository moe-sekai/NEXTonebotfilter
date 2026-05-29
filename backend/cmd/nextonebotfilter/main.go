package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/exmeaning/nextonebotfilter/internal/filter"
	"github.com/exmeaning/nextonebotfilter/internal/server"
	"github.com/exmeaning/nextonebotfilter/internal/store"
)

func main() {
	var (
		dbPath      = flag.String("db", "data/nextonebotfilter.db", "path to sqlite database")
		consoleAddr = flag.String("console", ":8787", "console (HTTP API + UI) listen address")
		webDir      = flag.String("web", "", "directory to serve as static console UI (use Next.js export output)")
		logPath     = flag.String("log", "", "also write logs to this file (in addition to stderr)")
		debug       = flag.Bool("debug", false, "enable debug logging")
	)
	flag.Parse()

	zerolog.TimeFieldFormat = time.RFC3339
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	var writers []io.Writer
	writers = append(writers, zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	if *logPath != "" {
		if err := os.MkdirAll(filepath.Dir(*logPath), 0o755); err != nil {
			fmt.Fprintln(os.Stderr, "create log dir:", err)
		}
		f, err := os.OpenFile(*logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			fmt.Fprintln(os.Stderr, "open log file:", err)
		} else {
			writers = append(writers, zerolog.ConsoleWriter{Out: f, NoColor: true, TimeFormat: time.RFC3339})
		}
	}
	log.Logger = log.Output(zerolog.MultiLevelWriter(writers...))

	if err := os.MkdirAll(filepath.Dir(*dbPath), 0o755); err != nil {
		log.Fatal().Err(err).Msg("create data dir")
	}
	db, err := store.Open(*dbPath)
	if err != nil {
		log.Fatal().Err(err).Msg("open store")
	}

	mgr := filter.New(db)
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := mgr.Start(rootCtx); err != nil {
		log.Fatal().Err(err).Msg("start filter manager")
	}

	api := server.NewAPI(db, mgr)
	mux := http.NewServeMux()
	api.Routes(mux)
	switch {
	case *webDir != "":
		fs := http.FileServer(http.Dir(*webDir))
		mux.Handle("/", fs)
		log.Info().Str("dir", *webDir).Msg("serving console UI from disk")
	default:
		if sub, ok := server.EmbeddedWebFS(); ok {
			mux.Handle("/", server.EmbeddedWebHandler(sub))
			log.Info().Msg("serving embedded console UI")
		} else {
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/" {
					fmt.Fprintf(w, "NEXTonebotfilter backend is running.\nConsole API at /api/*\n")
					return
				}
				http.NotFound(w, r)
			})
		}
	}

	consoleSrv := &http.Server{
		Addr:              *consoleAddr,
		Handler:           withCORS(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		log.Info().Str("addr", *consoleAddr).Msg("console API listening")
		if err := consoleSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("console server stopped")
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	log.Info().Msg("shutting down")
	mgr.Stop()
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutCancel()
	_ = consoleSrv.Shutdown(shutCtx)
}

func withCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		h.ServeHTTP(w, r)
	})
}
