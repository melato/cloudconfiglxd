package lxdutil

import (
	"fmt"
)

func AnnotateLXDError(name string, err error) error {
	if err == nil {
		return err
	}
	return fmt.Errorf("%s: %w", name, err)
}
