package main

import (
	"fmt"
	"github.com/google/uuid"
)

func main() {
	u := uuid.New()
	fmt.Println(u, "--")
	awaitingSequences := make(map[uuid.UUID]map[int]struct{})
	awaitingSequences[u] = make(map[int]struct{})
	awaitingSequences[u][1] = struct{}{}
	s, e := awaitingSequences[u][1]
	fmt.Println(s, "-", e)
}
