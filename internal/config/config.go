// Package config содержит конфигурацию клиента и сервера.
package config

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// openConfigSafe открывает файл конфигурации через os.Root,
// предотвращая path traversal за пределы каталога файла.
func openConfigSafe(configPath string) (*os.File, error) {
	dir := filepath.Dir(configPath)
	name := filepath.Base(configPath)

	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, errors.Wrap(err, "не удалось открыть корневой каталог конфигурации")
	}
	defer root.Close()

	f, err := root.Open(name)
	if err != nil {
		return nil, errors.Wrap(err, "не удалось открыть файл конфигурации")
	}

	return f, nil
}
