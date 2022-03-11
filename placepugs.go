package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
)

var pugs []pug

// pug represents an individual pug image stored in the catalogue file.
type pug struct {
	File        string `json:"file"`
	Desc        string `json:"desc"`
	Link        string `json:"link"`
	Orientation string `json:"orientation"`
	Width       uint64 `json:"width"`
	Height      uint64 `json:"height"`
}

// main is the main entry point to the application, booting the HTTP server used to serve responses
func main() {
	f, err := ioutil.ReadFile("images/catalogue.json")
	if err != nil {
		log.Fatalf("err: failed to open catalogue file")
		return
	}

	err = json.Unmarshal(f, &pugs)
	if err != nil {
		log.Fatalf("err failed to parse catalogue file: %v", err)
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

	file, err := pugFromSize(nw, nh)
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

// fileFromSize finds a file that matches the width and height closely, or alternative one that is as close as possible
// to the aspect ratio of the request
func pugFromSize(w uint64, h uint64) ([]byte, error) {
	var selectedPug *pug

	// any exact matches means game on
	for _, p := range pugs {
		if p.Height == h && p.Width == w {
			selectedPug = &p
			break
		}
	}

	// if any matching the aspect ratio of the request
	if selectedPug == nil {
	}

	// else, find based on portrait vs landscape
	if selectedPug == nil {
		var pa []pug
		var orientation string

		if w > h {
			orientation = "landscape"
		} else {
			orientation = "portrait"
		}

		for _, p := range pugs {
			if p.Orientation == orientation {
				pa = append(pa, p)
			}
		}

		if len(pa) > 0 {
			selectedPug = &pa[rand.Intn(len(pa))]
		}
	}

	return ioutil.ReadFile("images/" + selectedPug.File)
}
