package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

const VAR_NAME_VERSION = "version"

type VersionInfo struct {
	Version  string `json:"version"`
	Path     string `json:"-"` // this field doesn't need to be returned to client
	Checksum string `json:"checksum"`
}

var versions = []VersionInfo{
	VersionInfo{
		Version: "v1.0.0",
		Path:    "update_v1.0.0.zip",
	},
	VersionInfo{
		Version:  "v1.0.2",
		Path:     "update_v1.0.2.zip",
		Checksum: "c2a0e601af0da61e9015018a044749166a66123a",
	},
}

func getAvailableVersions() []VersionInfo {
	return versions
}

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", listVersion)
	router.HandleFunc("/version/{"+VAR_NAME_VERSION+"}", showVersion)
	router.HandleFunc("/download/{"+VAR_NAME_VERSION+"}", download)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func listVersion(w http.ResponseWriter, r *http.Request) {
	version := r.FormValue("version")
	hwversion := r.FormValue("hwversion")
	if len(version) > 0 {
		fmt.Println("Check newer version against " + version + " hwversion: " + hwversion)
		results := make([]*VersionInfo, 0)
		newerVersion := checkNewerVersionFor(version)
		if newerVersion != nil {
			fmt.Println("append " + newerVersion.Version)
			results = append(results, newerVersion)
		}
		json.NewEncoder(w).Encode(results)
	} else {
		json.NewEncoder(w).Encode(versions)
	}
}

func checkNewerVersionFor(version string) *VersionInfo {
	for _, ele := range versions {
		if strings.Compare(ele.Version, version) > 0 {
			return &ele
		}
	}
	return nil
}

func findVersion(version string) *VersionInfo {
	for _, ele := range versions {
		if strings.EqualFold(ele.Version, version) {
			return &ele
		}
	}
	return nil
}

func showVersion(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	version := vars[VAR_NAME_VERSION]
	vi := findVersion(version)
	json.NewEncoder(w).Encode(vi)
}

func download(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	version := vars[VAR_NAME_VERSION]
	vi := findVersion(version)

	if vi == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if _, err := os.Stat(vi.Path); os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	reader, _ := os.Open(vi.Path)

	defer reader.Close()
	fi, _ := reader.Stat()
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(vi.Path))
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
	io.Copy(w, reader)
}
