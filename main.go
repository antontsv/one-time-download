package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

const fileDir = "files"
const maxTimeToAllowDownload = 1

func getBindPort(defaultPort int) int {
	portStr := os.Getenv("BIND_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 0 || port > 65535 {
		port = defaultPort
		log.Printf("Will use default port %d\n", port)
	}
	return port
}

func main() {
	limitedNumberHandler := &oneTimeStaticHandler{
		fs:            http.FileServer(http.Dir(fileDir)),
		filesAccessed: make(map[string]*int),
	}
	limitedNumberHandler.disallowAccess("README.md")

	http.Handle("/", limitedNumberHandler)
	port := getBindPort(8080)
	log.Printf("Starting server on port %d...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal(err)
	}
}

type oneTimeStaticHandler struct {
	fs            http.Handler
	filesAccessed map[string]*int
}

func respondWithMessage(w http.ResponseWriter, msg string, httpStatus int) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(httpStatus)
	_, err := w.Write([]byte(fmt.Sprintf("<center><h1>%s</h1></center>", msg)))
	if err != nil {
		log.Printf("Error sending HTML response: %v", err)
	}
}

func (oneTime *oneTimeStaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Print(r)
	v := oneTime.timesAccessed(r.RequestURI)
	if v == nil {
		respondWithMessage(w, "Not Found!", http.StatusNotFound)
		return
	}
	if *v >= maxTimeToAllowDownload {
		respondWithMessage(w, "File is no longer available for download", http.StatusGone)
		return
	}
	if r.Method == http.MethodGet {
		(*v)++
		w.Header().Add("X-Download-Times", strconv.Itoa(*oneTime.timesAccessed(r.RequestURI)))
		oneTime.fs.ServeHTTP(w, r)
	}

}

func (oneTime *oneTimeStaticHandler) timesAccessed(path string) *int {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	requestedFile := fmt.Sprintf("%s%s", fileDir, path)
	if _, err := os.Stat(requestedFile); !os.IsNotExist(err) {
		v, ok := oneTime.filesAccessed[requestedFile]
		if !ok {
			var init = 0
			oneTime.filesAccessed[requestedFile] = &init
			v = &init
		}
		return v
	}
	return nil
}

func (oneTime *oneTimeStaticHandler) disallowAccess(path string) bool {
	v := oneTime.timesAccessed(path)
	if v != nil {
		*v = maxTimeToAllowDownload
		return true
	}
	return false
}