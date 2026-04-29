package metrics

import "go.opentelemetry.io/otel/attribute"

// --- HTTP attribute helpers ---

func AttrHTTPMethod(method string) attribute.KeyValue {
	return attribute.String("http.request.method", method)
}

func AttrHTTPRoute(route string) attribute.KeyValue {
	return attribute.String("http.route", route)
}

func AttrHTTPStatus(code int) attribute.KeyValue {
	return attribute.Int("http.response.status_code", code)
}

// --- Business metric attribute helpers ---

func AttrMimeType(mt string) attribute.KeyValue {
	return attribute.String("avatar.mime_type", mt)
}

func AttrBusinessStatus(s string) attribute.KeyValue {
	return attribute.String("avatar.status", s)
}

func AttrAvatarOperation(op string) attribute.KeyValue {
	return attribute.String("avatar.operation", op)
}

// --- DB attribute helpers ---

func AttrDBOperation(op string) attribute.KeyValue {
	return attribute.String("db.operation", op)
}

func AttrDBStatus(s string) attribute.KeyValue {
	return attribute.String("db.status", s)
}

// --- S3 attribute helpers ---

func AttrS3Operation(op string) attribute.KeyValue {
	return attribute.String("s3.operation", op)
}

func AttrS3Status(s string) attribute.KeyValue {
	return attribute.String("s3.status", s)
}

// --- Kafka attribute helpers ---

func AttrMessagingDirection(dir string) attribute.KeyValue {
	return attribute.String("messaging.direction", dir)
}

func AttrMessagingOp(op string) attribute.KeyValue {
	return attribute.String("messaging.operation", op)
}
