package mimecheck

import "net/http"

func IsJSONContentTypeCorrect(r *http.Request) bool {
	if len(r.Header.Values("Content-Type")) == 0 {
		return false
	}

	for contentTypeCurrentIndex, contentType := range r.Header.Values("Content-Type") {
		if contentType == "application/json" {
			break
		}

		if contentTypeCurrentIndex == len(r.Header.Values("Content-Type"))-1 {
			return false
		}
	}

	return true
}
