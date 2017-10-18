package limfs

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setup(t *testing.T) (name string) {
	dir, err := ioutil.TempDir("", "static_files")
	if err != nil {
		t.Fatalf("cannot create temp directory for file server: %v", err)
	}
	return dir
}

func TestLimit(t *testing.T) {
	dir := setup(t)
	defer os.RemoveAll(dir)

	content := []byte("this file will be served from root")
	existingFileName := "info.txt"
	tmpfn := filepath.Join(dir, existingFileName)
	if err := ioutil.WriteFile(tmpfn, content, 0666); err != nil {
		t.Fatalf("cannot create sample file in temp directory: %v", err)
	}

	subdirName := "misc"
	nestedDir := filepath.Join(dir, subdirName)
	if err := os.Mkdir(nestedDir, 0700); err != nil {
		t.Fatalf("cannot create nested dir inside file server directory: %v", err)
	}

	maxTimes := 2
	handler := New(dir, maxTimes)

	nonExistent := "non-existing.file"
	times := handler.timesAccessed(nonExistent)
	if times != nil {
		t.Error("Non-existing file should not have associated access times counter")
	}
	disallowResult := handler.DisallowAccess(nonExistent)
	times = handler.timesAccessed(nonExistent)
	if disallowResult || times != nil {
		t.Error("Disallow access on non-existing file should return false and not change access the counter")
	}

	times = handler.timesAccessed(subdirName)
	if times != nil {
		t.Error("Existing sub directory should not have associated access times counter")
	}

	times = handler.timesAccessed(existingFileName)
	if times == nil || *times != 0 {
		t.Error("Existing file should be initialized with zero counter")
	}

	disallowResult = handler.DisallowAccess(existingFileName)
	times = handler.timesAccessed(existingFileName)
	if !disallowResult || times == nil || *times != maxTimes {
		t.Error("Disallow access method should set access times to max allowed")
	}

	subdirfn := "info_from_subdir.txt"
	content = []byte("this file will be served from subdir")
	fn := filepath.Join(nestedDir, subdirfn)
	if err := ioutil.WriteFile(fn, content, 0666); err != nil {
		t.Fatalf("cannot create sample file in sub directory directory: %v", err)
	}
	times = handler.timesAccessed(fmt.Sprintf("%s/%s", subdirName, subdirfn))
	if times == nil || *times != 0 {
		t.Error("Existing file from subdirectory should be initialized with zero counter")
	}
}

func TestResponse(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	log.SetFlags(0)
	dir := setup(t)
	defer os.RemoveAll(dir)

	content := []byte("this is sample file")
	existingFileName := "info.txt"
	tmpfn := filepath.Join(dir, existingFileName)
	if err := ioutil.WriteFile(tmpfn, content, 0666); err != nil {
		t.Fatalf("cannot create sample file in temp directory: %v", err)
	}
	maxTimes := 3
	handler := New(dir, maxTimes)

	for i := 1; i <= maxTimes; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/"+existingFileName, nil)
		handler.ServeHTTP(rec, req)
		if rec.Result().StatusCode != http.StatusOK {
			t.Errorf("Expected status OK on existing file, got %d", rec.Result().StatusCode)
		}
		remaining := rec.Result().Header.Get("X-Times-Remaining")
		expected := fmt.Sprintf("%d", maxTimes-i)
		if strings.Compare(expected, remaining) != 0 {
			t.Errorf("Expected proper times remaining reported: got %s, expected %s", remaining, expected)
		}
		if string(rec.Body.Bytes()) != string(content) {
			t.Errorf("Unexpected content")
		}

	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/"+existingFileName, nil))
	if rec.Result().StatusCode != http.StatusGone {
		t.Errorf("Expected status Gone on existing file with access time exceeded, got %d", rec.Result().StatusCode)
	}
	if !strings.Contains(string(rec.Body.Bytes()), "File is no longer available for download") {
		t.Errorf("Expected proper message for file with expired download")
	}

	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/non-existing.file", nil))
	if rec.Result().StatusCode != http.StatusNotFound {
		t.Errorf("Expected status Not Found on non existing file, got %d", rec.Result().StatusCode)
	}
	if !strings.Contains(string(rec.Body.Bytes()), "Not Found") {
		t.Errorf("Expected proper message for non existing file")
	}

}
