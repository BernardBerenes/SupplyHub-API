package products

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"mime/multipart"
	"net/textproto"
	"testing"
)

func newFileHeader(t *testing.T, filename string, data []byte) *multipart.FileHeader {
	t.Helper()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": []string{`form-data; name="photo"; filename="` + filename + `"`},
	})
	if err != nil {
		t.Fatalf("failed to create part: %v", err)
	}
	if _, err := part.Write(data); err != nil {
		t.Fatalf("failed to write part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("failed to close writer: %v", err)
	}

	reader := multipart.NewReader(body, writer.Boundary())
	form, err := reader.ReadForm(int64(len(data)) + 1024)
	if err != nil {
		t.Fatalf("failed to read form: %v", err)
	}

	return form.File["photo"][0]
}

func pngBytes(t *testing.T) []byte {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.White)

	buf := &bytes.Buffer{}
	if err := png.Encode(buf, img); err != nil {
		t.Fatalf("failed to encode png: %v", err)
	}

	return buf.Bytes()
}

func TestParsePhoto_ValidPNG(t *testing.T) {
	fh := newFileHeader(t, "photo.png", pngBytes(t))

	photo, err := parsePhoto(fh)
	if err != nil {
		t.Fatalf("expected valid photo, got error: %v", err)
	}

	if photo.Extension != ".png" {
		t.Fatalf("expected .png extension, got %s", photo.Extension)
	}
	if photo.ContentType != "image/png" {
		t.Fatalf("expected image/png content type, got %s", photo.ContentType)
	}
}

func TestParsePhoto_RejectsNonImage(t *testing.T) {
	fh := newFileHeader(t, "photo.png", []byte("not an image, just declaring itself as one"))

	if _, err := parsePhoto(fh); err == nil {
		t.Fatal("expected error for non-image content, got none")
	}
}

func TestParsePhoto_RejectsOversized(t *testing.T) {
	fh := newFileHeader(t, "photo.png", pngBytes(t))
	fh.Size = MAX_PHOTO_SIZE_BYTES + 1

	if _, err := parsePhoto(fh); err == nil {
		t.Fatal("expected error for oversized photo, got none")
	}
}
