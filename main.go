package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

const VAR_NAME_VERSION = "version"

type VersionInfo struct {
	Version  string `json:"version"`
	Path     string `json:"path"`
	Checksum string `json:"checksum"`
}

var versions = map[string]VersionInfo{
	"v1.0.0": VersionInfo{
		Version: "v1.0.0",
		Path:    "./update_v1.0.0.zip",
	},
	"v1.0.1": VersionInfo{
		Version: "v1.0.1",
		Path:    "./update_v1.0.1.zip",
	},
}

func getAvailableVersions() map[string]VersionInfo {
	return versions
}

func main() {

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	router.HandleFunc("/version/{"+VAR_NAME_VERSION+"}", showVersion)
	router.HandleFunc("/download/{"+VAR_NAME_VERSION+"}", download)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func Index(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(getAvailableVersions())
}

func showVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	version := vars[VAR_NAME_VERSION]
	vi := getAvailableVersions()[version]
	json.NewEncoder(w).Encode(vi)
}

func download(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	version := vars[VAR_NAME_VERSION]
	vi := getAvailableVersions()[version]
	if _, err := os.Stat(vi.Path); os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	reader, _ := os.Open(vi.Path)
	defer reader.Close()
	fi, _ := reader.Stat()
	w.Header().Set("Content-Disposition", "attachment; filename=a.txt")
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
	io.Copy(w, reader)
}
