package main

import (
	"fmt"
	"github.com/seldonio/trtis-scheduler/scheduler/scheduler"
	"math/rand"
	"time"
)

func main() {
	fmt.Println("I'm a scheduler!")

	rand.Seed(time.Now().Unix())

	podQueue := make(chan *scheduler.PodJob, 300)
	defer close(podQueue)

	quit := make(chan struct{})
	defer close(quit)

	s := scheduler.NewScheduler(podQueue, quit)
	s.Run(quit)
}
