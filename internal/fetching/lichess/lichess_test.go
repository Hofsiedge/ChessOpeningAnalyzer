package lichess

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/fetching"
	"github.com/notnil/chess"
)

type server struct {
	Response   interface{}
	StatusCode int
	HasBody    bool
}

func (srv server) mockLichess(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	w.WriteHeader(srv.StatusCode)
	if srv.HasBody {
		switch srv.Response.(type) {
		case []byte:
			if _, err := w.Write(srv.Response.([]byte)); err != nil {
				panic(fmt.Errorf("mock server error: %w", err))
			}
		default:
			jsonBody, err := json.Marshal(srv.Response)
			if err != nil {
				panic(fmt.Errorf("mock server error: %w", err))
			}
			if _, err = w.Write(jsonBody); err != nil {
				panic(fmt.Errorf("mock server error: %w", err))
			}
		}
	}
}

type FetchArgs struct {
	Username string
	Filter   fetching.FilterOptions
	Workers  int
}

type testCase struct {
	Name             string
	Server           server
	ExpectedResponse []*fetching.UserGame
	ExpectedError    error
	Args             FetchArgs
}

func evaluateTestCases(testCases []testCase, t *testing.T) {
	for _, testCase := range testCases {
		ts := httptest.NewServer(http.HandlerFunc(testCase.Server.mockLichess))
		tsURL, _ := url.Parse(ts.URL)
		fetcher := Fetcher{URL: *tsURL}
		resp, err := fetcher.Fetch(
			testCase.Args.Username,
			testCase.Args.Filter,
			testCase.Args.Workers)
		if testCase.ExpectedError != nil {
			if err == nil {
				t.Errorf("case: %s. Expected error but got nil", testCase.Name)
			} else if !errors.Is(err, testCase.ExpectedError) {
				t.Errorf(
					"case: %s. Expected \"%v\" but got \"%v\"",
					testCase.Name, testCase.ExpectedError, err)
			}
			continue
		}
		if reflect.TypeOf(testCase.ExpectedResponse) != reflect.TypeOf(resp) {
			t.Errorf(
				"case: %s. Type mismatch. Expected \"%T\" but got \"%T\"",
				testCase.Name, testCase.ExpectedResponse, resp)
		}
		if !reflect.DeepEqual(testCase.ExpectedResponse, resp) {
			t.Errorf(
				"case: %s. Expected \"%v\" but got \"%v\"",
				testCase.Name, testCase.ExpectedResponse, resp)
		}
	}
}

func readFixture(path string) []byte {
	file, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("Could not open fixture: %s - %w\n", path, err))
	}
	var rawData []byte
	if rawData, err = io.ReadAll(file); err != nil {
		panic(fmt.Errorf("Could not read fixture: %s - %w\n", path, err))
	}
	return rawData
}

func TestLichessRequest(t *testing.T) {
	const testDataPath = "../../../testdata/fetching/lichess/"

	testCases := []testCase{{
		Name: "User not found",
		Server: server{
			StatusCode: 404,
		},
		ExpectedError: fetching.UserNotFoundError,
		Args: FetchArgs{
			Username: "non-existent_user",
		},
	}, {
		Name: "Single full game",
		Server: server{
			StatusCode: 200,
			HasBody:    true,
			Response:   readFixture(testDataPath + "single_game.pgn"),
		},
		ExpectedResponse: []*fetching.UserGame{{
			White:   true,
			EndTime: time.Date(2023, 6, 1, 1, 2, 3, 0, time.UTC),
			Moves: []string{
				"e4", "e5", "Nf3", "Nf6", "Nc3", "Nc6", "d4", "d6", "d5",
				"Ne7", "Bb5+", "Bd7", "Bxd7+", "Qxd7", "Be3", "Qg4",
				"O-O", "Rd8", "Bxa7", "Nfxd5", "Nxd5", "Nxd5", "exd5",
				"Qb4", "b3", "Qa5", "Be3", "Rd7", "Bd2", "Qxd5", "c4",
				"Qe4", "Ng5", "Qg6", "h4", "Be7", "h5", "Qd3", "Nh3",
				"d5", "Re1", "Rd6", "Re3", "Qd4", "Qc1", "Qg4", "Qa3",
				"O-O", "Rxe5", "Rg6", "Qxe7", "Qxg2#",
			},
		}},
		Args: FetchArgs{
			Username: "Player1",
			Filter: fetching.FilterOptions{
				TimePeriodStart:  time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
				TimePeriodEnd:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				Color:            chess.White,
				NumberOfMovesCap: 0,
			},
		},
	}}
	evaluateTestCases(testCases, t)
}
