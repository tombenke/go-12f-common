// Package must package defines simple wrappers for operations that must succeed.
package must

import "io"

// Must panics if err is not nil.
//
// This is a helper function intended to be used with defer.
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func MustVal[T any](v T, err error) T {
	Must(err)

	return v
}

// Close tries to close the given closer parameter or panics.
//
// This is a helper function intended to be used with defer.
func Close(c io.Closer) {
	Must(c.Close())
}
