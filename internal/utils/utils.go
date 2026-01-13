package utils

import (
	"fmt"
	"regexp"
)

var charNameRegex = regexp.MustCompile("^[a-zA-Z0-9]+$")

func IsValidCharacterName(name string) bool {
	return len(name) >= 4 && len(name) <= 13 && charNameRegex.MatchString(name)
}

func ParseIP(host string) []byte {
	// Parse IP string like "127.0.0.1" into 4 bytes
	ip := make([]byte, 4)
	var a, b, c, d int
	_, _ = fmt.Sscanf(host, "%d.%d.%d.%d", &a, &b, &c, &d) // Ignore parse errors
	ip[0] = byte(a)
	ip[1] = byte(b)
	ip[2] = byte(c)
	ip[3] = byte(d)
	return ip
}
