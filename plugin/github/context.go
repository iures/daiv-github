package github

import (
	"context"
	"time"
)

// ContextKey is a type for context keys
type ContextKey string

const (
	// ContextKeyUsername is the key for the username in the context
	ContextKeyUsername ContextKey = "username"
	
	// ContextKeyOrganization is the key for the organization in the context
	ContextKeyOrganization ContextKey = "organization"
	
	// ContextKeyRepository is the key for the repository in the context
	ContextKeyRepository ContextKey = "repository"
	
	// ContextKeyTimeRange is the key for the time range in the context
	ContextKeyTimeRange ContextKey = "timeRange"
	
	// ContextKeyRequestID is the key for the request ID in the context
	ContextKeyRequestID ContextKey = "requestID"
)

// NewContextWithTimeout creates a new context with a timeout
func NewContextWithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}

// WithUsername adds a username to the context
func WithUsername(ctx context.Context, username string) context.Context {
	return context.WithValue(ctx, ContextKeyUsername, username)
}

// GetUsername gets the username from the context
func GetUsername(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(ContextKeyUsername).(string)
	return username, ok
}

// WithOrganization adds an organization to the context
func WithOrganization(ctx context.Context, organization string) context.Context {
	return context.WithValue(ctx, ContextKeyOrganization, organization)
}

// GetOrganization gets the organization from the context
func GetOrganization(ctx context.Context) (string, bool) {
	organization, ok := ctx.Value(ContextKeyOrganization).(string)
	return organization, ok
}

// WithRepository adds a repository to the context
func WithRepository(ctx context.Context, repository string) context.Context {
	return context.WithValue(ctx, ContextKeyRepository, repository)
}

// GetRepository gets the repository from the context
func GetRepository(ctx context.Context) (string, bool) {
	repository, ok := ctx.Value(ContextKeyRepository).(string)
	return repository, ok
}

// WithTimeRange adds a time range to the context
func WithTimeRange(ctx context.Context, timeRange TimeRange) context.Context {
	return context.WithValue(ctx, ContextKeyTimeRange, timeRange)
}

// GetTimeRange gets the time range from the context
func GetTimeRange(ctx context.Context) (TimeRange, bool) {
	timeRange, ok := ctx.Value(ContextKeyTimeRange).(TimeRange)
	return timeRange, ok
}

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, ContextKeyRequestID, requestID)
}

// GetRequestID gets the request ID from the context
func GetRequestID(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(ContextKeyRequestID).(string)
	return requestID, ok
}

// NewRequestContext creates a new context with all the request information
func NewRequestContext(username, organization, repository string, timeRange TimeRange) context.Context {
	ctx := context.Background()
	ctx = WithUsername(ctx, username)
	ctx = WithOrganization(ctx, organization)
	ctx = WithRepository(ctx, repository)
	ctx = WithTimeRange(ctx, timeRange)
	return ctx
} 
