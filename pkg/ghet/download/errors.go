package download

import (
	"fmt"

	"github.com/pkg/errors"
)

// ErrUnexpected is returned when an unexpected error occurs.
var ErrUnexpected = errors.New("unexpected error")

func unexpected(err error) error {
	if errors.Is(err, ErrUnexpected) {
		return err
	}
	return errors.WithStack(fmt.Errorf("%w: %v", ErrUnexpected, err))
}
