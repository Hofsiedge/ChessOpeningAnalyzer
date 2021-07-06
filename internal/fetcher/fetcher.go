package fetcher

import (
	"fmt"
	"github.com/notnil/chess"
	"strings"
	"time"
)

type UserGame struct {
	White   bool
	EndTime time.Time
	Moves   []*chess.Move
}

type ConvertibleToUserGame interface {
	UserGame(username string, until int) (*UserGame, error)
}

// ParseMoves parses first until moves from `pgn` PGN string
// If until == 0 all the moves are parsed
func ParseMoves(pgn string, until int) ([]*chess.Move, error) {
	if until < 0 {
		return nil, fmt.Errorf("expected until >= 0, got %v", until)
	}
	pgnReader := strings.NewReader(pgn)
	scanner := chess.NewScanner(pgnReader)
	if !scanner.Scan() && scanner.Err() != nil {
		return nil, fmt.Errorf("could not parse moves from a PGN: %v\nError: %v", pgn, scanner.Err())
	}
	game := scanner.Next()
	moves := game.Moves()
	var numberOfMoves int
	if L := len(moves); until == 0 || L < until {
		numberOfMoves = L
	} else {
		numberOfMoves = until
	}
	return game.Moves()[:numberOfMoves], nil
}