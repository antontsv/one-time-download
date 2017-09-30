package limfs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestLimit(t *testing.T) {
	content := []byte("this file will be served from root")
	dir, err := ioutil.TempDir("", "static_files")
	if err != nil {
		t.Fatalf("cannot create temp directory for file server: %v", err)
	}

	defer os.RemoveAll(dir)

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

	handler := New(dir, 1)

	times := handler.timesAccessed("non-existing.file")
	if times != nil {
		t.Error("Non-exising file should not have associated access times counter")
	}

	times = handler.timesAccessed(subdirName)
	if times != nil {
		t.Error("Existing sub directory should not have associated access times counter")
	}

	times = handler.timesAccessed(existingFileName)
	if times == nil || *times != 0 {
		t.Error("Existing file should be initialized with zero counter")
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
