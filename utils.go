package deepbot

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
)

func jsonEncode(data interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, 4096))
	encoder := json.NewEncoder(buf)
	encoder.SetIndent("", "  ")
	err := encoder.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func isFileExists(path string) (bool, error) {
	file, err := os.Open(path)
	if err == nil {
		_ = file.Close()
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func copyFile(dst, src string) error {
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()
	_, err = io.Copy(dstFile, srcFile)
	return err
}
