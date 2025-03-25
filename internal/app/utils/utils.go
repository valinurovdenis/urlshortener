// Package for auxiliary functions and structs.
package utils

import "bytes"

// Concatenates all given strings in one.
func AddStrings(strings ...string) string {
	bufferLength := 0
	for _, s := range strings {
		bufferLength += len(s)
	}
	buffer := bytes.NewBuffer(make([]byte, 0, bufferLength))
	for _, s := range strings {
		buffer.WriteString(s)
	}
	return buffer.String()
}
