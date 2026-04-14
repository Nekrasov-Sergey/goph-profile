package types

// MIMEType представляет MIME-тип изображения.
type MIMEType string

const (
	// MIMETypeJPEG — MIME-тип image/jpeg.
	MIMETypeJPEG MIMEType = "image/jpeg"
	// MIMETypePNG — MIME-тип image/png.
	MIMETypePNG MIMEType = "image/png"
	// MIMETypeWebP — MIME-тип image/webp.
	MIMETypeWebP MIMEType = "image/webp"
)
