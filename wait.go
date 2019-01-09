package main

import (
	"context"
	"errors"
	"fmt"

	"gonum.org/v1/gonum/graph"
)

type DirectedGraph interface {
	graph.Graph
	To(id int64) graph.Nodes
	Sources() []graph.Node
}

func NewWait(t *Todo) *Wait {
	return &Wait{
		graph:     t,
		loaded:    false,
		successes: make(map[int64]chan NodeID),
		failures:  make(map[int64]chan NodeID),
		next:      make(chan NodeID),
	}
}

// Wait is a staging area for nodes to wait for dependency resolution.
type Wait struct {
	graph     DirectedGraph
	loaded    bool
	successes map[int64]chan NodeID // Demultiplexed success channels
	failures  map[int64]chan NodeID // Demultiplexed error channels
	next      chan NodeID
}

// NextNode returns the next ready node.
func (w *Wait) NextNode() <-chan graph.Node {
	ch := make(chan graph.Node)
	go func() {
		id := <-w.next
		ch <- w.graph.Node(int64(id))
	}()

	return ch
}

// Load configures goroutines to listen for each operation outcome (success/failure) and demultiplex
// the result into multiple copies to fulfill the needs of each node's dependents.
func (w *Wait) Load() error {
	nodes := w.graph.Nodes()
	for nodes.Next() {
		n := nodes.Node()
		op, ok := n.(Operation)
		if !ok {
			return fmt.Errorf("node %d does not implement Operation interface", n.ID())
		}

		w.successes[op.ID()] = make(chan NodeID)
		w.failures[op.ID()] = make(chan NodeID)

		demux := func(n int, succ, fail <-chan NodeID, succOut, failOut chan<- NodeID) {
			var id NodeID
			var out chan<- NodeID
			select {
			case id = <-succ:
				out = succOut
			case id = <-fail:
				out = failOut
			}

			for i := 0; i < n; i++ {
				out <- id
			}
		}

		// Setup demultiplexing structure to wait for op's result. Why do we need demux? Every
		// operation has one result, either success or failure. Each operation may have multiple
		// dependents. The one result needs to be demultiplexed into multiple ones to make sure each
		// dependent receives one copy.
		go demux(w.graph.From(op.ID()).Len(), op.Success(), op.Failure(),
			w.successes[op.ID()], w.failures[op.ID()])
	}

	w.loaded = true

	return nil
}

// Stage puts every node onto the stage and set them to wait state.
func (w *Wait) Stage(ctx context.Context) error {
	if !w.loaded {
		return errors.New("wait area is not loaded")
	}

	queue := w.graph.Sources()
	visited := make(map[int64]struct{})

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		// Mark as visited
		visited[current.ID()] = struct{}{}

		if err := w.stageNode(ctx, current); err != nil {
			return err
		}

		children := w.graph.From(current.ID())
		for children.Next() {
			child := children.Node()
			if _, ok := visited[child.ID()]; !ok {
				queue = append(queue, child)
			}
		}
	}

	return nil
}

func (w *Wait) stageNode(ctx context.Context, n graph.Node) error {
	op := n.(Operation)
	parents := w.graph.To(n.ID())

	deps := []<-chan NodeID{}
	for parents.Next() {
		parent := parents.Node()

		edge := w.graph.Edge(parent.ID(), op.ID())

		r := edge.(*Relation)
		switch r.etype {
		case Success:
			deps = append(deps, w.successes[parent.ID()])
		case Error:
			deps = append(deps, w.failures[parent.ID()])
		}
	}

	switch op.ConnectorType() {
	case And:
		go w.andGate(ctx, NodeID(op.ID()), deps)
	case Or:
		go w.orGate(ctx, NodeID(op.ID()), deps)
	}

	return nil
}

// andGate is equivalent to fan-in-and logic.
func (w *Wait) andGate(ctx context.Context, id NodeID, deps []<-chan NodeID) {
	out := fanIn(ctx, deps...)
	for i := 0; i < len(deps); i++ {
		<-out
	}

	w.next <- id
}

// orGate is equivalent to fan-in-or logic.
func (w *Wait) orGate(ctx context.Context, id NodeID, deps []<-chan NodeID) {
	out := fanIn(ctx, deps...)
	for i := 0; i < len(deps); i++ {
		<-out
		w.next <- id
		return
	}
}

// fanIn will read exactly one input from each channel and multiplex them into one output channel.
func fanIn(ctx context.Context, inputs ...<-chan NodeID) <-chan NodeID {
	out := make(chan NodeID, len(inputs))

	for _, in := range inputs {
		go func(ch <-chan NodeID) {
			select {
			case <-ctx.Done():
				return
			case out <- <-ch:
			}
		}(in)
	}

	return out
}
