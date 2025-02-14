package must_test

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tombenke/go-12f-common/v2/must"
)

type errCloser struct{}

func (e errCloser) Close() error {
	return errors.New("cannot be closed")
}

func TestMust(t *testing.T) {
	require.NotPanics(t, func() {
		must.Must(nil)
	})

	require.Panics(t, func() {
		must.Must(errors.New("error"))
	})
}

func valNil() (uint, error) {
	return 42, nil
}

func valErr() (uint, error) {
	return 24, errors.New("error")
}

func TestMustVal(t *testing.T) {
	require.NotPanics(t, func() {
		must.MustVal(42, nil)
	})
	require.NotPanics(t, func() {
		v := must.MustVal(valNil())
		require.Equal(t, uint(42), v)
	})

	require.Panics(t, func() {
		must.MustVal(0, errors.New("error"))
	})
	require.Panics(t, func() {
		v := must.MustVal(valErr())
		require.Equal(t, uint(24), v)
	})
}

func TestClose(t *testing.T) {
	require.NotPanics(t, func() {
		must.Close(io.NopCloser(bytes.NewBuffer(nil)))
	})

	require.Panics(t, func() {
		must.Close(errCloser{})
	})
}
