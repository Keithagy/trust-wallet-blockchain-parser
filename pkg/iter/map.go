package iter

import "iter"

func Map[T, U any](slice []T, transform func(T) U) iter.Seq[U] {
	return func(yield func(U) bool) {
		for _, v := range slice {
			if !yield(transform(v)) {
				break
			}
		}
	}
}
