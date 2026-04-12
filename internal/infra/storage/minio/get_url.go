// Package minio реализует хранилище файлов на базе S3.
package minio

// GetURL возвращает URL для доступа к файлу.
func (s *MinIO) GetURL(key string) string {
	return s.client.EndpointURL().String() + "/" + s.bucket + "/" + key
}
