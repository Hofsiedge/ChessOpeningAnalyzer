package cli

import (
	"fmt"
	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/fetching"
	"github.com/spf13/cobra"
	"os"
	"time"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch your games from an online chess platform",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		/*
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
		*/
		username := args[1]
		// TODO: color, number of moves
		filter := fetching.FilterOptions{}
		var err error
		i := 2
		for _, field := range []*time.Time{&filter.TimePeriodStart, &filter.TimePeriodEnd} {
			*field, err = time.Parse("2006-01-02", args[i])
			if err != nil {
				_, _ = fmt.Fprintln(os.Stderr, err)
				os.Exit(0)
			}
			i++
		}
		// TODO: parse time period as a flag with default = this month
		fmt.Printf("Would fetch data from chesscom with params: %v, %v-%v\n", username, filter.TimePeriodStart, filter.TimePeriodEnd)
	},
}

func init() {
	rootCmd.AddCommand(fetchCmd)
}
