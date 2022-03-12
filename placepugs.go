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
		return
	}

	r := mux.NewRouter()
	r.HandleFunc("/{w:[0-9]+}/{h:[0-9]+}", handleImageRetrieval).Methods("GET")

	log.Printf("running placepugs on port %v", 8482)
	panic(http.ListenAndServe(":8482", r))
}

// handleImageRetrieval handles the retrieval and display of placepugs.
func handleImageRetrieval(rw http.ResponseWriter, r *http.Request) {
	var wp string
	var hp string
	var w uint64
	var h uint64
	var err error

	vars := mux.Vars(r)

	if wp = vars["w"]; wp == "" {
		badRequest(rw, "err: width not present")
		return
	}

	if hp = vars["h"]; hp == "" {
		badRequest(rw, "err: height not present")
		return
	}

	if w, err = strconv.ParseUint(wp, 10, 32); err != nil || w > 2000 {
		badRequest(rw, "err: width must be greater than 0 but less than 2000")
		return
	}

	if h, err = strconv.ParseUint(hp, 10, 32); err != nil || h > 2000 {
		badRequest(rw, "err: width must be greater than 0 but less than 2000")
		return
	}

	log.Printf("retrieving image of w:%v h:%v", w, h)

	pug := pugFromSize(w, h)

	file, err := ioutil.ReadFile("images/" + pug.File)
	if err != nil {
		internalServerError(rw, "err: failed to open image")
		return
	}

	img, _, err := image.Decode(bytes.NewReader(file))
	if err != nil {
		internalServerError(rw, "err: failed to decode image")
		return
	}

	resizedImg := resize.Resize(uint(w), uint(h), img, resize.Bicubic)
	rw.Header().Set("Content-Type", "image/jpeg")
	rw.Header().Set("X-Original-Link", pug.Link)

	if err = jpeg.Encode(rw, resizedImg, nil); err != nil {
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
	log.Printf("internal server error: %v", err)
	rw.WriteHeader(http.StatusInternalServerError)
	rw.Write([]byte(err))
}

// fileFromSize finds a file that matches the width and height closely, or alternative one that is as close as possible
// to the aspect ratio of the request
func pugFromSize(w uint64, h uint64) *pug {
	var selectedPugs []pug

	// any exact matches means game on
	for _, p := range pugs {
		if p.Height == h && p.Width == w {
			selectedPugs = append(selectedPugs, p)
			break
		}
	}

	// if any matching the aspect ratio of the request
	if len(selectedPugs) == 0 {
	}

	// else, find based on portrait vs landscape
	if len(selectedPugs) == 0 {
		var o string

		if w > h {
			o = "landscape"
		} else {
			o = "portrait"
		}

		for _, p := range pugs {
			if p.Orientation == o {
				selectedPugs = append(selectedPugs, p)
			}
		}
	}

	return &selectedPugs[rand.Intn(len(selectedPugs))]
}
