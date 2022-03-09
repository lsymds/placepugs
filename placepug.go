package main

import (
	"bytes"
	"image"
	"image/jpeg"
	"io/fs"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
)

var files []fs.FileInfo

// main is the main entry point to the application, booting the HTTP server used to serve responses
func main() {
	var err error

	files, err = ioutil.ReadDir("images")
	if err != nil || len(files) == 0 {
		log.Fatalf("err: images directory not present or empty")
	}

	r := mux.NewRouter()
	r.HandleFunc("/", handleIndex).Methods("GET")
	r.HandleFunc("/{w:[0-9]+}/{h:[0-9]+}", handleImageRetrieval).Methods("GET")

	panic(http.ListenAndServe(":8482", r))
}

// handleIndex returns the main, user visitable index page of placepugs
func handleIndex(w http.ResponseWriter, r *http.Request) {

}

// handleImageRetrieval handles the retrieval and display of placepugs.
func handleImageRetrieval(rw http.ResponseWriter, r *http.Request) {
	var w string
	var h string
	var nw uint64
	var nh uint64
	var err error

	vars := mux.Vars(r)

	if w = vars["w"]; w == "" {
		badRequest(rw, "err: width not present")
		return
	}

	if h = vars["h"]; h == "" {
		badRequest(rw, "err: height not present")
		return
	}

	if nw, err = strconv.ParseUint(w, 10, 32); err != nil || nw > 2000 {
		badRequest(rw, "err: width must be greater than 0 but less than 2000")
		return
	}

	if nh, err = strconv.ParseUint(h, 10, 32); err != nil || nh > 2000 {
		badRequest(rw, "err: width must be greater than 0 but less than 2000")
		return
	}

	nn := rand.Intn(len(files)-1) + 1
	file, err := ioutil.ReadFile("images/" + files[nn].Name())
	if err != nil {
		internalServerError(rw, "err: failed to open image")
		return
	}

	img, _, err := image.Decode(bytes.NewReader(file))
	if err != nil {
		internalServerError(rw, "err: failed to decode image")
		return
	}

	rsi := resize.Resize(uint(nw), uint(nh), img, resize.Bicubic)
	rw.Header().Set("Content-Type", "image/jpeg")

	if err = jpeg.Encode(rw, rsi, nil); err != nil {
		internalServerError(rw, "err: failed to encode response")
	}
}

// badRequest writes a bad request response with an error to the response writer
func badRequest(rw http.ResponseWriter, err string) {
	rw.WriteHeader(http.StatusBadRequest)
	rw.Write([]byte(err))
}

// internalServerError writes an internal server error response with an error to the response writer
func internalServerError(rw http.ResponseWriter, err string) {
	rw.WriteHeader(http.StatusInternalServerError)
	rw.Write([]byte(err))
}
