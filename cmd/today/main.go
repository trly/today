package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/trly/today/internal/caldav"
	"github.com/trly/today/internal/events"
	"github.com/trly/today/internal/httpapi"
	"github.com/trly/today/web"
)

func main() {
	cmd := &cli.Command{
		Name:  "Today",
		Usage: "serve Today's CalDAV events API",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "addr",
				Usage:   "HTTP server address (or $ADDR)",
				Value:   ":8080",
				Sources: cli.EnvVars("ADDR"),
			},
			&cli.StringFlag{
				Name:    "url",
				Usage:   "CalDAV server endpoint URL (or $CALDAV_URL)",
				Sources: cli.EnvVars("CALDAV_URL"),
			},
			&cli.StringFlag{
				Name:    "user",
				Usage:   "HTTP basic auth username (or $CALDAV_USER)",
				Sources: cli.EnvVars("CALDAV_USER"),
			},
			&cli.StringFlag{
				Name:    "password",
				Usage:   "HTTP basic auth password (or $CALDAV_PASSWORD)",
				Sources: cli.EnvVars("CALDAV_PASSWORD"),
			},
		},
		Action: run,
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(_ context.Context, cmd *cli.Command) error {
	cfg, err := configFromCommand(cmd)
	if err != nil {
		return err
	}
	client, err := caldav.NewClient(cfg.CalDAVURL, cfg.CalDAVUser, cfg.CalDAVPassword)
	if err != nil {
		return err
	}
	service := events.Service{Client: client, Location: time.Local}
	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)
	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler(service),
		Protocols:         protocols,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("listening on %s", cfg.Addr)
	return server.ListenAndServe()
}

type config struct {
	Addr           string
	CalDAVURL      string
	CalDAVUser     string
	CalDAVPassword string
}

func configFromCommand(cmd *cli.Command) (config, error) {
	cfg := config{
		Addr:           cmd.String("addr"),
		CalDAVURL:      cmd.String("url"),
		CalDAVUser:     cmd.String("user"),
		CalDAVPassword: cmd.String("password"),
	}
	if cfg.CalDAVURL == "" {
		return config{}, errors.New("--url is required (or set $CALDAV_URL)")
	}
	if cfg.CalDAVUser == "" && cfg.CalDAVPassword != "" {
		return config{}, errors.New("--password requires --user")
	}
	return cfg, nil
}

func spaHandler() http.Handler {
	fsys, err := fs.Sub(web.Static, "dist")
	if err != nil {
		return http.NotFoundHandler()
	}
	fileServer := http.FileServer(http.FS(fsys))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path != "" {
			if _, err := fs.Stat(fsys, path); err == nil {
				fileServer.ServeHTTP(w, r)
				return
			}
		}
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}

func handler(service httpapi.Service) http.Handler {
	api := httpapi.New(service)
	frontend := spaHandler()
	mux := http.NewServeMux()
	mux.Handle("/today.v1.", api)
	mux.Handle("/api-docs", api)
	mux.Handle("/", frontend)
	return mux
}
