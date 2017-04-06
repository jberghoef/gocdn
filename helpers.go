package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	emoji "gopkg.in/kyokomi/emoji.v1"

	"github.com/boltdb/bolt"
)

/* verifyContentType
Reads the original response headers to parse the content type.
================================================================================ */
func verifyContentType(header http.Header) (allowed bool) {
	contentTypes := strings.Split(header.Get("Content-Type"), "; ")
	allowed = false

	allowedFormats := [16]string{
		"text/css",
		"text/javascript",
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

	for _, f := range contentTypes {
		for _, a := range allowedFormats {
			if a == f {
				allowed = true
			}
		}
	}

	return allowed
}

/* defineCacheControl
Reads the original response headers to parse Cache-Control rules.
================================================================================ */
func defineCacheControl(header http.Header) (ignore bool, revalidate bool, maxAge int) {
	cacheRules := strings.Split(header.Get("Cache-Control"), ", ")
	ignore = false
	revalidate = false
	maxAge = 3600 * 1 // Defaults to 1 hour

	for _, rule := range cacheRules {
		switch {
		case strings.Contains(rule, "private") || strings.Contains(rule, "no-store"):
			ignore = true
		case strings.Contains(rule, "no-cache") || strings.Contains(rule, "must-revalidate") || strings.Contains(rule, "proxy-revalidate"):
			revalidate = true
		case strings.Contains(rule, "max-age"):
			i, err := strconv.Atoi(rule[8:])
			if err == nil {
				maxAge = i
			} else {
				fmt.Println("Error while converting max age string to int", "\n-", err)
			}
		}
	}

	return ignore, revalidate, maxAge
}

/* createHash
Creates a hash based on the origin url.
================================================================================ */
func createHash(originURL string) (refHash string) {
	h := md5.New()
	h.Write([]byte(originURL))
	refHash = hex.EncodeToString(h.Sum(nil))

	return refHash
}

/* removeFile
Removes to referenced file and it's database entry.
================================================================================ */
func removeFile(file File) {
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Cache"))
		b.Delete([]byte(file.Reference))
		return nil
	})

	os.Remove(file.LocalFile)
}

/* cleanCache
Checks the entire database for expired files.
================================================================================ */
func cleanCache() {
	now := time.Now().Unix()

	db.Batch(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Cache"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			file := File{}

			json.Unmarshal(v, &file)

			if now <= (file.Timestamp + file.MaxAge) {
				emoji.Println(":fire: Removing expired file:", file.LocalFile)
				b.Delete([]byte(k))

				os.Remove(file.LocalFile)
			}
		}

		return nil
	})
}
