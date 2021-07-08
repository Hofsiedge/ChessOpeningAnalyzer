package positions

import (
	"encoding/gob"
	"os"
)

// DumpGraph encodes the graph into a binary file at the provided path
func DumpGraph(graph *PositionGraph, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	encoder := gob.NewEncoder(file)
	return encoder.Encode(graph)
}

// LoadGraph decodes the graph from a file generated with DumpGraph
func LoadGraph(path string) (*PositionGraph, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	decoder := gob.NewDecoder(file)
	graph := new(PositionGraph)
	err = decoder.Decode(graph)
	return graph, err
}
