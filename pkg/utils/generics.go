// Package utils содержит вспомогательные функции общего назначения.
package utils

// Ptr возвращает указатель на переданное значение.
func Ptr[T any](v T) *T {
	return &v
}

// Deref безопасно разыменовывает указатель.
func Deref[T any](p *T) (v T) {
	if p == nil {
		return v
	}
	return *p
}
