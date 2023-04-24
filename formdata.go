package formdata

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"path"
)

type HttpClientable interface {
	Do(req *http.Request) (*http.Response, error)
}

type formField struct {
	header string
	size   int64
	raw    []byte
	file   *os.File
}

type FormData struct {
	io.ReadCloser
	boundary string
	fields   []*formField
	closed   bool
	pipeIn   *io.PipeWriter
	pipeOut  *io.PipeReader
}

func NewFormData() *FormData {
	boundary := "----GoFormBoundary"

	random := make([]byte, 12)
	rand.Read(random)
	boundary += base64.RawURLEncoding.EncodeToString(random)

	return &FormData{
		boundary: boundary,
	}
}

func (f *FormData) Boundary() string {
	return f.boundary
}

func (f *FormData) AddDataField(name string, contentType string, data []byte) {
	header := "--" + f.boundary + "\r\n"
	header += "Content-Disposition: form-data; name=\"" + name + "\"\r\n"
	if len(contentType) > 0 {
		header += "Content-Type: " + contentType + "\r\n"
	}
	header += "\r\n"

	f.fields = append(f.fields, &formField{
		header: header,
		size:   int64(len(data)),
		raw:    data,
	})
}

func (f *FormData) AddFileField(name string, contentType string, file *os.File) error {
	stat, err := file.Stat()
	if err != nil {
		return err
	}

	basename := path.Base(file.Name())

	header := "--" + f.boundary + "\r\n"
	header += "Content-Disposition: form-data; name=\"" + name + "\"; filename=\"" + basename + "\"\r\n"
	if len(contentType) > 0 {
		header += "Content-Type: " + contentType + "\r\n"
	}
	header += "\r\n"

	f.fields = append(f.fields, &formField{
		header: header,
		size:   stat.Size(),
		file:   file,
	})

	return nil
}

func (f *FormData) Do(client HttpClientable, req *http.Request) (*http.Response, error) {
	var contentLength int64 = 0
	for _, field := range f.fields {
		contentLength += int64(len(field.header)) + field.size + 2
	}
	contentLength += int64(len(f.boundary)) + 6

	req.Header.Set("content-type", "multipart/form-data; boundary="+f.boundary)
	req.ContentLength = contentLength

	io.Pipe()
	f.pipeOut, f.pipeIn = io.Pipe()

	go func() {
		for i := 0; i < len(f.fields); i++ {
			field := f.fields[i]

			f.pipeIn.Write([]byte(field.header))
			if field.file != nil {
				io.Copy(f.pipeIn, field.file)
			} else {
				f.pipeIn.Write(field.raw)
			}
			f.pipeIn.Write([]byte("\r\n"))
		}

		f.pipeIn.Write([]byte("--" + f.boundary + "--\r\n"))
		f.pipeIn.Close()
	}()

	return client.Do(req)
}

func (f *FormData) Read(p []byte) (int, error) {
	n, err := f.pipeOut.Read(p)
	return n, err
}

func (f *FormData) Close() error {
	f.cleanup()
	return nil
}

func (f *FormData) cleanup() {
	if f.closed {
		return
	}
	f.closed = true
	for _, field := range f.fields {
		if field.file != nil {
			field.file.Close()
		}
	}
}
