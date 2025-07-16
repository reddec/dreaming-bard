package common

//go:generate go run github.com/abice/go-enum@v0.6.1 -sql --marshal --values

// ENUM(user,assistant,tool_call,tool_result)
type Role string

// ENUM(write,summary,enhance,plan)
type Purpose string
