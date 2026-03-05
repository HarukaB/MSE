package odin

import "math"

func float32FromBits(b uint32) float32 {
	return math.Float32frombits(b)
}

func float64FromBits(b uint64) float64 {
	return math.Float64frombits(b)
}

func float32ToBits(f float32) uint32 {
	return math.Float32bits(f)
}

func float64ToBits(f float64) uint64 {
	return math.Float64bits(f)
}
