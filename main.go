package main

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
}

func main() {
	t := NewTodo("Monday")
	n1 := t.AddItem("shower")
	n2 := t.AddItem("dry hair")
	n3 := t.AddItem("dry body")
	t.AddRelation(n1, n2)
	t.AddRelation(n1, n3)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wait := NewWait(t)
	wait.Load()
	wait.Stage(ctx)
	for {
		timeout := time.After(5 * time.Second)
		select {
		case n := <-wait.NextNode():
			op := n.(Operation)
			go op.Run()
		case <-timeout:
			return
		}
	}
}
