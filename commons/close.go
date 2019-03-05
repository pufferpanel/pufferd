package commons

import (
	"io"
	"net/http"
)

func Close(closer io.Closer) {
	if closer != nil {
		closer.Close()
	}
}

func CloseResponse(response *http.Response) {
	if response != nil && response.Body != nil {
		Close(response.Body)
	}
}