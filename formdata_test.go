package formdata

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"io"
	"log"
	"net/http"
	"os"
	"testing"
)

type MockedClient struct {
	mock.Mock
	HttpClientable

	buffer *bytes.Buffer
}

func (m MockedClient) Do(req *http.Request) (*http.Response, error) {
	_, err := io.Copy(m.buffer, req.Body)
	return nil, err
}

func Test_Sample(t *testing.T) {
	client := &MockedClient{
		buffer: &bytes.Buffer{},
	}

	formData := NewFormData()
	formData.AddDataField("aaaa", "text/plain", []byte("hello world"))
	a, _ := os.OpenFile("test/sample-1.txt", os.O_RDONLY, 0)
	formData.AddFileField("files[]", "application/octet-stream", a)
	b, _ := os.OpenFile("test/sample-2.txt", os.O_RDONLY, 0)
	formData.AddFileField("files[]", "application/octet-stream", b)

	req, err := http.NewRequest("POST", "test://dummy", formData)
	if err != nil {
		log.Fatalln(err)
	}

	_, err = formData.Do(client, req)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, len(client.buffer.Bytes()), int(req.ContentLength))
}
