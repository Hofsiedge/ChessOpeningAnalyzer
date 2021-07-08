package cli

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestFetchArguments(t *testing.T) {
	cmd := NewFetchCommand(FetchCmdConfig{ChessComUrl: ""})
	outBuffer := new(bytes.Buffer)
	cmd.SetOut(outBuffer)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"lichess", "", "", ""})
	if err := cmd.Execute(); err == nil || !strings.HasPrefix(err.Error(), "only chess.com is supported") {
		t.Errorf("expected an \"unsupported platform\" message, got %v", err)
	}
	cmd.SetArgs([]string{"chesscom", "quUx", "", ""})
	if err := cmd.Execute(); err == nil || !strings.HasPrefix(err.Error(), "error parsing a date") {
		t.Errorf("expected an \"error parsing a date\" message, got %v", err)
	}
	cmd.SetArgs(strings.Split("chesscom Hofsiedge 2021-07-01 2021-07-10 -o qux.bin -m 3", " "))
	if err := cmd.Execute(); err == nil || !strings.HasPrefix(err.Error(), "error fetching games") {
		t.Errorf("expected an \"error fetching games\" message, got %v", err)
	}
}

func TestFetchCommand(t *testing.T) {
	file, err := os.Open("../../testdata/fetching/sample_response.json")
	if err != nil {
		t.Error(err)
		return
	}
	responseBody, _ := io.ReadAll(file)
	testServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write(responseBody)
		request.Body.Close()
	}))
	cmd := NewFetchCommand(FetchCmdConfig{ChessComUrl: testServer.URL})
	buffer := new(bytes.Buffer)
	cmd.SetOut(buffer)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"chesscom", "Hofsiedge", "2021-07-01", "2021-07-10", "-o", "../../testdata/test_qux.bin", "-m", "3"})
	if err := cmd.Execute(); err != nil {
		t.Error(err)
		return
	}
	logs := buffer.String()
	if logs != `Dumping a position graph to ../../testdata/test_qux.bin
Successfully saved a position graph!
` {
		t.Errorf("Output of fetch cmd doesn't match expected format - got:\n%v", buffer.String())
	}

	cmd.SetArgs(strings.Split("chesscom Hofsiedge 2021-07-01 2021-07-10 -o non-existent-dir/qux.bin -m 3", " "))
	if err := cmd.Execute(); err == nil || !strings.HasSuffix(err.Error(), "no such file or directory") {
		t.Errorf("expected a \"no such file or directory\" error, got %v", err)
	}
}
