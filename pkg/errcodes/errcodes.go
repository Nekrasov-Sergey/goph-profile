// Package errcodes содержит коды ошибок приложения.
package errcodes

import (
	"github.com/pkg/errors"
)

var (
	// ErrAvatarIDAlreadyExists — аватар с таким ID уже существует.
	ErrAvatarIDAlreadyExists = errors.New("ID аватара уже занято")
	// ErrAvatarNotFound — аватар не найден.
	ErrAvatarNotFound = errors.New("аватар не найден")
	// ErrAccessDenied — доступ запрещён.
	ErrAccessDenied = errors.New("доступ запрещён. Можно удалять только свои аватары")
	// ErrFileTooLarge — размер файла превышает допустимый лимит.
	ErrFileTooLarge = errors.New("файл слишком большой. Максимум: 10MB")
	// ErrInvalidFormat — неподдерживаемый формат файла.
	ErrInvalidFormat = errors.New("неверный формат файла. Поддерживаемые форматы: jpeg, png, webp")
	// ErrThumbnailNotReady — миниатюра ещё не готова.
	ErrThumbnailNotReady = errors.New("миниатюра ещё не готова")
	// ErrInvalidSize — недопустимый размер миниатюры.
	ErrInvalidSize = errors.New("неверный размер. Допустимые значения: original, 100x100, 300x300")
)
