package util

import (
	"strings"
)

func ParseHeaderMap(raw string) map[string]string {
	headers := make(map[string]string)
	if raw == "" {
		return headers
	}

	for _, pair := range strings.Split(raw, ",") {
		kv := strings.SplitN(pair, ":", 2)
		if len(kv) == 2 {
			headers[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}

	return headers
}
