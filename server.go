// Package main provides a simple http server that can store a file and serve it
// for download once before deleting it.
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/nu7hatch/gouuid"
)

// main parses a single port flag, sets up the routes, and starts the server on
// the env-defined port or 1111 by default.
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "1111"
	}
	http.HandleFunc("/new", HostFile)
	http.HandleFunc("/", ServeFile)
	http.ListenAndServe(":"+port, nil)
}

// HostFile expects a POST with a file. It grabs the file's contents, generates
// an id, saves the file locally as that id, then returns the id.
func HostFile(res http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		fourohfour(res)
		return
	}

	contents := getFileContents(req)

	id, err := uuid.NewV4()
	if err != nil {
		fourohfour(res)
		log.Fatal(err)
	}

	file, err := createFile(id.String())
	if err != nil {
		fourohfour(res)
		log.Fatal(err)
	}

	file.Write(contents)

	fmt.Fprint(res, id.String())
}

// ServeFile catches all other requests. It expects a "file" param in the
// request body, specifying an id. It searches for a file named with that id,
// and if it exists, serves that file then deletes it. If not, 404.
func ServeFile(res http.ResponseWriter, req *http.Request) {
	dirname, err := os.Getwd()
	if err != nil {
		fourohfour(res)
		log.Fatal(err)
	}

	fPath := filepath.Join(dirname, "files", req.URL.String()+".tar.gz")
	content, err := ioutil.ReadFile(fPath)
	if err != nil {
		fourohfour(res)
		return
	}

	fmt.Fprint(res, string(content))

	if err = os.Remove(fPath); err != nil {
		log.Fatal(err)
	}
}

// fourohfour writes a 404 response to a passed in http.ResponseWriter.
func fourohfour(res http.ResponseWriter) {
	res.WriteHeader(404)
	fmt.Fprint(res, "not found")
}

// getFileContents reads the given request body and returns it as a byte slice.
func getFileContents(req *http.Request) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)
	return buf.Bytes()
}

// params is just a struct used to return from the getFilename function below.
type params struct {
	File string
}

// createFile creates a blank file using the given name at ./files/NAME.tar.
func createFile(id string) (file *os.File, err error) {
	dirname, err := os.Getwd()
	if err != nil {
		return
	}

	file, err = os.Create(filepath.Join(dirname, "files", id+".tar.gz"))
	if err != nil {
		log.Fatal(err)
	}

	return
}
