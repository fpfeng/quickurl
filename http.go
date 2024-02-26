package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
)

const DownThemAllArchiveFilename = "DownThemAll"

type HttpHandlers struct {
	QuickURL *QuickURL
}

var MIME = map[string]string{
	"tar.gz": "application/gzip",
	"zip":    "application/zip",
}

func (quh *HttpHandlers) CreateArchive(w http.ResponseWriter, r *http.Request) {
	filename := mux.Vars(r)["filename"]
	format := mux.Vars(r)["archive"]
	contentType, ok := MIME[format]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	archiveFilename := fmt.Sprintf("%v.%v", filename, format) // TODO is it too simple?
	log.Debugf("request %v", archiveFilename)

	if path, ok := quh.QuickURL.ServingEntries[filename]; ok {
		result, err := quh.QuickURL.CreateArchive([]string{path}, format)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatal(err)
			return
		}

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, archiveFilename))
		http.ServeContent(w, r, archiveFilename, time.Now(), bytes.NewReader(result))
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (quh *HttpHandlers) OriginalFile(w http.ResponseWriter, r *http.Request) {
	filename := mux.Vars(r)["filename"]
	log.Debug(filename)
	if path, exists := quh.QuickURL.ServingEntries[filename]; exists {
		http.ServeFile(w, r, path)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (quh *HttpHandlers) DownThemAll(w http.ResponseWriter, r *http.Request) {
	format := mux.Vars(r)["archive"]
	contentType, ok := MIME[format]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	archiveFilename := fmt.Sprintf("%v_%v.%v", DownThemAllArchiveFilename, time.Now().Unix(), format) // TODO is it too simple?
	log.Debugf("request %v", archiveFilename)

	result, err := quh.QuickURL.CreateArchive(maps.Values(quh.QuickURL.ServingEntries), format)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal(err)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, archiveFilename))
	http.ServeContent(w, r, archiveFilename, time.Now(), bytes.NewReader(result))
}
