package main

import (
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
	n1 := t.AddItem("item 1")
	n2 := t.AddItem("item 2")
	t.AddRelation(n1, n2)

	wait := NewWait(t)
	wait.Load()
}
