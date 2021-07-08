package cli

import (
	"fmt"
	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/positions"
	"github.com/spf13/cobra"
)

var PrintDateFlag bool

func NewPrintCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "print",
		Short: "print a position graph",
		Args:  cobra.ExactArgs(1),
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
}
