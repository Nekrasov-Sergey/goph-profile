// Package imageutils содержит утилиты для работы с изображениями.
package imageutils

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"io"

	"github.com/chai2010/webp"
	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/Nekrasov-Sergey/goph-profile/internal/types"
)

// FormatToMimeType преобразует формат изображения в MIME-тип.
func FormatToMimeType(format types.ImageFormat) (types.MIMEType, error) {
	switch format {
	case types.ImageFormatJPEG:
		return types.MIMETypeJPEG, nil
	case types.ImageFormatPNG:
		return types.MIMETypePNG, nil
	case types.ImageFormatWebP:
		return types.MIMETypeWebP, nil
	default:
		return "", errors.Errorf("неизвестный формат изображения: %s", format)
	}
}

// MimeTypeToFormat преобразует MIME-тип в формат изображения.
func MimeTypeToFormat(mimeType types.MIMEType) (types.ImageFormat, error) {
	switch mimeType {
	case types.MIMETypeJPEG:
		return types.ImageFormatJPEG, nil
	case types.MIMETypePNG:
		return types.ImageFormatPNG, nil
	case types.MIMETypeWebP:
		return types.ImageFormatWebP, nil
	default:
		return "", errors.Errorf("неизвестный MIME-тип изображения: %s", mimeType)
	}
}

// Encode кодирует изображение в указанный формат.
func Encode(img image.Image, mimeType types.MIMEType) ([]byte, error) {
	switch mimeType {
	case types.MIMETypeJPEG:
		return encodeJPEG(img)
	case types.MIMETypePNG:
		return encodePNG(img)
	case types.MIMETypeWebP:
		return encodeWebP(img)
	default:
		return nil, errors.Errorf("неизвестный MIME-тип изображения: %s", mimeType)
	}
}

func encodeJPEG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		return nil, errors.Wrap(err, "не удалось кодировать в JPEG")
	}
	return buf.Bytes(), nil
}

func encodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, errors.Wrap(err, "не удалось кодировать в PNG")
	}
	return buf.Bytes(), nil
}

func encodeWebP(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, &webp.Options{Quality: 85}); err != nil {
		return nil, errors.Wrap(err, "не удалось кодировать в WebP")
	}
	return buf.Bytes(), nil
}

func ChangeMimeType(
	reader io.ReadCloser,
	currentMimeType types.MIMEType,
	reqMimeType types.MIMEType,
) (r io.ReadCloser, size int64, err error) {
	defer multierr.AppendInvoke(&err, multierr.Close(reader))

	var data []byte

	if currentMimeType == reqMimeType {
		data, err = io.ReadAll(reader)
		if err != nil {
			return nil, 0, errors.Wrap(err, "не удалось прочитать данные")
		}
	} else {
		img, _, decodeErr := image.Decode(reader)
		if decodeErr != nil {
			return nil, 0, errors.Wrap(decodeErr, "не удалось декодировать изображение")
		}

		data, err = Encode(img, reqMimeType)
		if err != nil {
			return nil, 0, err
		}
	}

	return io.NopCloser(bytes.NewReader(data)), int64(len(data)), nil
}

func ResolveMimeType(currentMimeType types.MIMEType, reqFormat types.ImageFormat) (types.MIMEType, error) {
	if reqFormat == "" {
		return currentMimeType, nil
	}
	return FormatToMimeType(reqFormat)
}
