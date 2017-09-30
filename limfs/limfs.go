// Package limfs provides a functionality of static file server
// with enforces a limit on maximum number of downloads
package limfs

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// LimitedHandler gives ability to limit number of downloads
// for static file server
type LimitedHandler struct {
	dirPath         string
	fs              http.Handler
	filesAccessed   map[string]*int
	maxTimesPerFile int
}

// New creates a file server handler with pre-defined max number of downloads
func New(fileDir string, maxDownloadsPerFile int) *LimitedHandler {
	return &LimitedHandler{
		dirPath:         fileDir,
		fs:              http.FileServer(http.Dir(fileDir)),
		filesAccessed:   make(map[string]*int),
		maxTimesPerFile: maxDownloadsPerFile,
	}
}

func respondWithMessage(w http.ResponseWriter, msg string, httpStatus int) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(httpStatus)
	_, err := w.Write([]byte(fmt.Sprintf("<center><h1>%s</h1></center>", msg)))
	if err != nil {
		log.Printf("Error sending HTML response: %v", err)
	}
}

func (ltdHandler *LimitedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Print(r)
	v := ltdHandler.timesAccessed(r.RequestURI)
	if v == nil {
		respondWithMessage(w, "Not Found!", http.StatusNotFound)
		return
	}
	if *v >= ltdHandler.maxTimesPerFile {
		respondWithMessage(w, "File is no longer available for download", http.StatusGone)
		return
	}
	if r.Method == http.MethodGet {
		(*v)++
		w.Header().Add("X-Times-Remaining", strconv.Itoa(ltdHandler.maxTimesPerFile-*ltdHandler.timesAccessed(r.RequestURI)))
		ltdHandler.fs.ServeHTTP(w, r)
	}

}

func (ltdHandler *LimitedHandler) timesAccessed(path string) *int {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	requestedFile := fmt.Sprintf("%s%s", ltdHandler.dirPath, path)
	if info, err := os.Stat(requestedFile); !os.IsNotExist(err) && !info.IsDir() {
		v, ok := ltdHandler.filesAccessed[requestedFile]
		if !ok {
			var init = 0
			ltdHandler.filesAccessed[requestedFile] = &init
			v = &init
		}
		return v
	}
	return nil
}

//DisallowAccess blocks access to speficied file
func (ltdHandler *LimitedHandler) DisallowAccess(path string) bool {
	v := ltdHandler.timesAccessed(path)
	if v != nil {
		*v = ltdHandler.maxTimesPerFile
		return true
	}
	return false
}
