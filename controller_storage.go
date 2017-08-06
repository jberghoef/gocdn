package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"

	"time"

	"encoding/json"

	emoji "gopkg.in/kyokomi/emoji.v1"
)

/* verifyAndRetrieveFile
Checks original source for the requested file and retrieves it when available.
================================================================================ */
func verifyAndRetrieveFile(localFile string, originURL string) {

	// Retrieve and verify the Headers before moving forward
	response, err := http.Head(originURL)
	if err != nil {
		fmt.Println("Error while retreiving Headers from", originURL, "\n-", err)
		return
	}

	allowed := verifyContentType(response.Header)

	// Check if we need to download the file
	if response.StatusCode == http.StatusOK && allowed {
		ignore, revalidate, maxAge := defineCacheControl(response.Header)

		if ignore != true {

			// Retrieve file from host
			response, err := http.Get(originURL)
			if err == nil {

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
					n, err := io.Copy(output, response.Body)
					if err != nil {
						fmt.Println("Error while writing", fileDir, "\n-", err)
						return
					}

					relPath, _ := filepath.Rel(cacheDir, localFile)
					emoji.Println("\t:truck: Retreiving file:", originURL, "\n\t:memo: Writing", n, "bytes to", relPath)

					go registerFileToDb(originURL, relPath, maxAge, revalidate, response.Header)

				} else {
					fmt.Println("Unable to create directory for cache file!")
				}
			} else {
				fmt.Println("Error while downloading", originURL, "\n-", err)
				return
			}
			defer response.Body.Close()

			return
		}
	}
}

/* registerFileToDb
Registers metadata of the retrieved file to the database.
================================================================================ */
func registerFileToDb(originURL string, localFile string, maxAge int, revalidate bool, header http.Header) {

	refHash := createHash(originURL)

	file := File{
		Reference:  refHash,
		URL:        originURL,
		LocalFile:  localFile,
		Timestamp:  time.Now().Unix(),
		MaxAge:     int64(maxAge),
		Revalidate: revalidate,
		ETAG:       header.Get("etag"),
		Header:     header}

	fileJSON, _ := json.Marshal(file)

	db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Cache"))
		err := b.Put([]byte(refHash), fileJSON)
		return err
	})
}

/* validateCache
Checks whether the file is still valid based on saved cache rules.
================================================================================ */
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

		go removeFile(file)
	}

	return valid
}
