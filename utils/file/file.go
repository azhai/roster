package file

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
)

const (
	FILE_MODE = 0666
	DIR_MODE  = 0777
)

func CreateFile(path string, mode os.FileMode) (fp *os.File, err error) {
	// create dirs if file not exists
	if dir := filepath.Dir(path); dir != "." {
		err = os.MkdirAll(dir, DIR_MODE)
	}
	if err == nil {
		fp, err = os.Create(path)
		if err == nil && mode != 0666 {
			fp.Chmod(mode)
		}
	}
	return
}

func OpenFile(path string) (fp *os.File, size int64, err error) {
	var info os.FileInfo
	// detect if file exists
	info, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			fp, err = CreateFile(path, FILE_MODE)
		}
		return
	}
	size = info.Size()
	if size > 0 {
		fp, err = os.Open(path)
	} else {
		fp, err = os.OpenFile(path, os.O_RDWR, FILE_MODE)
	}
	return
}

func ReadFileLines(path string) ([]string, error) {
	fp, _, err := OpenFile(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()
	return ReadLines(fp)
}

func ReadLines(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}
