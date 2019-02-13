package main

import (
	"math/rand"
)

func Chance(prob float32) bool {
	return prob < rand.Float32()
}