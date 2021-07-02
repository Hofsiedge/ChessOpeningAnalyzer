package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ChessComErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ChessComFetcher struct {
	URL string
}

type ChessComUser struct {
	Rating   int    `json:"rating"`
	Result   string `json:"result"`
	Id       string `json:"@id"`
	Username string `json:"username"`
}

type ChessComGame struct {
	Url         string `json:"url"`
	Pgn         string `json:"pgn"`
	TimeControl string `json:"time_control"`
	EndTime     int    `json:"end_time"`
	// Rated       bool         `json:"rated"`
	// Fen         string       `json:"-"`
	TimeClass string       `json:"time_class"`
	Rules     string       `json:"rules"`
	White     ChessComUser `json:"white"`
	Black     ChessComUser `json:"black"`
}

type ChessComArchive struct {
	Games []ChessComGame `json:"games"`
}

type ChessComFetchParams struct {
	userName string
	year     int
	month    int
	filter   FilterPredicate
}

type FilterPredicate func(game *ChessComGame) bool

func (f *ChessComFetcher) Fetch(p ChessComFetchParams) ([]ChessComGame, error) {
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
		apiError := new(ChessComErrorResponse)
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
	games := make([]ChessComGame, 0)

	if p.filter == nil {
		p.filter = func(game *ChessComGame) bool {
			return game.Rules != "chess"
		}
	}

	// read open bracket
	if _, err = decoder.Token(); err != nil {
		return nil, fmt.Errorf("error decoding chess game JSON: %v", err)
	}
	for decoder.More() {
		game := new(ChessComGame)
		if err = decoder.Decode(game); err != nil {
			// TODO: find out if there is a way to apply errors.Is in this situation
			if err.Error() == "not at beginning of value" {
				break
			}
			return nil, fmt.Errorf("error decoding a chess game: %v", err)
		}
		// filtering out chess variants other than standard
		if p.filter(game) {
			continue
		}
		games = append(games, *game)
	}
	if _, err = decoder.Token(); err != nil {
		return nil, fmt.Errorf("error decoding the end of chess games JSON: %v", err)
	}

	return games, nil
}
