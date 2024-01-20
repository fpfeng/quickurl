package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

func main() {
	cliConfig := StartCLI(os.Args)
	quickURL := NewQuickURL(cliConfig)
	handlers := HttpHandlers{QuickURL: quickURL}
	router := mux.NewRouter()
	router.HandleFunc(fmt.Sprintf("/%v", DownThemAllArchiveFilename),
		handlers.DownThemAll).Queries("archive", "{archive}")
	router.HandleFunc("/{filename}",
		handlers.CreateArchive).Queries("archive", "{archive}")
	router.HandleFunc("/{filename}", handlers.OriginalFile)
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%v", quickURL.ListeningPort),
		Handler: router,
	}
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %+v\n", err)
		}
	}()

	go func() {
		quickURL.PrintAccessURLs()
	}()

	log.Debug("started..")

	<-done
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	log.Debug("end..")

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("shutdown failed: %+v", err)
	}

}
