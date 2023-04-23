package main

import (
	"ScanTodo/utils"
	"fmt"
	"github.com/google/uuid"
	"time"
)

func main() {
	u := uuid.New()
	fmt.Println(u, "--")
	awaitingSequences := make(map[uuid.UUID]map[int]struct{})
	awaitingSequences[u] = make(map[int]struct{})
	awaitingSequences[u][1] = struct{}{}
	s, e := awaitingSequences[u][1]
	fmt.Println(s, "-", e)
	utils.TimeToBytes(time.Now())
	//var nano int64
	//nano += 10 << 1
	//fmt.Println(nano, "ce")
}
