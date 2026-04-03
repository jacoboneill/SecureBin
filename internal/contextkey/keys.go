// Package contextkey defines the context keys used to store and retrieve context values
package contextkey

type contextKey string

const (
	UserCtxKey      contextKey = "user"
	SessionIDCtxKey contextKey = "sessionID"
)
