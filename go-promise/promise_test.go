package promise

import (
	"context"
	"errors"
	"testing"

	conc "github.com/sourcegraph/conc/pool"
	"github.com/stretchr/testify/require"
)

var (
	ctx         = context.Background()
	errExpected = errors.New("expected error")
)

func TestNew(t *testing.T) {
	p := New(func(resolve func(any), reject func(error)) {
		resolve(nil)
	})
	require.NotNil(t, p)
}

func TestNewWithPool(t *testing.T) {
	tests := []struct {
		name string
		pool Pool
	}{
		{
			name: "default",
			pool: newDefaultPool(),
		},
		{
			name: "conc",
			pool: func() Pool {
				return FromConcPool(conc.New())
			}(),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := NewWithPool(func(resolve func(string), reject func(error)) {
				resolve(test.name)
			}, test.pool)

			val, err := p.Await(ctx)
			require.NoError(t, err)
			require.NotNil(t, val)
			require.Equal(t, test.name, *val)
		})
	}
}

func TestPromise_Then(t *testing.T) {
	p1 := New(func(resolve func(string), reject func(error)) {
		resolve("Hello, ")
	})
	p2 := Then(p1, ctx, func(data string) (string, error) {
		return data + "world!", nil
	})

	p3 := Then(p2, ctx, func(_ string) (string, error) {
		return "", errExpected
	})

	val, err := p1.Await(ctx)
	require.NoError(t, err)
	require.NotNil(t, val)
	require.Equal(t, "Hello, ", *val)

	val, err = p2.Await(ctx)
	require.NoError(t, err)
	require.NotNil(t, val)
	require.Equal(t, "Hello, world!", *val)

	_, err = p3.Await(ctx)
	require.EqualError(t, err, errExpected.Error())
}

func TestPromise_Catch(t *testing.T) {
	p1 := New(func(resolve func(any), reject func(error)) {
		reject(errExpected)
	})

	val, err := p1.Await(ctx)
	require.Error(t, err)
	require.Equal(t, errExpected, err)
	require.Nil(t, val)
}

func TestPromise_Panic(t *testing.T) {
	p1 := New(func(resolve func(any), reject func(error)) {
		panic("random error")
	})
	p2 := New(func(resolve func(any), reject func(error)) {
		panic(errExpected)
	})

	val, err := p1.Await(ctx)
	require.Error(t, err)
	require.Equal(t, errors.New("random error"), err)
	require.Nil(t, val)

	val, err = p2.Await(ctx)
	require.Error(t, err)
	require.ErrorIs(t, err, errExpected)
	require.Nil(t, val)
}
