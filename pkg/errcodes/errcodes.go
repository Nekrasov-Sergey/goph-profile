// Package errcodes содержит коды ошибок приложения.
package errcodes

import (
	"github.com/pkg/errors"
)

var (
	ErrAvatarIDAlreadyExists = errors.New("ID аватара уже занято")
	ErrAvatarNotFound        = errors.New("аватар не найден")
	ErrAccessDenied          = errors.New("доступ запрещён. Можно удалять только свои аватары")
	ErrFileTooLarge          = errors.New("файл слишком большой. Максимум: 10MB")
	ErrInvalidFormat         = errors.New("неверный формат файла. Поддерживаемые форматы: jpeg, png, webp")
)
