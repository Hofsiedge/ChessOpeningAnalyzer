package positions

import (
	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/fetching"
	"github.com/notnil/chess"
	"testing"
	"time"
)

func TestAddGame(t *testing.T) {
	graph, err := NewPositionGraph(2)
	if err != nil {
		t.Errorf("could not construct a positon graph: %v", err)
		return
	}
	moves := make([][]*chess.Move, 2)
	PGNs := []string{
		"\n1. e4 e5 2. Nf3 Nc6 3. d4 exd4 1-0\n\n",
		"\n1. e4 c5 2. c3 Nc6 3. d4 cxd4 1-0\n\n",
	}
	for i := range PGNs {
		if moves[i], err = fetching.ParseMoves(PGNs[i], 4); err != nil {
			t.Errorf("could not parse moves from a test PGN: %v;\nError: %v", PGNs[i], err)
		}
	}
	games := make([]fetching.UserGame, 2)
	for i, m := range moves {
		games[i] = fetching.UserGame{
			White:   true,
			EndTime: time.Date(2021, 1, 1, 1, 0, 0, 0, time.UTC),
			Moves:   m,
		}
	}
	for _, game := range games {
		if err := graph.AddGame(game); err != nil {
			t.Errorf("%v", err)
		}
	}

	if len(graph.blackPositions.Moves) != 0 {
		t.Errorf("expected moves as black to be empty but got %v", graph.blackPositions.Moves)
	}
	expectedVariations := [][]string{
		{"e2e4", "e7e5", "g1f3", "b8c6"},
		{"e2e4", "c7c5", "c2c3", "b8c6"},
	}
	variations := graph.GetVariations()
	for _, expected := range expectedVariations {
		v := <-variations
		if lv, le := len(v), len(expected); lv != le {
			t.Errorf("wrong number of moves - expected %v, got %v", le, lv)
		}
		for i, move := range v {
			if expected[i] != move.Move.String() {
				t.Errorf("wrong move %v - expected %v", move.Move.String(), expected[i])
			}
		}
	}
}
