package secrets

import "strings"

func Canonical(prefix, name string) string {
	if prefix == "" {
		return name
	}
	if strings.HasSuffix(prefix, "/") {
		return prefix + name
	}
	return prefix + "/" + name
}

func ToEnvKey(path string) string {
	return strings.ToUpper(strings.ReplaceAll(path, "/", "_"))
}
