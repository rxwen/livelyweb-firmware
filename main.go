package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
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
	configFile, err := os.Open(CONFIG_FILE_NAME)
	if err != nil {
		log.Error("failed to open " + CONFIG_FILE_NAME)
		return versions
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	if err = jsonParser.Decode(&versions); err != nil {
		log.Error("error parsing config file " + err.Error())
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
		log.Info("Check newer version against " + version + " hwversion: " + hwversion)
		newerVersion := checkNewerVersionFor(version)
		if newerVersion != nil {
			log.Info("append version " + newerVersion.Version)
			results = append(results, *newerVersion)
		}
		if len(results) < 1 {
			log.Warn("no newer version available now")
		} else {
			results[len(results)-1].Path = ""
		}
	} else {
		for index, ele := range getAvailableVersions() {
			log.Info("append " + ele.Version)
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
	log.Info("download package for " + version)
	vi := findVersion(version)

	if vi == nil {
		log.Warn("can't find version info for " + version)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	log.Info("check local file: " + vi.Path)
	if _, err := os.Stat(vi.Path); os.IsNotExist(err) {
		log.Error("local file: " + vi.Path + " doesn't exist")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	reader, err := os.Open(vi.Path)
	if err != nil {
		log.Error("failed to open " + CONFIG_FILE_NAME)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	defer reader.Close()
	fi, _ := reader.Stat()
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(vi.Path))
	w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
	w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
	io.Copy(w, reader)
}
