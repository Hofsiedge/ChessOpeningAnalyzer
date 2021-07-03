package fetcher

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

type server struct {
	Response   interface{}
	StatusCode int
	HasBody    bool
}

type testCase struct {
	Server           server
	Params           ChessComFetchParams
	ExpectedResponse []ChessComGame
	IsErr            bool
	ExpectedError    string
}

func evaluateTestCases(testCases []testCase, t *testing.T) {
	for i, testCase := range testCases {
		ts := httptest.NewServer(http.HandlerFunc(testCase.Server.mockChessCom))
		fetcher := ChessComFetcher{URL: ts.URL}
		resp, err := fetcher.Fetch(testCase.Params)
		if testCase.IsErr {
			if err == nil {
				t.Errorf("case %v. Expected error but got nil", i)
			} else if !strings.HasPrefix(err.Error(), testCase.ExpectedError) {
				t.Errorf("case %v. Expected \"%v\" but got \"%v\"", i, testCase.ExpectedError, err)
			}
			continue
		}
		if !reflect.DeepEqual(testCase.ExpectedResponse, resp) {
			t.Errorf("case %v. Expected \"%v\" but got \"%v\"", i, testCase.ExpectedResponse, resp)
		}
	}
}

func (srv server) mockChessCom(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(srv.StatusCode)
	if srv.HasBody {
		switch srv.Response.(type) {
		case []byte:
			w.Write(srv.Response.([]byte))
		default:
			jsonBody, _ := json.Marshal(srv.Response)
			w.Write(jsonBody)
		}
	}
	r.Body.Close()
}

func TestChessComRequest(t *testing.T) {
	testCases := []testCase{{
		Server: server{
			Response:   ChessComErrorResponse{0, "Date cannot be set in the future"},
			StatusCode: 404,
			HasBody:    true,
		},
		Params: ChessComFetchParams{
			userName: "Hofsiedge",
			year:     2100,
			month:    10,
		},
		IsErr:         true,
		ExpectedError: "non-OK StatusCode: 404; error: {0 Date cannot be set in the future}",
	}, {
		Server: server{
			Response:   ChessComErrorResponse{0, "User \\\"NonExistentUser\\\" not found."},
			StatusCode: 404,
			HasBody:    true,
		},
		Params: ChessComFetchParams{
			userName: "NonExistentUser",
			year:     2020,
			month:    6,
		},
		IsErr:         true,
		ExpectedError: "non-OK StatusCode: 404; error: {0 User \\\"NonExistentUser\\\" not found.}",
	}}
	evaluateTestCases(testCases, t)
}

func TestChessComUnmarshalling(t *testing.T) {
	testCases := []struct {
		fixture       string
		isError       bool
		expectedError string
		want          []ChessComGame
		name          string
	}{{
		fixture: "../../testdata/fetcher/empty.json",
		want:    []ChessComGame{},
		name:    "UnmarshalEmpty",
	}, {
		fixture: "../../testdata/fetcher/trivial.json",
		want: []ChessComGame{{
			Url:         "game_url",
			Pgn:         "PGN",
			TimeControl: "60",
			EndTime:     1622664410,
			TimeClass:   "bullet",
			Rules:       "chess",
			White:       ChessComUser{Rating: 525, Result: "timeout", Id: "qux", Username: "qux"},
			Black:       ChessComUser{Rating: 595, Result: "win", Id: "buzz", Username: "buzz"},
		}},
		name: "UnmarshalTrivial",
	}}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				file, err := os.Open(testCase.fixture)
				if err != nil {
					log.Fatalf("Could not open fixture: %v - %v\n", testCase.fixture, err)
				}
				var rawData []byte
				if rawData, err = io.ReadAll(file); err != nil {
					log.Fatalf("Could not read fixture: %v - %v", testCase.fixture, err)
				}
				_, _ = w.Write(rawData)
				_ = r.Body.Close()
			}))
			fetcher := ChessComFetcher{URL: ts.URL}
			games, err := fetcher.Fetch(ChessComFetchParams{})
			if err == nil && testCase.isError {
				t.Errorf("Expected error, got nil")
			}
			if err != nil {
				if err.Error() != testCase.expectedError {
					t.Errorf("Expected \"%v\" but got \"%v\"", testCase.expectedError, err)
				}
				return
			}
			if !reflect.DeepEqual(testCase.want, games) {
				t.Errorf("case %v. Expected \"%v\" but got \"%v\"", testCase.name, testCase.want, games)
			}
		})
	}
}
