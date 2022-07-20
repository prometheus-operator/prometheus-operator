package operator

import (
	"bytes"
	"compress/gzip"
	"io"
	"strings"
)

func GzipConfig(buf *bytes.Buffer, conf []byte) error {
	w := gzip.NewWriter(buf)
	defer w.Close()
	if _, err := w.Write(conf); err != nil {
		return err
	}
	return nil
}

func GunzipConfig(b []byte) (string, error) {
	buf := bytes.NewBuffer(b)
	reader, err := gzip.NewReader(buf)
	if err != nil {
		return "", err
	}
	uncompressed := new(strings.Builder)
	_, err = io.Copy(uncompressed, reader)
	return uncompressed.String(), nil
}
