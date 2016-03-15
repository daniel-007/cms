package main

import (
	"strings"
)

func authInfo(s string) map[string]string {
	m := make(map[string]string)
	for _, auth := range strings.Split(s, ";") {
		kv := strings.Split(auth, "=")
		if len(kv) == 2 {
			m[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return m
}
