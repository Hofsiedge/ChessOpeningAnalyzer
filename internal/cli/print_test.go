package cli

import (
	"bytes"
	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/fetching"
	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/positions"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestPrint(t *testing.T) {
	cmd := NewPrintCmd()
	buffer := new(bytes.Buffer)
	cmd.SetOut(buffer)

	graph, _ := positions.NewPositionGraph(3)
	moves := []string{"e4 e5 Nf3 Nc6 d4 exd4",
		"e4 c5 c3 Nc6 d4 cxd4",
		"e4 e5 Nf3 Nf6 Nc3 Nc6",
	}
	for _, variation := range moves {
		game := fetching.UserGame{
			White:   true,
			EndTime: time.Date(2021, 7, 8, 0, 0, 0, 0, time.UTC),
			Moves:   strings.Split(variation, " "),
		}
		if err := graph.AddGame(game); err != nil {
			t.Error(err)
			return
		}
	}
	path := "../../testdata/cli/qux.bin"
	// wd, _ := os.Getwd()
	// fmt.Println(wd)
	if err := positions.DumpGraph(graph, path); err != nil {
		t.Error(err)
		return
	}
	cmd.SetArgs([]string{path})
	if err := cmd.Execute(); err != nil {
		t.Error(err)
		return
	}

	file, err := os.OpenFile("../../testdata/cli/sample.txt", os.O_RDONLY, 666)
	if err != nil {
		t.Error(err)
		return
	}
	defer file.Close()

	expected, _ := io.ReadAll(file)
	expectedStr := string(expected)
	got := buffer.String()
	if got != expectedStr {
		t.Errorf("results do not match")
	}
}
