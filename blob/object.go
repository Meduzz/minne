package blob

import (
	"fmt"
	"strings"
)

type (
	// Object identifies a single file
	Object string
)

func NewObject(name string) (Object, error) {
	name = strings.TrimSpace(name)

	if name == "" {
		return "", fmt.Errorf("object should not be empty")
	}

	if !strings.HasPrefix(name, "/") {
		return "", fmt.Errorf("object should be absolute paths")
	}

	return Object(name), nil
}

func NewFormatObject(name, format string) (Object, error) {
	return NewObject(fmt.Sprintf(format, name))
}

func NewJsonObject(name string) (Object, error) {
	return NewFormatObject(name, "/%s.json")
}

func (kv Object) File() string {
	return string(kv)
}
