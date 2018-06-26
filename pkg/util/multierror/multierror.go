package multierror

import (
	"fmt"
)

// derived from https://github.com/golang/appengine/blob/master/errors.go

// MultiError is returned by batch operations.
type MultiError []error

func (m MultiError) String() string {
	s, n := "", 0
	for _, e := range m {
		if e != nil {
			if n == 0 {
				s = e.Error()
			}
			n++
		}
	}
	switch n {
	case 0:
		return "(0 errors)"
	case 1:
		return s
	case 2:
		return s + " (and 1 other error)"
	}
	return fmt.Sprintf("%s (and %d other errors)", s, n-1)
}
