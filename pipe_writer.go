package main

import "io"

type PipeWriter struct {
	io.Writer
}

func (w PipeWriter) WriteAt(p []byte, offset int64) (n int, err error) {
	return w.Write(p)
}
