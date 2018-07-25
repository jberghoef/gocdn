package main

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/boltdb/bolt"
)

/* File
The file model
====================================================================== */
type File struct {
	Reference  string      `json:"reference"`
	URL        string      `json:"url"`
	LocalFile  string      `json:"localfile"`
	Timestamp  int64       `json:"timestamp"`
	MaxAge     int64       `json:"maxage"`
	Revalidate bool        `json:"revalidate"`
	ETAG       string      `json:"etag"`
	Header     http.Header `json:"header"`
}

/* Register
Registers metadata of the retrieved file to the database.
====================================================================== */
func (f *File) Register() {
	fileJSON, _ := json.Marshal(f)

	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Cache"))
		err := b.Put([]byte(f.Reference), fileJSON)
		return err
	})
}

/* Remove
Removes file and it's database entry.
====================================================================== */
func (f *File) Remove() {
	db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Cache"))
		b.Delete([]byte(f.Reference))
		return nil
	})

	os.Remove(f.LocalFile)
}

/* Retrieve
Retrieve file metadata from storage.
====================================================================== */
func (f *File) Retrieve(originURL string) {
	refHash := createHash(originURL)

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Cache"))
		v := b.Get([]byte(refHash))

		json.Unmarshal(v, &f)
		return nil
	})
}
