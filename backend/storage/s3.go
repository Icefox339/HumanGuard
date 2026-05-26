package storage

import (
	"io"
	"os"
	"path/filepath"
)

type S3Client interface {
	Save(path string, reader io.Reader) (int64, error)
	Get(path string) (io.ReadCloser, error)
	Delete(path string) error
}

type LocalS3 struct {
	baseDir string
}

func NewLocalS3(baseDir string) (*LocalS3, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, err
	}
	return &LocalS3{baseDir: baseDir}, nil
}

func (l *LocalS3) Save(path string, reader io.Reader) (int64, error) {
	fullPath := filepath.Join(l.baseDir, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return 0, err
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	return io.Copy(file, reader)
}

func (l *LocalS3) Get(path string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(l.baseDir, path))
}

func (l *LocalS3) Delete(path string) error {
	return os.Remove(filepath.Join(l.baseDir, path))
}
