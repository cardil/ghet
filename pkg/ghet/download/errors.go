package download

import (
	"fmt"

	"emperror.dev/errors"
)

// ErrUnexpected is returned when an unexpected error occurs.
var ErrUnexpected = errors.New("unexpected error")

func unexpected(err error) error {
	if errors.Is(err, ErrUnexpected) {
		return err
	}
	return errors.WithStack(
		errors.Wrap(ErrUnexpected, fmt.Sprintf("%+v", err)),
	)
}
