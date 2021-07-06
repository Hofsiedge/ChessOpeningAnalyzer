package chesscom

import (
	"encoding/json"
	"fmt"
	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/fetcher"
	"io"
	"net/http"
	"strings"
	"time"
)

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Fetcher struct {
	URL string
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

func (g Game) UserGame(username string, until int) (*fetcher.UserGame, error) {
	moves, err := fetcher.ParseMoves(g.Pgn+"\n", until)
	if err != nil {
		return nil, err
	}
	game := &fetcher.UserGame{
		White:   g.White.Username == username,
		EndTime: time.Unix(g.EndTime, 0),
		Moves:   moves,
	}
	return game, nil
}

type Archive struct {
	Games []Game `json:"games"`
}

type FetchParams struct {
	userName string
	year     int
	month    int
	filter   FilterPredicate
	until    int
}

type FilterPredicate func(game *Game) bool

func (f *Fetcher) Fetch(p FetchParams) ([]*fetcher.UserGame, error) {
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
	genericGames := make([]*fetcher.UserGame, 0, len(chessComGames))
	for _, game := range chessComGames {
		userGame, err := game.UserGame(p.userName, p.until)
		if err != nil {
			return nil, err
		}
		genericGames = append(genericGames, userGame)
	}
	return genericGames, nil
}

func parseChessComGames(decoder *json.Decoder, filter FilterPredicate) ([]*Game, error) {
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
