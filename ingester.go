package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Ingester interface {
	Ingest(contentType string, dstPath string, r io.Reader) (int, error)
}

func GetIngesterFor(fileType FileType) Ingester {
	switch fileType {
	case Zip:
		return ZipIngester{}
	default:
		return nil
	}
}

func storeFile(contentType string, dst string, fileName string, buf *bufio.Reader) (int, error) {
	directory := fmt.Sprintf("%s/%s/", dst, filepath.Dir(fileName))
	err := os.MkdirAll(directory, os.ModePerm)
	if err != nil {
		return http.StatusInternalServerError,
			errors.New(fmt.Sprintf("Error while creating directory %s: %s", directory, err.Error()))
	}

	dstFilePath := fmt.Sprintf("%s/%s", dst, fileName)
	file, err := os.Create(dstFilePath)
	if err != nil {
		return http.StatusInternalServerError,
			errors.New(fmt.Sprintf("Error while creating file %s: %s", dstFilePath, err.Error()))
	}
	defer file.Close()

	_, err = io.Copy(file, buf)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	log.Print(dstFilePath, " Extracted ---- Content-Type: ", contentType)

	return http.StatusOK, nil
}

func IngestFile(dstPath string, fileName string, r *bufio.Reader) (int, error) {
	contentType, fileType := PeekFileType(fileName, r)
	if fileType.IsUnknown() {
		return storeFile(contentType, dstPath, fileName, r)
	}

	fileIngester := GetIngesterFor(fileType)
	if fileIngester == nil {
		return storeFile(contentType, dstPath, fileName, r)
	}

	return fileIngester.Ingest(contentType, fmt.Sprintf("%s/%s/", dstPath, fileName), r)
}
