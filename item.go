package main

import (
	"time"

	"github.com/sirupsen/logrus"
	"gonum.org/v1/gonum/graph"
)

type (
	NodeID        int64
	EdgeType      int
	ConnectorType int
)

const (
	Success EdgeType = iota
	Error
	Done
)

const (
	And ConnectorType = iota
	Or
)

type (
	Operation interface {
		graph.Node
		Run()
		ConnectorType() ConnectorType
		Success() <-chan NodeID
		Failure() <-chan NodeID
	}

	Item struct {
		graph.Node
		ctype   ConnectorType
		name    string
		success chan NodeID
		err     chan NodeID
	}

	Relation struct {
		graph.Edge
		etype EdgeType
	}
)

func (it *Item) Run() {
	logrus.Infof("item %d - %s has started", it.ID(), it.name)
	time.Sleep(time.Second)
	it.success <- NodeID(it.ID())
	logrus.Infof("item %d - %s has completed", it.ID(), it.name)
}

func (it *Item) ConnectorType() ConnectorType {
	return it.ctype
}

func (it *Item) Success() <-chan NodeID {
	return it.success
}

func (it *Item) Failure() <-chan NodeID {
	return it.err
}
