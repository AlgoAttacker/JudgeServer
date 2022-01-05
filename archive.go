package main

import (
	"archive/tar"
	"bytes"
	"io/ioutil"
)

func createTarIncludesSource(name string, content []byte) (*bytes.Reader, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	hdr := &tar.Header{
		Name: name,
		Size: int64(len(content)),
	}

	if err := tw.WriteHeader(hdr); err != nil {
		return nil, err
	}

	if _, err := tw.Write(content); err != nil {
		return nil, err
	}

	defer tw.Close()

	return bytes.NewReader(buf.Bytes()), nil
}

func createTarIncludesFolder(dir string) (*bytes.Reader, error) {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		hdr := &tar.Header{
			Name: file.Name(),
			Size: file.Size(),
		}

		if err := tw.WriteHeader(hdr); err != nil {
			return nil, err
		}

		file, err := ioutil.ReadFile(dir + "/" + file.Name())
		if err != nil {
			return nil, err
		}

		if _, err := tw.Write(file); err != nil {
			return nil, err
		}
	}

	defer tw.Close()

	return bytes.NewReader(buf.Bytes()), nil
}
