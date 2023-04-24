# go-form-data

multipart/form-data without memory buffering.

# Example

```go
package main

import (
	formdata "github.com/jc-lab/go-form-data"
	"log"
	"net/http"
	"os"
)

func main() {
	formData := formdata.NewFormData()

	formData.AddDataField("aaaa", "text/plain", []byte("hello world"))

	a, _ := os.OpenFile("aaa.txt", os.O_RDONLY, 0)
	formData.AddFileField("files[]", "application/octet-stream", a)
	
	b, _ := os.OpenFile("bbb.txt", os.O_RDONLY, 0)
	formData.AddFileField("files[]", "application/octet-stream", b)

	req, err := http.NewRequest("POST", "http://127.0.0.1/api/upload", formData)
	if err != nil {
		log.Fatalln(err)
	}
	
	resp, err := formData.Do(http.DefaultClient, req)
	if err != nil {
		log.Fatalln(err)
	}

	println(resp.Status)
	println(resp.StatusCode)
	println(resp.Body)
}
```

# License

[Apache-2.0](./LICENSE)