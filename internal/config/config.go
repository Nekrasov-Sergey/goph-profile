// Package config содержит конфигурацию клиента и сервера.
package config

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"go.uber.org/multierr"
)

// openConfigSafe открывает файл конфигурации через os.Root,
// предотвращая path traversal за пределы каталога файла.
func openConfigSafe(configPath string) (_ *os.File, err error) {
	dir := filepath.Dir(configPath)
	name := filepath.Base(configPath)

	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, errors.Wrap(err, "не удалось открыть корневой каталог конфигурации")
	}
	defer multierr.AppendInvoke(&err, multierr.Close(root))

	f, err := root.Open(name)
	if err != nil {
		return nil, errors.Wrap(err, "не удалось открыть файл конфигурации")
	}

	return f, nil
}
