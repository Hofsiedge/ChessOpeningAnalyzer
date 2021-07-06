package positions

import (
	"fmt"
	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/fetcher"
	"github.com/notnil/chess"
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
	Move *chess.Move
}

type PositionGraph struct {
	depth          int
	whitePositions *PositionNode
	blackPositions *PositionNode
	positionMap    map[FEN]*PositionNode
}

func NewPositionGraph(depth int) (*PositionGraph, error) {
	if depth <= 1 {
		return nil, fmt.Errorf("expected depth > 1, got: %v", depth)
	}
	graph := PositionGraph{
		depth:       depth,
		positionMap: make(map[FEN]*PositionNode, 30),
	}
	for _, positions := range []**PositionNode{&graph.whitePositions, &graph.blackPositions} {
		*positions = &PositionNode{
			Position: &Position{
				FEN:       FEN(chess.StartingPosition().String()),
				Score:     0,
				Evaluated: true,
			},
			LastPlayed: time.Time{},
			Moves:      make([]*Move, 0, 5),
		}
	}
	return &graph, nil
}

// AddGame adds the first moves of the game to the position graph
func (g *PositionGraph) AddGame(game fetcher.UserGame) error {
	board := chess.NewGame()
	var currentNode *PositionNode
	if game.White {
		currentNode = g.whitePositions
	} else {
		currentNode = g.blackPositions
	}
	for _, move := range game.Moves {
		if err := board.Move(move); err != nil {
			return err
		}
		var nextNode *PositionNode
		pos := FEN(board.Position().String())
		if node, found := g.positionMap[pos]; found {
			nextNode = node
		} else {
			nextNode = &PositionNode{
				Position: &Position{
					FEN: pos,
				},
				LastPlayed: game.EndTime,
				Moves:      make([]*Move, 0, 1),
			}
			g.positionMap[pos] = nextNode
		}
		currentNode.Moves = append(currentNode.Moves, &Move{nextNode, move})
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
		for _, startingNode := range []*PositionNode{g.whitePositions, g.blackPositions} {
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
