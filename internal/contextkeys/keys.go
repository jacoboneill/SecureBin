package contextkeys

type contextKey string

const (
	UserCtxKey      contextKey = "user"
	SessionIDCtxKey contextKey = "sessionID"
)
