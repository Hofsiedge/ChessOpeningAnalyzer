package positions

import (
	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/fetching"
	"reflect"
	"testing"
	"time"
)

func TestGraphBinary(t *testing.T) {
	graph, err := NewPositionGraph(4)
	if err != nil {
		t.Error(err)
	}
	var moves []string
	pgn := "\n1. e4 e5 2. Nf3 Nc6 3. d4 exd4 1-0\n\n"
	if moves, err = fetching.ParseMoves(pgn, 4); err != nil {
		t.Errorf("could not parse moves from a test PGN: %v;\nError: %v", pgn, err)
	}
	game := fetching.UserGame{
		White:   true,
		EndTime: time.Date(2021, 1, 1, 1, 0, 0, 0, time.UTC),
		Moves:   moves,
	}
	if err := graph.AddGame(game); err != nil {
		t.Errorf("%v", err)
	}
	if err = DumpGraph(graph, "../../../graph.bin"); err != nil {
		t.Errorf("could not dump a PositionGraph: %v", err)
		return
	}
	if newGraph, err := LoadGraph("../../../graph.bin"); err != nil || !reflect.DeepEqual(newGraph, graph) {
		t.Errorf("could not load a PositionGraph - error: %v", err)
	}
}
