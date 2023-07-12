package lichess

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/fetching"
	"github.com/notnil/chess"
)

var (
	ErrInvalidUserName = errors.New("invalid username")
	ErrRequestError    = errors.New("request error")
	ErrInvalidPGNTags  = errors.New("invalid PGN tag pairs")
)

type Fetcher struct {
}

func (f *Fetcher) Fetch(username string, filter fetching.FilterOptions, _ int) ([]*fetching.UserGame, error) {
	var (
		err      error
		response *http.Response
	)
	requestURL, err := makeLichessURL(username, filter)
	log.Printf("performing GET request to %s", requestURL.String())
	if response, err = http.Get(requestURL.String()); err != nil {
		log.Printf("attempted to perform a GET request to %s", &requestURL)
		return nil, fmt.Errorf("lichess.Fetch: http.Get error: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status code %d", ErrRequestError, response.StatusCode)
	}

	gamesCh, errsCh := parseLichessPGN(response.Body, username, filter)
	games := make([]*fetching.UserGame, 0)
	invalidGames := 0
	for {
		select {
		case game, ok := <-gamesCh:
			if !ok {
				gamesCh = nil
			} else {
				games = append(games, game)
			}
		case err, ok := <-errsCh:
			if !ok {
				errsCh = nil
			} else {
				if errors.Is(err, ErrInvalidPGNTags) {
					invalidGames++
				} else if err != nil {
					return nil, err
				}
			}
		}
		if (gamesCh == nil) && (errsCh == nil) {
			break
		}
	}
	if invalidGames != 0 {
		log.Printf("got %d invalid games", invalidGames)
	}
	return games, nil
}

func makeLichessURL(username string, filter fetching.FilterOptions) (url.URL, error) {
	path, err := url.JoinPath("api", "games", "user", strings.ToLower(username))
	if err != nil {
		return url.URL{}, ErrInvalidUserName
	}
	timeConverter := func(t time.Time) []string {
		return []string{fmt.Sprintf("%d", t.UTC().UnixMilli())}
	}
	queryParams := url.Values{
		"since": timeConverter(filter.TimePeriodStart),
		"until": timeConverter(filter.TimePeriodEnd),
	}
	if filter.Color != chess.NoColor {
		if filter.Color == chess.White {
			queryParams.Add("color", "white")
		} else {
			queryParams.Add("color", "black")
		}
	}
	requestURL := url.URL{
		Scheme:   "https",
		Host:     "lichess.org",
		Path:     path,
		RawQuery: queryParams.Encode(),
	}
	return requestURL, nil
}

func parseLichessPGN(reader io.Reader, username string, filter fetching.FilterOptions) (<-chan *fetching.UserGame, <-chan error) {
	decoder := chess.NewScanner(reader)
	games := make(chan *fetching.UserGame)
	errs := make(chan error)

	go func() {
		for decoder.Scan() {
			game := decoder.Next()

			log.Printf(game.String())

			// reading color
			userPlaysWhite, err := userIsWhite(game, username)
			if err != nil {
				errs <- err
				continue
			}
			if (filter.Color == chess.White) != userPlaysWhite {
				continue
			}

			// reading date and time
			timestamp, err := getTimeFromGame(game)
			if err != nil {
				errs <- err
				continue
			}
			if !(timestamp.After(filter.TimePeriodStart) && timestamp.Before(filter.TimePeriodEnd)) {
				continue
			}

			moves, err := fetching.ParseMoves(game, filter.NumberOfMovesCap)
			if err != nil {
				errs <- fmt.Errorf("lichess.parseLichessPGN: could not read moves: %w", err)
				continue
			}

			userGame := fetching.UserGame{
				White:   userPlaysWhite,
				EndTime: timestamp,
				Moves:   moves,
			}
			games <- &userGame
		}
		if decoder.Err() != nil && !errors.Is(decoder.Err(), io.EOF) {
			errs <- fmt.Errorf("lichess.parseLichessPGN: %w", decoder.Err())
		}
		close(games)
		close(errs)
	}()

	return games, errs
}

func userIsWhite(game *chess.Game, username string) (bool, error) {
	whitePlayer := game.GetTagPair("White")
	blackPlayer := game.GetTagPair("Black")
	if whitePlayer == nil || blackPlayer == nil {
		return false, fmt.Errorf(
			"lichess.parseLichessPGN: could not determine players. %w",
			ErrInvalidPGNTags)
	}
	var userPlaysWhite bool
	switch strings.ToLower(username) {
	case strings.ToLower(whitePlayer.Value):
		userPlaysWhite = true
	case strings.ToLower(blackPlayer.Value):
		userPlaysWhite = false
	default:
		return false, fmt.Errorf(
			"lichess.parseLichessPGN: user does not play white or black. %w",
			ErrInvalidPGNTags)
	}
	return userPlaysWhite, nil
}

func getTimeFromGame(game *chess.Game) (time.Time, error) {
	dateTag := game.GetTagPair("UTCDate")
	timeTag := game.GetTagPair("UTCTime")
	if dateTag == nil || timeTag == nil {
		return time.Time{}, fmt.Errorf(
			"lichess.parseLichessPGN: unable to determine game time and date. %w",
			ErrInvalidPGNTags)
	}
	timeString := dateTag.Value + " " + timeTag.Value + " UTC"
	timestamp, err := time.Parse("2006.01.02 15:04:05 MST", timeString)
	if err != nil {
		return time.Time{}, fmt.Errorf(
			"lichess.parseLichessPGN: unable to parse timestamp. %w: %w",
			ErrInvalidPGNTags, err)
	}
	return timestamp, nil

}
