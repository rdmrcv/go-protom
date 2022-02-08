package protobson

import (
	"fmt"
	"strings"
)

// DecodeError represents an error that occurs when unmarshalling BSON bytes into a native Go type.
type DecodeError struct {
	keys    []string
	wrapped error
}

// Unwrap returns the underlying error
func (de *DecodeError) Unwrap() error {
	return de.wrapped
}

// Error implements the error interface.
func (de *DecodeError) Error() string {
	// The keys are stored in reverse order because the de.keys slice is builtup while propagating the error up the
	// stack of BSON keys, so we call de.Keys(), which reverses them.
	keyPath := strings.Join(de.Keys(), ".")
	return fmt.Sprintf("error decoding key %s: %v", keyPath, de.wrapped)
}

// Keys returns the BSON key path that caused an error as a slice of strings. The keys in the slice are in top-down
// order. For example, if the document being unmarshalled was {a: {b: {c: 1}}} and the value for c was supposed to be
// a string, the keys slice will be ["a", "b", "c"].
func (de *DecodeError) Keys() []string {
	reversedKeys := make([]string, 0, len(de.keys))
	for idx := len(de.keys) - 1; idx >= 0; idx-- {
		reversedKeys = append(reversedKeys, de.keys[idx])
	}

	return reversedKeys
}

func newDecodeError(key string, original error) error {
	de, ok := original.(*DecodeError)
	if !ok {
		return &DecodeError{
			keys:    []string{key},
			wrapped: original,
		}
	}

	de.keys = append(de.keys, key)
	return de
}
