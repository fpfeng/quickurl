package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

type QuickURL struct {
	ListeningPort int
	ServingFiles  map[string]string
	PublicIPOnly  bool
}

func NewQuickURL(cliConfig *CLIParseResult) *QuickURL {
	if len(cliConfig.Files) == 0 {
		os.Exit(0)
	}
	qu := &QuickURL{
		ListeningPort: cliConfig.Port,
		ServingFiles:  map[string]string{},
		PublicIPOnly:  cliConfig.PublicIPOnly,
	}
	for _, filepath := range cliConfig.Files {
		qu.AddServingFile(filepath)
	}
	return qu
}

func (qu *QuickURL) AddServingFile(path string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatal(err)
	}
	filename := filepath.Base(path)
	// TODO duplicate check
	qu.ServingFiles[filename] = absPath
}

func (qu *QuickURL) ListFileNames() []string {
	keys := make([]string, 0, len(qu.ServingFiles))
	for k := range qu.ServingFiles {
		keys = append(keys, k)
	}
	return keys
}

func (qu *QuickURL) CreateTarGz(filePaths []string) ([]byte, error) {
	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	for _, filePath := range filePaths {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			return nil, err
		}

		header := &tar.Header{
			Name: fileInfo.Name(),
			Size: fileInfo.Size(),
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return nil, err
		}

		_, err = io.Copy(tarWriter, file)
		if err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

func (qu *QuickURL) CreateZip(filePaths []string) ([]byte, error) {
	var buffer bytes.Buffer
	zipWriter := zip.NewWriter(&buffer)

	for _, filePath := range filePaths {
		file, err := os.Open(filePath)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			return nil, err
		}

		fileWriter, err := zipWriter.Create(fileInfo.Name())

		if err != nil {
			return nil, err
		}

		_, err = io.Copy(fileWriter, file)
		if err != nil {
			return nil, err
		}
	}
	// TODO fix defer close would cause empty buffer
	zipWriter.Close()
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

func printArchiveURLs(url url.URL) {
	q := url.Query()
	q.Set("archive", "zip")
	url.RawQuery = q.Encode()
	fmt.Printf("%v\n", url.String())
	q.Set("archive", "tar.gz")
	url.RawQuery = q.Encode()
	fmt.Printf("%v\n\n", url.String())
}

func (qu *QuickURL) PrintAccessURLs() {
	// TODO replace with template
	for _, addr := range GetMachineAddresses(qu.PublicIPOnly) {
		for _, filename := range qu.ListFileNames() {
			url := BuildAccessURL(addr, qu.ListeningPort, filename)
			fmt.Printf("%v\n", url.String())
			printArchiveURLs(url)
		}
		if len(qu.ListFileNames()) > 1 {
			printArchiveURLs(BuildAccessURL(addr, qu.ListeningPort, DownThemAllArchiveFilename))
		}
	}
}
