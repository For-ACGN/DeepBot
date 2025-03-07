package deepbot

import (
	"bytes"
	"encoding/json"
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
