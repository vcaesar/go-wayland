package cursor

import "strings"

type MultiError struct {
	errs []error
}

func (e *MultiError) Err() error {
	if len(e.errs) == 0 {
		return nil
	}
	return e
}

func (e *MultiError) Add(err error) {
	if err == nil {
		return
	}
	e.errs = append(e.errs, err)
}

// Error returns `\n` separated string of all errors.
func (e *MultiError) Error() string {
	var s strings.Builder

	for _, err := range e.errs {
		s.WriteString(err.Error())
		s.WriteByte('\n')
	}

	return s.String()
}
