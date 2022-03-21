package server

import (
	"strconv"
	"strings"
	"unicode"
)

// oidRole oid identifier used to store user roles
const oidRole string = "1.2.840.10070.8.1"

// OidToString converts the int[] to string with
// point separator between the values
func OidToString(oid []int) string {
	var strs []string
	for _, value := range oid {
		strs = append(strs, strconv.Itoa(value))
	}
	return strings.Join(strs, ".")
}

// IsOidRole validates the role oid
func IsOidRole(oid string) bool {
	return oidRole == oid
}

// ParseRoles split the roles string by comma and removes spaces and non-printable characters like EOT
func ParseRoles(roles string) []string {
	return strings.Split(strings.TrimFunc(roles, func(r rune) bool { return !unicode.IsGraphic(r) }), ",")
}

// access map initialization
var access = map[string][]string{
	"/proto.WorkerService/StartJob":     {"admin", "user"},
	"/proto.WorkerService/StopJob":      {"admin", "user"},
	"/proto.WorkerService/GetJobStatus": {"admin", "user"},
	"/proto.WorkerService/GetJobStream": {"admin", "user"},
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
