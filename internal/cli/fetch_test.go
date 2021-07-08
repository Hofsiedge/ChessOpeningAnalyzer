package cli

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

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
}
