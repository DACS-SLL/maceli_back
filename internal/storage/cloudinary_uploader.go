package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryUploader struct {
	client *cloudinary.Cloudinary
	folder string
}

func NewCloudinaryUploader(cloudName string, apiKey string, apiSecret string, folder string) (*CloudinaryUploader, error) {
	client, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, err
	}

	return &CloudinaryUploader{
		client: client,
		folder: folder,
	}, nil
}

func (u *CloudinaryUploader) Upload(ctx context.Context, file io.Reader, originalFilename string) (string, error) {
	publicID := buildPublicID(originalFilename)

	result, err := u.client.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder:   u.folder,
		PublicID: publicID,
	})
	if err != nil {
		return "", err
	}

	return result.SecureURL, nil
}

func buildPublicID(originalFilename string) string {
	name := strings.TrimSuffix(filepath.Base(originalFilename), filepath.Ext(originalFilename))
	name = strings.ToLower(name)
	name = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")

	if name == "" {
		name = "imagen"
	}

	return fmt.Sprintf("%s-%d", name, time.Now().UnixNano())
}
