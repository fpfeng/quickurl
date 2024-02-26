package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"text/template"

	log "github.com/sirupsen/logrus"
)

type QuickURL struct {
	ListeningPort  int
	ServingEntries map[string]string
	PublicIPOnly   bool
}

type servingResource struct {
	Name string   `json:"name"`
	URLs []string `json:"urls"`
}

func NewQuickURL(cliConfig *CLIParseResult) *QuickURL {
	if len(cliConfig.Entries) == 0 {
		os.Exit(0)
	}
	qu := &QuickURL{
		ListeningPort:  cliConfig.Port,
		ServingEntries: map[string]string{},
		PublicIPOnly:   cliConfig.PublicIPOnly,
	}
	for _, entryPath := range cliConfig.Entries {
		qu.AddServingEntry(entryPath)
	}
	return qu
}

func (qu *QuickURL) AddServingEntry(path string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}
	name := filepath.Base(path)
	// TODO duplicate check
	qu.ServingEntries[name] = absPath
}

func (qu *QuickURL) ListServingEntryNames() []string {
	keys := make([]string, 0, len(qu.ServingEntries))
	for k := range qu.ServingEntries {
		keys = append(keys, k)
	}
	return keys
}

func (qu *QuickURL) CreateTarGz(fullPaths []string) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzipWriter)

	for _, path := range fullPaths {
		stat, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		mode := stat.Mode()
		if mode.IsRegular() {
			header, err := tar.FileInfoHeader(stat, path)
			if err != nil {
				return nil, err
			}
			if err := tarWriter.WriteHeader(header); err != nil {
				return nil, err
			}
			data, err := os.Open(path)
			if err != nil {
				return nil, err
			}
			if _, err := io.Copy(tarWriter, data); err != nil {
				return nil, err
			}
		} else if mode.IsDir() {
			filepath.Walk(path, func(subPath string, fi os.FileInfo, _err error) error {
				if fi.IsDir() {
					return nil
				}
				header, err := tar.FileInfoHeader(fi, subPath)
				if err != nil {
					return err
				}
				// https://golang.org/src/archive/tar/common.go?#L626
				relPath, err := filepath.Rel(path, subPath)
				if err != nil {
					return err
				}
				header.Name = filepath.ToSlash(relPath)
				if err := tarWriter.WriteHeader(header); err != nil {
					return err
				}
				data, err := os.Open(subPath)
				if err != nil {
					return err
				}
				if _, err := io.Copy(tarWriter, data); err != nil {
					return err
				}
				return nil
			})
		} else {
			return nil, errors.New("error: file type not supported")
		}
	}
	if err := tarWriter.Close(); err != nil {
		return nil, err
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (qu *QuickURL) CreateZip(fullPaths []string) ([]byte, error) {
	var buffer bytes.Buffer
	zipWriter := zip.NewWriter(&buffer)

	for _, path := range fullPaths {
		stat, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		mode := stat.Mode()
		if mode.IsRegular() {
			header, err := zip.FileInfoHeader(stat)
			if err != nil {
				return nil, err
			}

			fileWriter, err := zipWriter.Create(header.Name)
			if err != nil {
				return nil, err
			}

			data, err := os.Open(path)
			if err != nil {
				return nil, err
			}

			_, err = io.Copy(fileWriter, data)
			if err != nil {
				return nil, err
			}
		} else if mode.IsDir() {
			filepath.Walk(path, func(subPath string, fi os.FileInfo, _err error) error {
				if fi.IsDir() {
					return nil
				}
				if err != nil {
					return err
				}
				relPath, err := filepath.Rel(path, subPath)
				if err != nil {
					return err
				}
				relPath = filepath.ToSlash(relPath)
				fileWriter, err := zipWriter.Create(relPath)
				if err != nil {
					return err
				}

				data, err := os.Open(subPath)
				if err != nil {
					return err
				}
				if _, err := io.Copy(fileWriter, data); err != nil {
					return err
				}
				return nil
			})
		} else {
			return nil, errors.New("error: file type not supported")
		}
	}

	// TODO fix defer close would cause empty buffer
	if err := zipWriter.Close(); err != nil {
		return nil, err
	}
	result := buffer.Bytes()
	return result, nil

}

func (qu *QuickURL) CreateArchive(filePaths []string, format string) ([]byte, error) {
	switch format {
	case "tar.gz":
		return qu.CreateTarGz(filePaths)
	case "zip":
		return qu.CreateZip(filePaths)
	default:
		return nil, fmt.Errorf("unsupport format: %v", format)
	}
}

func generateArchiveURLs(url url.URL) []string {
	urls := make([]string, 0)
	q := url.Query()
	q.Set("archive", "zip")
	url.RawQuery = q.Encode()
	urls = append(urls, url.String())
	q.Set("archive", "tar.gz")
	url.RawQuery = q.Encode()
	urls = append(urls, url.String())
	return urls
}

func (qu *QuickURL) generateAccessURLs() []*servingResource {
	allURLs := map[string][]string{}
	for _, addr := range GetMachineAddresses(qu.PublicIPOnly) {
		for baseName, fullPath := range qu.ServingEntries {
			if _, ok := allURLs[baseName]; !ok {
				allURLs[baseName] = make([]string, 0)
			}

			stat, err := os.Stat(fullPath)
			if err != nil {
				log.Fatal(err)
			}
			url := BuildAccessURL(addr, qu.ListeningPort, baseName)
			if stat.IsDir() {
				allURLs[baseName] = append(allURLs[baseName], generateArchiveURLs(url)...)
			} else { // is file
				allURLs[baseName] = append(allURLs[baseName], url.String())
				allURLs[baseName] = append(allURLs[baseName], generateArchiveURLs(url)...)
			}
		}
		if len(qu.ServingEntries) > 1 {
			url := generateArchiveURLs(BuildAccessURL(addr, qu.ListeningPort, DownThemAllArchiveFilename))
			allURLs[DownThemAllArchiveFilename] = append(allURLs[DownThemAllArchiveFilename], url...)
		}
	}
	resources := make([]*servingResource, 0)
	for k, v := range allURLs {
		res := servingResource{Name: k, URLs: v}
		resources = append(resources, &res)
	}
	return resources
}

func (qu *QuickURL) PrintAccessURLs() {
	resources := qu.generateAccessURLs()
	tpl, err := template.New("URLs").Parse(
		`{{range .}} 
	{{range .URLs}}{{.}}
	{{end}}{{end}}`)
	if err != nil {
		log.Fatal(err)
	}

	err = tpl.Execute(os.Stdout, resources)
	if err != nil {
		log.Fatal(err)
	}
}
