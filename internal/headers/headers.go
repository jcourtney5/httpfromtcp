package headers

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
)

const crlf = "\r\n"

type Headers map[string]string

func NewHeaders() Headers {
	return map[string]string{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, []byte(crlf))
	if idx == -1 {
		return 0, false, nil
	}

	// we found the end of the headers so we are done
	if idx == 0 {
		return 2, true, nil
	}

	headerText := string(data[:idx])

	// split by first occurrence of ":"
	key, value, found := strings.Cut(headerText, ":")
	if !found {
		return 0, false, fmt.Errorf("invalid header: %s", headerText)
	}

	// make sure key does not have trailing whitespace between it and the semicolon
	if len(key) == 0 || strings.HasSuffix(key, " ") {
		return 0, false, fmt.Errorf("invalid header name: %s", key)
	}

	// trim spaces from key and value
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)

	// make sure all the characters in the header are valid
	if !isValidHeaderKey([]byte(key)) {
		return 0, false, fmt.Errorf("invalid characters in header key: %s", key)
	}

	// add to the headers map
	h.Set(key, value)

	return idx + 2, false, nil
}

func (h Headers) Set(key, value string) {
	key = strings.ToLower(key)
	existingValue, ok := h[key]
	if ok {
		h[key] = existingValue + ", " + value
	} else {
		h[key] = value
	}
}

var tokenChars = []byte{'!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~'}

func isValidHeaderKey(key []byte) bool {
	if len(key) == 0 {
		return false
	}
	for i := 0; i < len(key); i++ {
		if !isValidHeaderChar(key[i]) {
			return false
		}
	}
	return true
}

func isValidHeaderChar(c byte) bool {
	if c >= 'A' && c <= 'Z' ||
		c >= 'a' && c <= 'z' ||
		c >= '0' && c <= '9' {
		return true
	}

	return slices.Contains(tokenChars, c)
}
