package main

import (
	"bufio"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"net/http"
)

const DefaultRepoPath = "c:\\repo"
const MaxArchiveSize = 4 * 1024 * 1024 * 1024 * 1024

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	http.ServeFile(w, r, "assets/index.html")
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
	fileName := p.FileName()

	buf := bufio.NewReader(p)
	contentType, fileType := PeekFileType(fileName, buf)
	if fileType.IsUnknown() {
		http.Error(w, fmt.Sprintf("Content-Type %s not allowed.", contentType), http.StatusBadRequest)
		return
	}
	log.Print("Receiving filename ", fileName, " ---- Content-Type: ", contentType)
	rc, err := IngestFile(viper.GetString("repopath"), fileName, buf)
	if err != nil {
		http.Error(w, err.Error(), rc)
		return
	}
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
