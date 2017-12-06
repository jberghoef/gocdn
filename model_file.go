package main

import (
	"net/http"
)

/* File
The file model
================================================================================ */
type File struct {
	Reference   string      `json:"reference"`
	URL         string      `json:"url"`
	LocalFile   string      `json:"localfile"`
	Timestamp   int64       `json:"timestamp"`
	MaxAge      int64       `json:"maxage"`
	Revalidate  bool        `json:"revalidate"`
	ETAG        string      `json:"etag"`
	ContentType string      `json:"contenttype"`
	Header      http.Header `json:"header"`
}

func assesContentType(contentType string) bool {
	allowed := false

	allowedFormats := [18]string{
		"text/css",
		"text/html",
		"text/javascript",
		"image/svg+xml",
		"image/vnd.microsoft.icon",
		"image/x-icon",
		"image/gif",
		"image/png",
		"image/jpeg",
		"image/bmp",
		"image/webp",
		"audio/midi",
		"audio/mpeg",
		"audio/webm",
		"audio/ogg",
		"audio/wav",
		"video/webm",
		"video/ogg",
	}

	for _, a := range allowedFormats {
		if a == contentType {
			allowed = true
		}
	}

	return allowed
}
