package Server

import (
	"math/rand"
)

type Entity struct {
	ability []int8
}

//Random to Engine
func (entity *Entity) getScore(index int) int {
	return rand.Intn(20) + 1 + int(entity.ability[index])
}
