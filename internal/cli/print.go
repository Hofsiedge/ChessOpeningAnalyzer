package cli

import (
	"fmt"
	"github.com/Hofsiedge/ChessOpeningAnalyzer/internal/positions"
	"github.com/spf13/cobra"
	"os"
)

var PrintDateFlag bool

var printCmd = &cobra.Command{
	Use:   "print",
	Short: "print a position graph",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		graph, err := positions.LoadGraph(path)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(0)
		}
		fmt.Println(graph.Print(PrintDateFlag))
	},
}
