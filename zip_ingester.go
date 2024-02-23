package main

import (
	"bufio"
	"github.com/zhyee/zipstream"
	"io"
	"log"
	"net/http"
	"sync"
)

type ZipIngester struct {
}

func receiveFile(dst string, file io.Reader, wg *sync.WaitGroup, pipeWriter *io.PipeWriter) {
	defer wg.Done()

	_, err := io.Copy(pipeWriter, file)
	if err != nil {
		log.Printf("Error while receiving mime file: %s", err.Error())
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Print("Left receiveFile, dst:", dst)
}

func unzipFile(dst string, wg *sync.WaitGroup, pipeReader *io.PipeReader) {
	defer wg.Done()
	defer pipeReader.Close()

	zr := zipstream.NewReader(pipeReader)
	for {
		e, err := zr.GetNextEntry()
		if err != nil {
			if err != io.EOF {
				log.Fatal(err)
			}
			log.Print(err.Error())
			break
		}

		if e.IsDir() {
			continue
		}

		data, err := e.Open()

		if err != nil {
			log.Fatalf("unable to open zip file: %s", err)
		}
		log.Print("Zip uncompressed entry size: ", e.UncompressedSize64)
		buf := bufio.NewReaderSize(data, 4<<20)
		IngestFile(dst, e.Name, buf)
		data.Close()
	}
	log.Print("Left unzipFile, dst: ", dst)
}

func (z ZipIngester) Ingest(contentType string, dstPath string, r io.Reader) (int, error) {
	log.Print("Ingesting ", dstPath, " ---- Content-Type: ", contentType)

	wg := sync.WaitGroup{}
	wg.Add(2)
	pipeReader, pipeWriter := io.Pipe()
	defer pipeWriter.Close()

	go receiveFile(dstPath, r, &wg, pipeWriter)
	go unzipFile(dstPath, &wg, pipeReader)
	wg.Wait()
	return http.StatusOK, nil
}
