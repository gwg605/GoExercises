package base

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestLoadDataFromLocation(t *testing.T) {
	content := []byte("[{\"key1\":\"val1\"},{\"key2\":\"val2\"}]")
	fileName := TempFileName("common_tests_", ".json")

	err := os.WriteFile(fileName, content, 0664)
	if err != nil {
		t.Fatalf("Unable to create %s file. Error=%v", fileName, err)
	}

	loadedContent, status, err := LoadDataFromLocation("file://"+fileName, Opts{})
	if err != nil {
		t.Fatalf("Unable to load data from location. Error=%v", err)
	}
	if status != http.StatusOK {
		t.Fatalf("Wrong status %d. Expected %d", status, http.StatusOK)
	}

	if len(content) != len(loadedContent) {
		t.Fatalf("loadedContent has wrong length (%d)", len(loadedContent))
	}

	for i, v := range content {
		if loadedContent[i] != v {
			t.Fatalf("loadedContent has wrong value (%v) at #%d", v, i)
		}
	}

	_, status, err = LoadDataFromLocation("file://not.found", Opts{})
	if err == nil {
		t.Fatalf("LoadDataFromLocation(file://not.found) must return error")
	}
	if status != http.StatusNotFound {
		t.Fatalf("Wrong status %d. Expected %d", status, http.StatusNotFound)
	}

	_, status, err = LoadDataFromLocation("\r", Opts{})
	if err == nil {
		t.Fatalf("LoadDataFromLocation(file://not.found) must return error")
	}
	if status != http.StatusBadRequest {
		t.Fatalf("Wrong status %d. Expected %d", status, http.StatusBadRequest)
	}

	_, status, err = LoadDataFromLocation("schema://not.found/path", Opts{"Header": "Value"})
	if err == nil {
		t.Fatalf("LoadDataFromLocation(schema://not.found/path) must return error")
	}
	if status != http.StatusInternalServerError {
		t.Fatalf("Wrong status %d. Expected %d", status, http.StatusInternalServerError)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(content)
	}))
	defer srv.Close()

	downloadedContent, _, err := LoadDataFromLocation(srv.URL, Opts{})
	if err != nil {
		t.Fatalf("Unable to download data from %s location. Error=%v", srv.URL, err)
	}

	if len(content) != len(downloadedContent) {
		t.Fatalf("downloadedContent has wrong length (%d)", len(downloadedContent))
	}

	for i, v := range content {
		if downloadedContent[i] != v {
			t.Fatalf("downloadedContent has wrong value (%v) at #%d", v, i)
		}
	}
}
