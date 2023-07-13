package cli

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestFetchArguments(t *testing.T) {
	cmd := NewFetchCommand(FetchCmdConfig{
		ChessComURL: "",
		LichessURL:  url.URL{Scheme: "", Host: ""},
	})
	outBuffer := new(bytes.Buffer)
	cmd.SetOut(outBuffer)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"icc", "", "", ""})
	if err := cmd.Execute(); err == nil || !errors.Is(err, ErrUnsupportedPlatform) {
		t.Errorf("expected \"%v\" error, got \"%v\"", ErrUnsupportedPlatform, err)
	}
	cmd.SetArgs([]string{"chesscom", "quUx", "", ""})
	if err := cmd.Execute(); err == nil || !errors.Is(err, ErrInvalidDate) {
		t.Errorf("expected \"%v\" error, got \"%v\"", ErrInvalidDate, err)
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
	cmd := NewFetchCommand(FetchCmdConfig{ChessComURL: testServer.URL})
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
