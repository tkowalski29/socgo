package handlers

import (
	"compress/gzip"
	"net/http"
	"strings"
)

// GzipMiddleware compresses responses using gzip
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client supports gzip
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Check if response should be compressed
		contentType := r.Header.Get("Content-Type")
		if !shouldCompress(contentType) {
			next.ServeHTTP(w, r)
			return
		}

		// Create gzip writer
		gw := gzip.NewWriter(w)
		defer gw.Close()

		// Set headers
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		// Create response writer that writes to gzip
		gzipWriter := &gzipResponseWriter{
			ResponseWriter: w,
			gzipWriter:     gw,
		}

		next.ServeHTTP(gzipWriter, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gzipWriter *gzip.Writer
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.gzipWriter.Write(b)
}

func (g *gzipResponseWriter) WriteString(s string) (int, error) {
	return g.gzipWriter.Write([]byte(s))
}

// shouldCompress determines if content should be compressed
func shouldCompress(contentType string) bool {
	compressibleTypes := []string{
		"text/html",
		"text/plain",
		"text/css",
		"text/javascript",
		"application/javascript",
		"application/json",
		"application/xml",
		"application/xml+rss",
		"text/xml",
	}

	for _, t := range compressibleTypes {
		if strings.Contains(contentType, t) {
			return true
		}
	}
	return false
}
