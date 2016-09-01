package net

import "io"

type Writer struct {
	conn io.WriteCloser
}

func NewWriter(conn io.WriteCloser) (w *Writer) {
	return &Writer{
		conn: conn,
	}
}

func (w *Writer) Write(buffer []byte) (n int, e error) {

	return
}

func (w *Writer) Close() error {
	return w.conn.Close()
}
