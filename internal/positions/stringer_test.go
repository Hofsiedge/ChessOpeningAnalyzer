package positions

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func TestMove_String(t *testing.T) {
	move := &Move{
		To: &PositionNode{
			Position: &Position{
				FEN:       "",
				Score:     0,
				Evaluated: false,
			},
			LastPlayed: time.Time{},
			Moves:      nil,
		},
		Move: "e4",
	}
	if moveStr := fmt.Sprint(move); moveStr != "e4" {
		t.Errorf("Expect \"e4\", got \"%v\"", moveStr)
	}
	move.To.Position.Evaluated = true
	move.To.Position.Score = 1.3
	if moveStr := fmt.Sprint(move); moveStr != "e4    -> 1.3" {
		t.Errorf("Expect \"e4    -> 1.3\", got \"%v\"", moveStr)
	}
}

func TestPositionNode_print(t *testing.T) {
	type fields struct {
		Position   *Position
		LastPlayed time.Time
		Moves      []*Move
	}
	type args struct {
		prefix    string
		printDate bool
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantOut string
	}{{
		name: "TestTrivialPositionNodeString",
		fields: fields{
			Position: &Position{
				FEN:       "",
				Score:     0,
				Evaluated: false,
			},
			LastPlayed: time.Time{},
			Moves: []*Move{{
				To: &PositionNode{
					Position:   &Position{},
					LastPlayed: time.Now(),
					Moves:      nil,
				},
				Move: "e4",
			}, {
				To: &PositionNode{
					Position:   &Position{},
					LastPlayed: time.Now(),
					Moves:      nil,
				},
				Move: "d4",
			}},
		},
		args:    args{"", false},
		wantOut: "├─── e4\n└─── d4\n",
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &PositionNode{
				Position:   tt.fields.Position,
				LastPlayed: tt.fields.LastPlayed,
				Moves:      tt.fields.Moves,
			}
			out := &bytes.Buffer{}
			n.print(out, tt.args.prefix, tt.args.printDate)
			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("print() = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}
