package cli

import (
	"fmt"
	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/fetching/chesscom"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "openinganalyzer",
	Short: "Fetches your games and analyzes your openings",
	Long: `Chess Opening Analyzer fetches your games from popular online chess platforms,
builds a position graph, analyzes it with a UCI engine of your choice and provides you with
information on what are your weak moves in terms of precision, not just won/drawn/lost percentage.`,
	Args: cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			_ = cmd.Help()
			os.Exit(0)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// TODO: flags - color, workers
	// TODO: set default time period to current month
	// fetch
	fetchCmd := NewFetchCommand(FetchCmdConfig{ChessComUrl: chesscom.ChessComPubAPIUrl})
	rootCmd.AddCommand(fetchCmd)
	// print
	printCmd := NewPrintCmd()
	rootCmd.AddCommand(printCmd)
}
