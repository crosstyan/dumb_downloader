package utils

import (
	"github.com/samber/mo"
)

// Map maps an array of objects to another array of objects.
//
// https://stackoverflow.com/questions/71624828/is-there-a-way-to-map-an-array-of-objects-in-golang
func Map[T, U any](ts []T, f func(T) U) []U {
	us := make([]U, len(ts))
	for i := range ts {
		us[i] = f(ts[i])
	}
	return us
}

// TryGet gets the first value from a map that matches one of the keys.
func TryGet[T any](m map[string]T, keys ...string) mo.Option[T] {
	for _, k := range keys {
		v, ok := m[k]
		if ok {
			return mo.Some[T](v)
		}
	}
	return mo.None[T]()
}
