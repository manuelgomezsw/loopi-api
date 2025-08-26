package middleware

import "context"

type contextKey string

const (
	ContextUserID      contextKey = "user_id"
	ContextEmail       contextKey = "email"
	ContextRole        contextKey = "role"
	ContextFranchiseID contextKey = "franchise_id"
)

func GetUserID(ctx context.Context) int {
	if id, ok := ctx.Value(ContextUserID).(int); ok {
		return id
	}
	return 0
}

func GetFranchiseID(ctx context.Context) int {
	if id, ok := ctx.Value(ContextFranchiseID).(int); ok {
		return id
	}
	return 0
}

func GetRole(ctx context.Context) string {
	if role, ok := ctx.Value(ContextRole).(string); ok {
		return role
	}
	return ""
}
