package types

import "iter"

// ThumbnailSize представляет размер миниатюры.
type ThumbnailSize string

const (
	// ThumbnailSize100 — миниатюра 100×100.
	ThumbnailSize100 ThumbnailSize = "100x100"
	// ThumbnailSize300 — миниатюра 300×300.
	ThumbnailSize300 ThumbnailSize = "300x300"
)

// ThumbnailDimensions возвращает итератор по размерам миниатюр.
func ThumbnailDimensions() iter.Seq2[ThumbnailSize, uint] {
	return func(yield func(ThumbnailSize, uint) bool) {
		if !yield(ThumbnailSize100, 100) {
			return
		}
		yield(ThumbnailSize300, 300)
	}
}
