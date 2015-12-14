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
const VAR_NAME_HARDWAREVERSION = "hwversion"
const CONFIG_FILE_NAME = "package.json"

type VersionInfo struct {
	Version  string `json:"version"`
	Path     string `json:"path,omitempty"` // this field doesn't need to be returned to client
	Checksum string `json:"checksum"`
}

func getAvailableVersions() []VersionInfo {
	var versions []VersionInfo
	//versions := make([]VersionInfo, 0)
	configFile, err := os.Open(CONFIG_FILE_NAME)
	if err != nil {
		fmt.Println("failed to open " + CONFIG_FILE_NAME)
		return versions
	}

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&versions); err != nil {
		fmt.Println("parsing config file " + err.Error())
		return versions
	}
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
	version := r.FormValue(VAR_NAME_VERSION)
	hwversion := r.FormValue(VAR_NAME_HARDWAREVERSION)
	results := make([]VersionInfo, 0)
	if len(version) > 0 {
		fmt.Println("Check newer version against " + version + " hwversion: " + hwversion)
		newerVersion := checkNewerVersionFor(version)
		if newerVersion != nil {
			fmt.Println("append " + newerVersion.Version)
			results = append(results, *newerVersion)
		}
		if len(results) < 1 {
			fmt.Println("no newer version available now")
		} else {
			results[len(results)-1].Path = ""
		}
	} else {
		for index, ele := range getAvailableVersions() {
			fmt.Println("append " + ele.Version)
			results = append(results, ele)
			results[index].Path = ""
		}
	}
	json.NewEncoder(w).Encode(results)
}

func checkNewerVersionFor(version string) *VersionInfo {
	var result *VersionInfo = nil
	for _, ele := range getAvailableVersions() {
		if strings.Compare(ele.Version, version) > 0 {
			version = ele.Version
			result = &ele
		}
	}
	return result
}

func findVersion(version string) *VersionInfo {
	for _, ele := range getAvailableVersions() {
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
	vi.Path = ""
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
