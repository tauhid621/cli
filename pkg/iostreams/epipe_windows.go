package iostreams

import (
	"errors"
	"syscall"
)

func isEpipeError(err error) bool {
	return errors.Is(err, syscall.ERROR_NO_DATA)
}
