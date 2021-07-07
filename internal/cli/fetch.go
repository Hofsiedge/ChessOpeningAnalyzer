package cli

import (
	"fmt"
	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/fetching"
	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/fetching/chesscom"
	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/positions"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var (
	OutputFlag  string
	MoveCapFlag int
)
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch your games from an online chess platform",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		platform := args[0]
		var fetcher fetching.GameFetcher
		switch platform {
		case "chesscom":
			fetcher = &chesscom.Fetcher{
				URL: chesscom.ChessComPubAPIUrl,
			}
		default:
			fmt.Println("Only chess.com is supported for now.")
			os.Exit(0)
		}
		username := args[1]
		// TODO: color, number of moves
		filter := fetching.FilterOptions{}
		var err error
		for i, field := range []*time.Time{&filter.TimePeriodStart, &filter.TimePeriodEnd} {
			*field, err = time.Parse("2006-01-02", args[2+i])
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
				os.Exit(0)
			}
		}
		filter.NumberOfMovesCap = MoveCapFlag
		// TODO: set default time period to current month
		fmt.Printf("Would fetch data from chesscom with params: %v, %v\n", username, filter)
		// TODO: workers flag
		var games []*fetching.UserGame
		if games, err = fetcher.Fetch(username, filter, 1); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "error fetching games: %v\n", err)
			os.Exit(0)
		}
		graph, _ := positions.NewPositionGraph(5)
		for _, game := range games {
			if err = graph.AddGame(*game); err != nil {
				fmt.Printf("Error adding a game to the graph: %v\n", err)
			}
		}
		fmt.Printf("Dumping a position graph to %v\n", OutputFlag)
		if err = positions.DumpGraph(graph, OutputFlag); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(0)
		}
		fmt.Println("Successfully saved a position graph!")
	},
}

func init() {
	fetchCmd.Flags().StringVarP(&OutputFlag, "output", "o", "opening_graph.bin", "TODO")
	fetchCmd.Flags().IntVarP(&MoveCapFlag, "moves", "m", 5, "TODO")
	rootCmd.AddCommand(fetchCmd)
}