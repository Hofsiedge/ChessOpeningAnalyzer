package cli

import (
	"fmt"
	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/fetching"
	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/fetching/chesscom"
	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/positions"
	"github.com/spf13/cobra"
	"time"
)

var (
	FetchOutputFlag string
	MoveCapFlag     int
)

type FetchCmdConfig struct {
	ChessComUrl string
}

func NewFetchCommand(cfg FetchCmdConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "fetch your games from an online chess platform",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			platform := args[0]
			var fetcher fetching.GameFetcher
			switch platform {
			case "chesscom":
				fetcher = &chesscom.Fetcher{
					URL: cfg.ChessComUrl,
				}
			default:
				return fmt.Errorf("only chess.com is supported for now")
			}
			username := args[1]
			filter := fetching.FilterOptions{}
			var err error
			for i, field := range []*time.Time{&filter.TimePeriodStart, &filter.TimePeriodEnd} {
				*field, err = time.Parse("2006-01-02", args[2+i])
				if err != nil {
					return fmt.Errorf("error parsing a date: %v", err)
				}
			}
			filter.NumberOfMovesCap = MoveCapFlag
			var games []*fetching.UserGame
			if games, err = fetcher.Fetch(username, filter, 1); err != nil {
				return fmt.Errorf("error fetching games: %v\n", err)
			}
			graph, _ := positions.NewPositionGraph(MoveCapFlag)
			for _, game := range games {
				if err = graph.AddGame(*game); err != nil {
					if _, oerr := fmt.Fprintf(cmd.OutOrStdout(), "Error adding a game to the graph: %v\n", err); oerr != nil {
						return oerr
					}
				}
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Dumping a position graph to %v\n", FetchOutputFlag); err != nil {
				return err
			}
			if err = positions.DumpGraph(graph, FetchOutputFlag); err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), "Successfully saved a position graph!")
			return err
		},
	}
	cmd.Flags().StringVarP(&FetchOutputFlag, "output", "o", "openings.out", "output file")
	cmd.Flags().IntVarP(&MoveCapFlag, "moves", "m", 5, "how deep you want a position graph to be")
	return cmd
}
