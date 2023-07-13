package cli

import (
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/fetching"
	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/fetching/chesscom"
	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/fetching/lichess"
	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/positions"
	"github.com/spf13/cobra"
)

var (
	FetchOutputFlag string
	MoveCapFlag     int
)

var (
	ErrUnsupportedPlatform = errors.New("unsupported platform")
	ErrInvalidDate         = errors.New("invalid date")
	ErrFetchingError       = errors.New("data fetching error")
)

type FetchCmdConfig struct {
	ChessComURL string
	LichessURL  url.URL
}

func NewFetchCommand(cfg FetchCmdConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:        "fetch platform username start_date end_date [-m number_of_moves]",
		SuggestFor: []string{"etch", "ftch", "fech", "fetc", "feth", "get", "download"},
		Short:      "fetch your games from an online chess platform",
		Long: `fetch your games from an online chess platform (chesscom/lichess).
dates are specified in YYYY-MM-DD format. optionally accepts number of moves as -m flag`,
		ValidArgs: []string{"platform", "username", "start_date", "end_date"},
		Example: `$ openinganalyzer fetch chesscom YourUsername 2021-10-01 2021-12-31 -m 5
  Fetch from chess.com, username - YourUsername, start_date - 01.10.2021,
  end_date - 31.12.2021, number of moves - 5`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			platform := args[0]
			var fetcher fetching.GameFetcher
			switch platform {
			case "chesscom":
				fetcher = &chesscom.Fetcher{
					URL: cfg.ChessComURL,
				}
			case "lichess":
				fetcher = &lichess.Fetcher{
					URL: cfg.LichessURL,
				}
			default:
				return fmt.Errorf("%w: %s. Only chesscom and lichess are supported for now", ErrUnsupportedPlatform, platform)
			}
			username := args[1]
			filter := fetching.FilterOptions{}
			var err error
			for i, field := range []*time.Time{&filter.TimePeriodStart, &filter.TimePeriodEnd} {
				*field, err = time.Parse("2006-01-02", args[2+i])
				if err != nil {
					return fmt.Errorf("%w (%s): %w", ErrInvalidDate, field, err)
				}
			}
			filter.NumberOfMovesCap = MoveCapFlag
			var games []*fetching.UserGame
			if games, err = fetcher.Fetch(username, filter, 1); err != nil {
				return fmt.Errorf("%w: %w", ErrFetchingError, err)
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
