package common

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io/ioutil"
)

func Decode(data string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	r, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return "", err
	}

	b2, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}

	return string(b2), nil
}

func Encode(data string) (string, error) {
	var buf bytes.Buffer

	w, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return "", err
	}

	if _, err = w.Write([]byte(data)); err != nil {
		return "", err
	}

	w.Close()

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}
