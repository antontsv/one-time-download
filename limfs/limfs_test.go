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

type file struct {
	name    string
	content string
	path    string
}

type dir struct {
	name string
	path string
}

type fs struct {
	root    dir
	subdirs []dir
	files   []file
}

func setup() (sampleFs *fs, err error) {
	tdir, err := ioutil.TempDir("", "static_files")
	sample := &fs{}
	if err != nil {
		return sample, fmt.Errorf("cannot create temp directory for file server: %v", err)
	}
	sample.root = dir{name: "TEST ROOT", path: tdir}

	content := "this file will be served from root"
	existingFileName := "info.txt"
	tmpfn := filepath.Join(tdir, existingFileName)
	if err := ioutil.WriteFile(tmpfn, []byte(content), 0666); err != nil {
		return sample, fmt.Errorf("cannot create sample file in temp directory: %v", err)
	}
	sample.files = append(sample.files,
		file{name: existingFileName, content: content, path: existingFileName})

	subdirName := "misc"
	nestedDir := filepath.Join(tdir, subdirName)
	if err := os.Mkdir(nestedDir, 0700); err != nil {
		return sample, fmt.Errorf("cannot create nested dir inside file server directory: %v", err)
	}
	sample.subdirs = append(sample.subdirs, dir{name: subdirName, path: nestedDir})

	subdirfn := "info_from_subdir.txt"
	content = "this file will be served from subdir"
	fn := filepath.Join(nestedDir, subdirfn)
	if err := ioutil.WriteFile(fn, []byte(content), 0666); err != nil {
		return sample, fmt.Errorf("cannot create sample file in sub directory directory: %v", err)
	}
	sample.files = append(sample.files,
		file{name: subdirfn, content: content, path: fmt.Sprintf("%s/%s", subdirName, subdirfn)})

	return sample, nil
}

func tearDown(sampleFs *fs) {
	if sampleFs != nil && sampleFs.root.path != "" {
		os.RemoveAll(sampleFs.root.path)
	}
}

func TestLimit(t *testing.T) {
	fs, err := setup()
	defer tearDown(fs)
	if err != nil {
		t.Fatalf("unable to setup for a test: %v", err)
	}

	dir := fs.root.path

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

	times = handler.timesAccessed(fs.subdirs[0].name)
	if times != nil {
		t.Error("Existing sub directory should not have associated access times counter")
	}

	for i := len(fs.files) - 1; i >= 0; i-- {
		times = handler.timesAccessed(fs.files[i].path)
		if times == nil || *times != 0 {
			t.Error("Existing file should be initialized with zero counter")
		}

		disallowResult = handler.DisallowAccess(fs.files[i].path)
		times = handler.timesAccessed(fs.files[i].path)
		if !disallowResult || times == nil || *times != maxTimes {
			t.Error("Disallow access method should set access times to max allowed")
		}
	}

}

func TestResponse(t *testing.T) {
	log.SetOutput(ioutil.Discard)
	log.SetFlags(0)
	fs, err := setup()
	defer tearDown(fs)
	if err != nil {
		t.Fatalf("unable to setup for a test: %v", err)
	}

	dir := fs.root.path

	maxTimes := 3
	handler := New(dir, maxTimes)

	for fi := len(fs.files) - 1; fi >= 0; fi-- {

		existingFileURI := "/" + fs.files[fi].path

		for i := 1; i <= maxTimes; i++ {
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, existingFileURI, nil)
			handler.ServeHTTP(rec, req)
			if rec.Result().StatusCode != http.StatusOK {
				t.Errorf("Expected status OK on existing file, got %d", rec.Result().StatusCode)
			}
			remaining := rec.Result().Header.Get("X-Times-Remaining")
			expected := fmt.Sprintf("%d", maxTimes-i)
			if strings.Compare(expected, remaining) != 0 {
				t.Errorf("Expected proper times remaining reported: got %s, expected %s", remaining, expected)
			}
			if string(rec.Body.Bytes()) != fs.files[fi].content {
				t.Errorf("Unexpected content")
			}

		}
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, existingFileURI, nil))
		if rec.Result().StatusCode != http.StatusGone {
			t.Errorf("Expected status Gone on existing file with access time exceeded, got %d", rec.Result().StatusCode)
		}
		if !strings.Contains(string(rec.Body.Bytes()), "File is no longer available for download") {
			t.Errorf("Expected proper message for file with expired download")
		}
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/non-existing.file", nil))
	if rec.Result().StatusCode != http.StatusNotFound {
		t.Errorf("Expected status Not Found on non existing file, got %d", rec.Result().StatusCode)
	}
	if !strings.Contains(string(rec.Body.Bytes()), "Not Found") {
		t.Errorf("Expected proper message for non existing file")
	}

}
