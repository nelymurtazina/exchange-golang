package requestid

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

type contextKey struct{}
var ContextKey = contextKey{}

const (
	MetadataKey = "x-request-id"
)

func FromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ContextKey).(string); ok {
		return v
	}
	return ""
}

func WithContext(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ContextKey, requestID)
}

// ExtractOrGenerate:
// Сначала пытается взять x-request-id из входящей metadata
// Валидирует его
// Если нет или невалидный — генерирует новый UUID
func ExtractOrGenerate(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return generate()
	}

	values := md.Get(MetadataKey)
	if len(values) == 0 || values[0] == "" {
		return generate()
	}

	clientID := values[0]

	if len(clientID) > 64 || strings.ContainsAny(clientID, " ,;") {
		return generate()
	}

	return clientID
}

func generate() string {
	return uuid.New().String()
}