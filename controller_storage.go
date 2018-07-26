package main

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/boltdb/bolt"

	"time"

	"encoding/json"

	"gopkg.in/kyokomi/emoji.v1"
	"os"
	"io"
	"bytes"
)

/* verifyAndRetrieveFile
Checks original source for the requested file and retrieves it when available.
====================================================================== */
func verifyAndRetrieveFile(localFile string, originURL string, w http.ResponseWriter, r *http.Request) {

	// Retrieve and verify the Headers before moving forward
	head, err := http.Head(originURL)
	if err != nil {
		fmt.Println("Error while retreiving Headers from", originURL, "\n-", err)
		return
	}

	if head.StatusCode == http.StatusOK {
		// Retrieve file from host
		response, err := http.Get(originURL)
		defer response.Body.Close()

		var buf bytes.Buffer
		tee := io.TeeReader(response.Body, &buf)

		if err == nil {
			for name := range head.Header {
				w.Header().Set(name, head.Header.Get(name))
			}
			io.Copy(w, tee)

			allowed := verifyContentType(head.Header)
			ignore, revalidate, maxAge := defineCacheControl(head.Header)
			if allowed == true && ignore != true {
				// Create directories
				fileDir := filepath.Dir(localFile)
				if os.MkdirAll(fileDir, 0755) == nil {

					// Prepare file for writing
					output, err := os.Create(localFile)
					if err != nil {
						fmt.Println("Error while creating", fileDir, "\n-", err)
						return
					}
					defer output.Close()

					// Write data to file
					n, err := io.Copy(output, &buf)
					if err != nil {
						fmt.Println("Error while writing", fileDir, "\n-", err)
						return
					}

					relPath, _ := filepath.Rel(cacheDir, localFile)
					emoji.Println("\t:truck: Retrieving file:", originURL, "\n\t:memo: Writing", n, "bytes to", relPath)

					file := File{
						Reference:  createHash(originURL),
						URL:        originURL,
						LocalFile:  localFile,
						Timestamp:  time.Now().Unix(),
						MaxAge:     int64(maxAge),
						Revalidate: revalidate,
						ETAG:       head.Header.Get("etag"),
						Header:     head.Header}

					file.Register()
				}
			}
		} else {
			fmt.Println("Error while downloading", originURL, "\n-", err)
			http.Redirect(w, r, originURL, 302)
		}
	}
}

/* validateCache
Checks whether the file is still valid based on saved cache rules.
====================================================================== */
func validateCache(originURL string) (valid bool) {
	now := time.Now().Unix()
	refHash := createHash(originURL)

	valid = false

	file := File{}

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Cache"))
		v := b.Get([]byte(refHash))

		json.Unmarshal(v, &file)
		return nil
	})

	if now <= (file.Timestamp + file.MaxAge) {
		valid = true
	} else if file.Revalidate {
		response, err := http.Head(originURL)
		if err != nil {
			fmt.Println("Error while retreiving Headers from", originURL, "\n-", err)
			return
		}

		if response.Header.Get("etag") == file.ETAG {
			valid = true
		} else {
			file.ETAG = response.Header.Get("etag")
		}
	}

	if !valid {
		relPath, _ := filepath.Rel(cacheDir, file.LocalFile)
		emoji.Println(":anger: File no longer valid:", relPath)
		file.Remove()
	}

	return valid
}
