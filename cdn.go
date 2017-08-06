package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"
	"github.com/carlescere/scheduler"
	emoji "gopkg.in/kyokomi/emoji.v1"
	"encoding/json"
)

var protocol = flag.String("protocol", "", "The protocol used by your website.")
var origin = flag.String("origin", "", "The original location of your website.")
var cacheDir string
var db *bolt.DB

/* requestHandler
Decides whether to serve a local file or redirect to the original file.
Will consequently trigger a request to retrieve the file when unavailable.
================================================================================ */
func requestHandler(w http.ResponseWriter, r *http.Request) {
	localFile := filepath.Join(cacheDir + r.URL.Path)
	originURL := *protocol + "://" + *origin + r.URL.Path

	_, err := os.Stat(localFile)
	if err == nil {
		valid := validateCache(originURL)

		if valid {
			sendResponse(w, r, originURL, localFile)

			relPath, _ := filepath.Rel(cacheDir, localFile)
			emoji.Println(":floppy_disk: Serving cached source:", relPath)
		} else {
			go verifyAndRetrieveFile(localFile, originURL)
			redirectResponse(w, r, originURL)
		}
	} else {
		go verifyAndRetrieveFile(localFile, originURL)
		redirectResponse(w, r, originURL)
	}
	return
}

func sendResponse(w http.ResponseWriter, r *http.Request, originURL string, localFile string) {
	refHash := createHash(originURL)

	file := File{}

	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Cache"))
		v := b.Get([]byte(refHash))

		json.Unmarshal(v, &file)
		return nil
	})

	for k := range file.Header {
		w.Header().Set(k, file.Header.Get(k))
	}

	data, err := ioutil.ReadFile(localFile)
	if err != nil {
		fmt.Println("Error while reading file", localFile, "\n-", err)
		return
	}
	fmt.Fprintf(w, "%s", data)
}

func redirectResponse(w http.ResponseWriter, r *http.Request, originURL string) {
	emoji.Println(":surfer: Redirecting to origin:", originURL)
	http.Redirect(w, r, originURL, 302)
}

/* init
Prepares the awesomeness.
================================================================================ */
func init() {
	dir, _ := os.Getwd()
	cacheDir = filepath.Join(dir + "/cache")

	// Check if cache dir exists. Otherwise create it.
	_, err := os.Stat(cacheDir)
	if err != nil {
		if os.Mkdir(cacheDir, 0755) != nil {
			fmt.Println("Unable to create cache directory!")
			return
		}
	}
}

/* main
Spins up the awesomeness.
================================================================================ */
func main() {
	flag.Parse()

	if *protocol == "" || *origin == "" {
		fmt.Println("Please provide a protocol and origin.\nUse '--help' for more information.")
		return
	}

	// Open database
	var err error
	db, err = bolt.Open("cache.db", 0600, nil)
	if err != nil {
		fmt.Println("Error while opening database", "\n-", err)
	}
	defer db.Close()

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Cache"))
		if err != nil {
			fmt.Println("Error while creating Cache bucket.", err)
		}
		return nil
	})

	// Start HTTP server
	http.HandleFunc("/", requestHandler)

	switch{
	case *protocol == "http":
		http.ListenAndServe(":8080", nil)
	case *protocol == "https":
		// Work in progress
		http.ListenAndServeTLS(":8080", "", "", nil)
	default:
		fmt.Println("Protocol not recognized.\nPlease choose between 'http' and 'https'.")
		return
	}


	scheduler.Every(5).Minutes().Run(cleanCache)
}
