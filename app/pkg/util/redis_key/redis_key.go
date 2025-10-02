package rediskey

import "fmt"

func LoginTokenKey(token string) string {
	return fmt.Sprintf("login::{%s}", token)
}
