package nubio

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/ejuju/nubio/pkg/httpmux"
)

type Config struct {
	Address      string `json:"address"`        // Local HTTP server address.
	Profile      string `json:"profile"`        // Path to JSON file where profile data is stored.
	TrueIPHeader string `json:"true_ip_header"` // Ex: "X-Forwarded-For", useful when reverse proxying.
}

func Run(args ...string) (exitcode int) {
	// Init logger.
	slogh := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger := slog.New(slogh)
	logger.Debug("logger ready")

	// Load config.
	configPath := "local.config.json"
	if len(args) > 0 {
		configPath = args[0]
	}
	rawConfig, err := os.ReadFile(configPath)
	if err != nil {
		logger.Error("read config", "error", err)
		return 1
	}
	config := &Config{}
	err = json.Unmarshal(rawConfig, config)
	if err != nil {
		logger.Error("parse config", "error", err)
		return 1
	}

	// Load user profile.
	rawProfile, err := os.ReadFile(config.Profile)
	if err != nil {
		logger.Error("read profile", "error", err)
		return 1
	}
	profile := &Profile{}
	err = json.Unmarshal(rawProfile, profile)
	if err != nil {
		logger.Error("parse profile", "error", err)
		return 1
	}

	// Init and register HTTP endpoints.
	endpoints := httpmux.Map{
		PathPing:        {"GET": http.HandlerFunc(servePing)},
		PathFaviconSVG:  {"GET": http.HandlerFunc(serveFaviconSVG)},
		PathRobotsTXT:   {"GET": http.HandlerFunc(serveRobotsTXT)},
		PathSitemapXML:  {"GET": serveSitemapXML(profile.Domain)},
		PathHome:        {"GET": ExportAndServeHTML(profile)},
		PathProfileJSON: {"GET": ExportAndServeJSON(profile)},
		PathProfilePDF:  {"GET": ExportAndServePDF(profile)},
		PathProfileTXT:  {"GET": ExportAndServeText(profile)},
		PathProfileMD:   {"GET": ExportAndServeMarkdown(profile)},
	}

	router := endpoints.Handler(http.NotFoundHandler())

	// Wrap global middleware.
	//
	// Note: the panic recovery middleware relies on the:
	//	- True IP middleware
	//	- Request ID middleware
	//	- Logging middleware (to know if a response has been sent).
	//
	// This also means that any panic occuring in one of the above mentioned
	// middlewares propagates up and will cause the program to exit.
	router = httpmux.Wrap(router,
		httpmux.NewTrueIPMiddleware(config.TrueIPHeader),
		httpmux.NewRequestIDMiddleware(),
		httpmux.NewLoggingMiddleware(handleAccessLog(logger)),
		httpmux.NewPanicRecoveryMiddleware(handlePanic(logger)),
	)

	// Run HTTP server in separate Goroutine.
	// TODO: Support HTTPS.
	s := &http.Server{
		Addr:              config.Address,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       10 * time.Second,
		MaxHeaderBytes:    50_000,
	}
	errc := make(chan error, 1)
	go func() { errc <- s.ListenAndServe() }()

	// Wait for interrupt or server error.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	select {
	case err := <-errc:
		logger.Error("critical failure", "error", err)
		return
	case sig := <-interrupt:
		logger.Debug("shutting down", "signal", sig.String())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown HTTP server gracefully.
	err = s.Shutdown(ctx)
	if err != nil {
		logger.Error("shutdown HTTP server", "error", err)
	}

	// Done.
	logger.Debug("shutdown successful")
	return 0
}
