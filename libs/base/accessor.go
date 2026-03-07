package base

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

func RequestDataFromLocation(location string, options Opts, writer io.Writer) (int64, int, error) {
	urlstr := ExpandString(location, options)
	u, err := url.Parse(urlstr)
	if err != nil {
		return 0, http.StatusBadRequest, fmt.Errorf("unable to parse '%s' url. Error=%v", urlstr, err)
	}
	if u.Scheme == "file" || u.Scheme == "" {
		infile, err := os.Open(u.Path)
		//log.Printf("requestDataFromLocation() - Use '%s' file. Error=%v", u.Path, err)
		if err != nil {
			return 0, http.StatusNotFound, err
		}
		written, err := io.Copy(writer, infile)
		return written, http.StatusOK, err
	}

	//log.Printf("requestDataFromLocation() - Use '%s' url", urlstr)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", urlstr, nil)
	if err != nil {
		return 0, http.StatusInternalServerError, err
	}

	for opt, value := range options {
		req.Header.Set(opt, value)
		//log.Printf("requestDataFromLocation() - Set '%s'='%s'", opt, value)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("requestDataFromLocation() - Cannot connect to '%s'. Error=%v", urlstr, err)
		return 0, http.StatusInternalServerError, err
	}

	defer func() { _ = resp.Body.Close() }()

	//log.Printf("requestDataFromLocation() - '%s' returns %d", urlstr, resp.StatusCode)

	switch resp.StatusCode {
	case http.StatusOK:
		written, err := io.Copy(writer, resp.Body)
		return written, resp.StatusCode, err
	case http.StatusNotModified:
		return 0, resp.StatusCode, nil
	}

	return 0, resp.StatusCode, fmt.Errorf("server returns %d code", resp.StatusCode)
}

func LoadDataFromLocation(location string, options Opts) ([]byte, int, error) {
	var buffer bytes.Buffer

	_, status, err := RequestDataFromLocation(location, options, &buffer)
	if err != nil {
		return nil, status, err
	}

	return buffer.Bytes(), status, nil
}

func SaveDataFromLocationToFile(location string, options Opts, dest string) (int64, int, error) {
	var status = http.StatusInternalServerError

	file, err := os.Create(dest)
	if err != nil {
		return 0, status, err
	}
	defer func() {
		closeErr := file.Close()
		if closeErr != nil && err == nil {
			err = closeErr
			status = http.StatusInternalServerError
		}
	}()

	written, status, err := RequestDataFromLocation(location, options, file)
	if err != nil {
		return 0, status, err
	}

	return written, status, err
}
