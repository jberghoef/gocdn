package main

import "net/http"

/* File
The file model
================================================================================ */
type File struct {
	Reference  string      `json:"reference"`
	URL        string      `json:"url"`
	LocalFile  string      `json:"localfile"`
	Timestamp  int64       `json:"timestamp"`
	MaxAge     int64       `json:"maxage"`
	Revalidate bool        `json:"revalidate"`
	ETAG       string      `json:"etag"`
	Header 	   http.Header `json:"header"`
}
