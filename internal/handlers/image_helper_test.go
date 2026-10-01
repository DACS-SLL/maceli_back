package handlers

import (
	"bytes"
	"context"
	"github.com/gin-gonic/gin"
	"image"
	"image/png"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"testing"
)

type testUploader struct{ called bool }

func (u *testUploader) Upload(_ context.Context, _ io.Reader, _ string) (string, error) {
	u.called = true
	return "https://res.cloudinary.com/test/image/upload/test.png", nil
}

func TestUploadInspectsContent(t *testing.T) {
	var pngData bytes.Buffer
	if err := png.Encode(&pngData, image.NewRGBA(image.Rect(0, 0, 10, 10))); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		data []byte
		want bool
	}{
		{"valid.png", pngData.Bytes(), true},
		{"fake.png", []byte("<script>alert('x')</script>"), false},
		{"wrong.jpg", pngData.Bytes(), false},
		{"vector.svg", []byte("<svg></svg>"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)
			part, err := writer.CreateFormFile("imagen", tc.name)
			if err != nil {
				t.Fatal(err)
			}
			_, _ = part.Write(tc.data)
			_ = writer.Close()
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest("POST", "/upload", &body)
			c.Request.Header.Set("Content-Type", writer.FormDataContentType())
			uploader := &testUploader{}
			_, saved, _, _ := saveUploadedImage(c, uploader)
			if saved != tc.want || uploader.called != tc.want {
				t.Fatalf("saved=%v called=%v", saved, uploader.called)
			}
		})
	}
}
