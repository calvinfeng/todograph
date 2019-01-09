package main

import (
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

// NewTodo returns a todo graph.
func NewTodo(title string) *Todo {
	return &Todo{simple.NewDirectedGraph(), title}
}

// Todo is a directed acyclic graph of todo items.
type Todo struct {
	*simple.DirectedGraph
	Title string
}

// AddItem creates a new item in the todo list.
func (t *Todo) AddItem(n string) graph.Node {
	node := &Item{t.NewNode(), And, n, make(chan NodeID), make(chan NodeID)}
	t.AddNode(node)
	return node
}

// AddRelation establishes relationships between todo items.
func (t *Todo) AddRelation(from, to graph.Node) graph.Edge {
	edge := &Relation{t.NewEdge(from, to), Success}
	t.SetEdge(edge)
	return edge
}

// Adjacency returns an adjacency list for each node of the graph.
func (t *Todo) Adjacency() map[int64][]int64 {
	adj := make(map[int64][]int64)

	nodeit := t.Nodes()
	for nodeit.Next() {
		n := nodeit.Node()
		adj[n.ID()] = []int64{}
	}

	edgeit := t.Edges()
	for edgeit.Next() {
		e := edgeit.Edge()
		adj[e.From().ID()] = append(adj[e.From().ID()], e.To().ID())
	}

	return adj
}

// Sources returns all source nodes which have zero in-degree.
func (t *Todo) Sources() []graph.Node {
	roots := []graph.Node{}

	nodeit := t.Nodes()
	for nodeit.Next() {
		n := nodeit.Node()
		if t.To(n.ID()).Len() == 0 {
			roots = append(roots, n)
		}
	}

	return roots
}
