package contextkey

type contextKey string

const (
	UserCtxKey      contextKey = "user"
	SessionIDCtxKey contextKey = "sessionID"
)
