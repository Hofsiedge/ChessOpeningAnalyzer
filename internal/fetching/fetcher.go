package fetching

import (
	"fmt"
	"github.com/notnil/chess"
	"strings"
	"time"
)

type UserGame struct {
	White   bool
	EndTime time.Time
	Moves   []string
}

type ConvertibleToUserGame interface {
	UserGame(username string, until int) (*UserGame, error)
}

type FilterOptions struct {
	TimePeriodStart  time.Time
	TimePeriodEnd    time.Time
	Color            chess.Color
	NumberOfMovesCap int
	// TODO: time control
}

type GameFetcher interface {
	Fetch(username string, filter FilterOptions, workers int) ([]*UserGame, error)
}

// ParseMoves parses first until moves from `pgn` PGN string
// If until == 0 all the moves are parsed
func ParseMoves(pgn string, until int) ([]string, error) {
	if until < 0 {
		return nil, fmt.Errorf("expected until >= 0, got %v", until)
	}
	pgnReader := strings.NewReader(pgn)
	scanner := chess.NewScanner(pgnReader)
	if !scanner.Scan() && scanner.Err() != nil {
		return nil, fmt.Errorf("could not parse moves from a PGN: %v\nError: %v", pgn, scanner.Err())
	}
	game := scanner.Next()
	notation := chess.AlgebraicNotation{}
	moves := game.Moves()
	gamePositions := game.Positions()
	var numberOfMoves int
	if L := len(moves); until == 0 || L < until {
		numberOfMoves = L
	} else {
		numberOfMoves = until
	}
	strMoves := make([]string, numberOfMoves)
	for i, move := range moves[:numberOfMoves] {
		strMoves[i] = notation.Encode(gamePositions[i], move)
	}
	return strMoves, nil
}
