package types

// ImageFormat представляет формат изображения.
type ImageFormat string

const (
	// ImageFormatJPEG — формат JPEG.
	ImageFormatJPEG ImageFormat = "jpeg"
	// ImageFormatPNG — формат PNG.
	ImageFormatPNG ImageFormat = "png"
	// ImageFormatWebP — формат WebP.
	ImageFormatWebP ImageFormat = "webp"
)
