package compress

import (
	"net/http"
	"strings"
)

const (
	jsonType = "application/json"
	htmlType = "text/html"
)

func GzipCompress(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ow := w

		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportGzip := strings.Contains(acceptEncoding, "gzip")
		if supportGzip && checkContentType(r) {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		contentEncoding := r.Header.Get("Content-Encoding")
		sendGzip := strings.Contains(contentEncoding, "gzip")
		if sendGzip {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			r.Body = cr
			defer cr.Close()
		}

		h.ServeHTTP(ow, r)
	}
}

func checkContentType(r *http.Request) bool {
	contentTypeValues := r.Header.Values("Content-Type")
	hasSupportType := false
	for _, value := range contentTypeValues {
		if strings.Contains(value, jsonType) || strings.Contains(value, htmlType) {
			hasSupportType = true
		}
	}

	return hasSupportType
}
