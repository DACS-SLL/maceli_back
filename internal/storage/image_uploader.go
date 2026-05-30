package storage

import (
	"context"
	"io"
)

type ImageUploader interface {
	Upload(ctx context.Context, file io.Reader, originalFilename string) (string, error)
}
