package main

import (
	"bufio"
	"fmt"
	"github.com/spf13/viper"
	"github.com/zhyee/zipstream"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

const DefaultRepoPath = "c:\\repo"
const MaxArchiveSize = 4 * 1024 * 1024 * 1024 * 1024

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	http.ServeFile(w, r, "assets/index.html")
}

func receiveFile(file io.Reader, wg *sync.WaitGroup, pipeWriter *io.PipeWriter) {
	_, err := io.Copy(pipeWriter, file)
	if err != nil {
		log.Printf("Error while receiving mime file: %s", err.Error())
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	wg.Done()
	pipeWriter.Close()
}

func unzipFile(dst string, wg *sync.WaitGroup, pipeReader *io.PipeReader) {
	zr := zipstream.NewReader(pipeReader)
	for {
		e, err := zr.GetNextEntry()
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			break
		}
		log.Println("entry name: ", e.Name)
		log.Println("entry comment: ", e.Comment)
		log.Println("entry reader version: ", e.ReaderVersion)
		log.Println("entry modify time: ", e.Modified)
		log.Println("entry compressed size: ", e.CompressedSize64)
		log.Println("entry uncompressed size: ", e.UncompressedSize64)
		log.Println("entry is a dir: ", e.IsDir())

		directory := fmt.Sprintf("%s/%s/", dst, filepath.Dir(e.Name))
		err = os.MkdirAll(directory, os.ModePerm)
		if err != nil {
			log.Printf("Error while creating directory %s: %s", directory, err.Error())
			//http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if e.IsDir() {
			continue
		}
		dstFilePath := fmt.Sprintf("%s/%s", dst, e.Name)
		file, err := os.Create(dstFilePath)
		if err != nil {
			log.Printf("Error while creating file %s: %s", dstFilePath, err.Error())
			//http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()

		data, err := e.Open()
		if err != nil {
			log.Fatalf("unable to open zip file: %s", err)
		}
		_, err = io.Copy(file, data)
		if err != nil {
			log.Printf("Error while writing unzipped file %s: %s", dstFilePath, err.Error())
			//http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Print(dstFilePath, " - Extracted.")
	}

	wg.Done()
	pipeReader.Close()
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	multipartReader, err := r.MultipartReader()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	p, err := multipartReader.NextPart()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if p.FormName() != "files" {
		http.Error(w, "files is expected in form date.", http.StatusBadRequest)
		return
	}
	buf := bufio.NewReader(p)
	sniff, _ := buf.Peek(512)
	contentType := http.DetectContentType(sniff)
	if contentType != "application/zip" {
		http.Error(w, "file type not allowed", http.StatusBadRequest)
		return
	}
	log.Print("Receiving filename ", p.FileName())

	dst := fmt.Sprintf("%s/%s/", viper.GetString("repopath"), p.FileName())
	err = os.MkdirAll(dst, os.ModePerm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	pipeReader, pipeWriter := io.Pipe()

	go receiveFile(buf, &wg, pipeWriter)
	go unzipFile(dst, &wg, pipeReader)
	wg.Wait()
}

func main() {
	viper.SetDefault("repopath", DefaultRepoPath)

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/archive", uploadHandler)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
