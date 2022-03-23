package server

import (
	"context"
	"strings"
	"unicode"
)

type userKey struct{}

type User struct {
	Name string
}

// oidRole oid identifier used to store user roles
var oidRole = []int{1, 2, 840, 10070, 8, 1}

// ParseRoles split the roles string by comma and removes spaces and non-printable characters like EOT
func ParseRoles(roles string) []string {
	return strings.Split(strings.TrimFunc(roles, func(r rune) bool { return !unicode.IsGraphic(r) }), ",")
}

// access map initialization
var access = map[string][]string{
	"/proto.WorkerService/StartJob":        {"admin", "user"},
	"/proto.WorkerService/StopJob":         {"admin", "user"},
	"/proto.WorkerService/GetJobStatus":    {"admin", "user"},
	"/proto.WorkerService/GetOutputStream": {"admin", "user"},
}

// HasAccess verifies the access for a method and user roles
func HasAccess(method string, roles []string) bool {
	permission, ok := access[method]
	if !ok {
		return false
	}
	for _, role := range roles {
		for _, value := range permission {
			if role == value {
				return true
			}
		}
	}
	return false
}

func contains(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func UserFromContext(ctx context.Context) (*User, bool) {
	if u := ctx.Value(userKey{}); u != nil {
		return u.(*User), true
	}
	return nil, false
}
