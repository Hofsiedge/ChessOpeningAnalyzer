package positions

import (
	"fmt"
	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/fetching"
	"github.com/notnil/chess"
	"strings"
	"time"
)

type FEN string

type Position struct {
	FEN       FEN
	Score     float32
	Evaluated bool
}

type PositionNode struct {
	Position   *Position
	LastPlayed time.Time
	Moves      []*Move
}

type Move struct {
	To   *PositionNode
	Move string
}

type PositionGraph struct {
	Depth          int
	WhitePositions *PositionNode
	BlackPositions *PositionNode
	PositionMap    map[FEN]*PositionNode
}

func NewPositionGraph(depth int) (*PositionGraph, error) {
	if depth <= 1 {
		return nil, fmt.Errorf("expected depth > 1, got: %v", depth)
	}
	graph := PositionGraph{
		Depth:       depth,
		PositionMap: make(map[FEN]*PositionNode, 30),
	}
	for _, positions := range []**PositionNode{&graph.WhitePositions, &graph.BlackPositions} {
		*positions = &PositionNode{
			Position: &Position{
				FEN:       FEN(chess.StartingPosition().String()),
				Score:     0,
				Evaluated: true,
			},
			LastPlayed: time.Time{},
		}
	}
	return &graph, nil
}

// truncateFEN removes move counters & last captures since they are not important for opening analysis.
// It allows to build a more representative cache in PositionGraph.AddGame
func truncateFEN(fen string) FEN {
	words := strings.Split(fen, " ")
	return FEN(strings.Join(words[:len(words)-3], " "))
}

// AddGame adds the first moves of the game to the position graph
func (g *PositionGraph) AddGame(game fetching.UserGame) error {
	board := chess.NewGame()
	var currentNode *PositionNode
	if game.White {
		currentNode = g.WhitePositions
	} else {
		currentNode = g.BlackPositions
	}
	for _, move := range game.Moves {
		if err := board.MoveStr(move); err != nil {
			return err
		}
		var nextNode *PositionNode
		pos := truncateFEN(board.Position().String())
		if node, found := g.PositionMap[pos]; found {
			nextNode = node
		} else {
			nextNode = &PositionNode{
				Position: &Position{
					FEN: pos,
				},
				LastPlayed: game.EndTime,
			}
			g.PositionMap[pos] = nextNode
			currentNode.Moves = append(currentNode.Moves, &Move{nextNode, move})
		}
		currentNode = nextNode
	}
	return nil
}

// TODO: accept a context or a `done` channel

// GetVariations returns all the move sequences from the first move to the last one
// that are present in the position graph to the output channel
func (g PositionGraph) GetVariations() <-chan []*Move {
	out := make(chan []*Move)
	go func() {
		for _, startingNode := range []*PositionNode{g.WhitePositions, g.BlackPositions} {
			for _, move := range startingNode.Moves {
				traverseMoves(move, make([]*Move, 0, 1), out)
			}
		}
		close(out)
	}()
	return out
}

func traverseMoves(move *Move, previousMoves []*Move, out chan<- []*Move) {
	moves := append(previousMoves, move)
	if len(move.To.Moves) == 0 {
		out <- moves
	}
	for _, nextMove := range move.To.Moves {
		traverseMoves(nextMove, moves, out)
	}
}
