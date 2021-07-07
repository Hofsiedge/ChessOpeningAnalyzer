package positions

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// String implements fmt.Stringer interface
func (m *Move) String() string {
	score := ""
	if p := m.To.Position; p.Evaluated {
		score = fmt.Sprintf("-> %v", m.To.Position.Score)
	}
	return fmt.Sprintf("%-5v %v", m.Move, score)
}

// String implements fmt.Stringer interface
func (n *PositionNode) String() string {
	return n.Print(false)
}

func (n *PositionNode) Print(printDate bool) string {
	buffer := new(bytes.Buffer)
	n.print(buffer, "", printDate)
	return buffer.String()
}

// String implements fmt.Stringer interface
func (g *PositionGraph) String() string {
	return g.Print(false)
}

func (g *PositionGraph) Print(printDates bool) string {
	lines := make([]string, 0)
	lines = append(lines, "Position graph.", fmt.Sprintf("Depth: %v", g.Depth))
	if len(g.WhitePositions.Moves) > 0 {
		lines = append(lines, fmt.Sprintf("White positions:\n%v", g.WhitePositions.Print(printDates)))
	}
	if len(g.BlackPositions.Moves) > 0 {
		lines = append(lines, fmt.Sprintf("Black positions:\n%v", g.BlackPositions.Print(printDates)))
	}
	return strings.Join(lines, "\n")
}

func (n *PositionNode) print(out io.Writer, prefix string, printDate bool) {
	lastMoveIndex := len(n.Moves) - 1
	var leftBorder, movePrefix string
	for i, move := range n.Moves {
		if i == lastMoveIndex {
			leftBorder = " "
			movePrefix = "└───"
		} else {
			leftBorder = "│"
			movePrefix = "├───"
		}
		date := ""
		if printDate && len(move.To.Moves) == 0 {
			date = move.To.LastPlayed.Format("(02.01.2006)")
		}
		_, _ = fmt.Fprintf(out, "%v %v %v\n", prefix+movePrefix, move, date)
		move.To.print(
			out,
			prefix+leftBorder+"     ",
			printDate,
		)
	}
}
