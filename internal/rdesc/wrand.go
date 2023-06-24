package rdesc

import "math/rand"

type Wrand[T any] []WrandItem[T]

type WrandItem[T any] struct {
	Weight float64
	Item   T
}

func (w Wrand[T]) Pick() T {
	var sum float64
	for _, item := range w {
		sum += item.Weight
	}

	r := rand.Float64() * sum

	for _, item := range w {
		if r < item.Weight {
			return item.Item
		}
		r -= item.Weight
	}

	return w[len(w)-1].Item
}
