package main

import (
	"context"
	"fmt"
	"testing"
)

func TestFanIn(t *testing.T) {
	inputs := make([]chan NodeID, 10)
	for i := 0; i < 10; i++ {
		inputs[i] = make(chan NodeID)
	}

	ctx, cancel := context.WithCancel(context.Background())

	var out <-chan NodeID
	t.Run("Async", func(t *testing.T) {
		readOnlyInputs := make([]<-chan NodeID, 10)
		for i := 0; i < 10; i++ {
			readOnlyInputs[i] = inputs[i]
		}

		out = fanIn(ctx, readOnlyInputs...)

		go func() {
			for i := 0; i < 10; i++ {
				inputs[i] <- NodeID(i)
			}
		}()

		for i := 0; i < 10; i++ {
			fmt.Println(<-out)
		}
	})

	t.Run("Cancel", func(t *testing.T) {
		cancel()
	})
}
