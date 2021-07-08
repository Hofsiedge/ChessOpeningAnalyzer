package chesscom

import (
	"encoding/json"
	"fmt"
	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/fetching"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

const ChessComPubAPIUrl = "https://api.chess.com/pub"

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Fetcher struct {
	URL string
}

type monthYearPair struct {
	Year  int
	Month int
}

func (f *Fetcher) workerPool(workers int, jobs <-chan monthYearPair, username string, filter fetching.FilterOptions) (<-chan []*fetching.UserGame, <-chan error) {
	results := make(chan []*fetching.UserGame, workers)
	errors := make(chan error, workers)

	wg := sync.WaitGroup{}
	wg.Add(workers)
	// worker pool
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for p := range jobs {
				games, err := f.fetchMonthGames(fetchParams{
					userName: username,
					year:     p.Year,
					month:    p.Month,
					until:    filter.NumberOfMovesCap,
				})
				if err != nil {
					errors <- fmt.Errorf("error parsing %d.%02d games of %v: %v", p.Year, p.Month, username, err)
					continue
				}
				results <- games
			}
		}()
	}
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()
	return results, errors
}

func (f *Fetcher) aggregate(results <-chan []*fetching.UserGame, errs <-chan error) (chan []*fetching.UserGame, chan error) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	// game aggregator
	games := make([]*fetching.UserGame, 0)
	go func() {
		for result := range results {
			games = append(games, result...)
		}
		wg.Done()
	}()
	// error aggregator
	aggregatedErrors := make([]string, 0)
	go func() {
		for err := range errs {
			aggregatedErrors = append(aggregatedErrors, err.Error())
		}
		wg.Done()
	}()

	gamesCh, errCh := make(chan []*fetching.UserGame, 1), make(chan error, 1)
	go func() {
		wg.Wait()
		gamesCh <- games
		if len(aggregatedErrors) != 0 {
			errCh <- fmt.Errorf(strings.Join(aggregatedErrors, "\n"))
		} else {
			errCh <- nil
		}
	}()
	return gamesCh, errCh
}

func (f *Fetcher) Fetch(username string, filter fetching.FilterOptions, workers int) ([]*fetching.UserGame, error) {
	jobs := make(chan monthYearPair, workers)
	results, errs := f.workerPool(workers, jobs, username, filter)
	games, err := f.aggregate(results, errs)

	// sending jobs
	currentDate := time.Date(filter.TimePeriodStart.Year(), filter.TimePeriodStart.Month(), 1, 0, 0, 0, 0, time.UTC)
	for currentDate.Before(filter.TimePeriodEnd) {
		year, month, _ := currentDate.Date()
		jobs <- monthYearPair{year, int(month)}
		currentDate = currentDate.AddDate(0, 1, 0)
	}
	close(jobs)
	return <-games, <-err
}

type User struct {
	Rating   int    `json:"rating"`
	Result   string `json:"result"`
	Id       string `json:"@id"`
	Username string `json:"username"`
}

type Game struct {
	Url         string `json:"url"`
	Pgn         string `json:"pgn"`
	TimeControl string `json:"time_control"`
	EndTime     int64  `json:"end_time"`
	// Rated       bool         `json:"rated"`
	// Fen         string       `json:"-"`
	TimeClass string `json:"time_class"`
	Rules     string `json:"rules"`
	White     User   `json:"white"`
	Black     User   `json:"black"`
}

func (g Game) UserGame(username string, until int) (*fetching.UserGame, error) {
	moves, err := fetching.ParseMoves(g.Pgn+"\n", until)
	if err != nil {
		return nil, err
	}
	game := &fetching.UserGame{
		White:   g.White.Username == username,
		EndTime: time.Unix(g.EndTime, 0),
		Moves:   moves,
	}
	return game, nil
}

type filterPredicate func(game *Game) bool

type fetchParams struct {
	userName string
	year     int
	month    int
	filter   filterPredicate
	until    int
}

func (f *Fetcher) fetchMonthGames(p fetchParams) ([]*fetching.UserGame, error) {
	var (
		err      error
		response *http.Response
	)
	if response, err = http.Get(
		fmt.Sprintf("%v/player/%v/games/%d/%02d", f.URL, strings.ToLower(p.userName), p.year, p.month),
	); err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		apiError := new(ErrorResponse)
		rawData, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("network error: %v", err)
		}
		if err = json.Unmarshal(rawData, apiError); err != nil {
			return nil, fmt.Errorf("could not unmarshal an error response")
		}
		return nil, fmt.Errorf("non-OK StatusCode: %v; error: %v", response.StatusCode, *apiError)
	}

	// response.Body should be treated as a stream due to potentially big number of games
	decoder := json.NewDecoder(response.Body)

	if p.filter == nil {
		p.filter = func(game *Game) bool {
			return game.Rules != "chess"
		}
	}
	var chessComGames []*Game
	chessComGames, err = parseChessComGames(decoder, p.filter)
	if err != nil {
		return nil, err
	}
	genericGames := make([]*fetching.UserGame, 0, len(chessComGames))
	for _, game := range chessComGames {
		userGame, err := game.UserGame(p.userName, p.until)
		if err != nil {
			return nil, err
		}
		genericGames = append(genericGames, userGame)
	}
	return genericGames, nil
}

func parseChessComGames(decoder *json.Decoder, filter filterPredicate) ([]*Game, error) {
	// read `{"games":`
	for i := 0; i < 3; i++ {
		if t, err := decoder.Token(); err != nil {
			return nil, fmt.Errorf("error reading JSON token (%v) from the start of a game archive: %v", t, err)
		}
	}
	games := make([]*Game, 0)
	for decoder.More() {
		game := new(Game)
		if err := decoder.Decode(game); err != nil {
			// TODO: find out if there is a way to apply errors.Is in this situation
			if err.Error() == "not at beginning of value" {
				break
			}
			return nil, fmt.Errorf("error decoding a chess game: %v", err)
		}
		// filtering out chess variants other than standard
		if filter(game) {
			continue
		}
		games = append(games, game)
	}
	if _, err := decoder.Token(); err != nil {
		return nil, fmt.Errorf("error decoding the end of chess games JSON: %v", err)
	}

	return games, nil
}
