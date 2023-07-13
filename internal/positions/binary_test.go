package positions

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/fetching"
	"github.com/notnil/chess"
)

func TestGraphBinary(t *testing.T) {
	graph, err := NewPositionGraph(4)
	if err != nil {
		t.Error(err)
	}
	var moves []string
	pgn := "\n1. e4 e5 2. Nf3 Nc6 3. d4 exd4 1-0\n\n"
	scanner := chess.NewScanner(strings.NewReader(pgn))
	if !scanner.Scan() {
		panic(fmt.Errorf("invalid test data. chess.Scanner could not parse PGN string: %s", pgn))
	}
	game := scanner.Next()
	if moves, err = fetching.ParseMoves(game, 4); err != nil {
		t.Errorf("could not parse moves from a test PGN: %v;\nError: %v", pgn, err)
	}
	userGame := fetching.UserGame{
		White:   true,
		EndTime: time.Date(2021, 1, 1, 1, 0, 0, 0, time.UTC),
		Moves:   moves,
	}
	if err := graph.AddGame(userGame); err != nil {
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
