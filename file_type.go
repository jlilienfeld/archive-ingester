package main

import (
	"bufio"
	"fmt"
	"net/http"
	"path"
)

type FileType int

const (
	Unknown FileType = iota
	Zip
	Xz
	Tar
	Gz
	SevenZip
	BZ2
	Rar
)

func PeekFileType(fileName string, r *bufio.Reader) (string, FileType) {
	sniff, _ := r.Peek(512)
	contentType := http.DetectContentType(sniff)
	fileType := GetFileType(contentType)
	if fileType == Zip && path.Ext(fileName) != ".zip" {
		// Do not unzip if it doesn't have the extension.  Many documents are zip files, and they become harder
		// to consume if unzipped.
		fileType = Unknown
	}
	return contentType, fileType
}

func GetFileType(contentType string) FileType {
	switch contentType {
	case "application/zip":
		return Zip
	default:
		return Unknown
	}
}

func (e FileType) IsUnknown() bool {
	return e == Unknown
}

func (e FileType) String() string {
	switch e {
	case Unknown:
		return "Unknown"
	case Zip:
		return "Zip"
	default:
		return fmt.Sprintf("%d", int(e))
	}
}
