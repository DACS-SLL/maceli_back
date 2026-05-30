package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LocalUploader struct {
	Directory string
	PublicURL string
}

func NewLocalUploader(directory string, publicURL string) *LocalUploader {
	return &LocalUploader{
		Directory: directory,
		PublicURL: strings.TrimRight(publicURL, "/"),
	}
}

func (u *LocalUploader) Upload(_ context.Context, file io.Reader, originalFilename string) (string, error) {
	if err := os.MkdirAll(u.Directory, os.ModePerm); err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(originalFilename))
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	destination := filepath.Join(u.Directory, filename)

	dst, err := os.Create(destination)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return u.PublicURL + "/" + filename, nil
}
