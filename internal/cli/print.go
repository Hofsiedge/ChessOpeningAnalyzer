package cli

import (
	"fmt"
	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/positions"
	"github.com/spf13/cobra"
)

var PrintDateFlag bool

func NewPrintCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "print [path]",
		Short: "print a position graph",
		Example: `$ openinganalyzer print openings.out -d
  Print out a move tree of the position graph stored in openings.out
  with dates next to leaf-moves`,
		ValidArgs: []string{"path"},
		Args:      cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]
			graph, err := positions.LoadGraph(path)
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(cmd.OutOrStdout(), graph.Print(PrintDateFlag))
			return err
		},
	}
	cmd.Flags().BoolVarP(&PrintDateFlag, "dates", "d", false, "print out the last date for each position")
	return cmd
}
