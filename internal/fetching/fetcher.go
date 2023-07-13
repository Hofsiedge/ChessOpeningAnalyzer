package fetching

import (
	"errors"
	"fmt"
	"time"

	"github.com/notnil/chess"
)

var (
	UserNotFoundError = errors.New("user not found")
	ArgumentError     = errors.New("invalid argument")
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

// ParseMoves parses first `until` moves from `game`
// If until == 0 all the moves are parsed
func ParseMoves(game *chess.Game, until int) ([]string, error) {
	if until < 0 {
		return nil, fmt.Errorf(
			"fetcher.ParseMoves: %w: expected until >= 0, got %v",
			ArgumentError, until)
	}
	if game == nil {
		return nil, fmt.Errorf(
			"fetcher.ParseMoves: %w: got a nil game",
			ArgumentError)
	}
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
