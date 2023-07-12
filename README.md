| main | development |
|:----:|:-----------:|
|[![test & coverage](https://github.com/Hofsiedge/ChessOpeningAnalyzer/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/Hofsiedge/ChessOpeningAnalyzer/actions/workflows/test.yml)|[![test & coverage](https://github.com/Hofsiedge/ChessOpeningAnalyzer/actions/workflows/test.yml/badge.svg?branch=development)](https://github.com/Hofsiedge/ChessOpeningAnalyzer/actions/workflows/test.yml)|
|[![codecov](https://codecov.io/gh/Hofsiedge/ChessOpeningAnalyzer/branch/main/graph/badge.svg?token=JNGF6F0B7C)](https://codecov.io/gh/Hofsiedge/ChessOpeningAnalyzer)|[![codecov](https://codecov.io/gh/Hofsiedge/ChessOpeningAnalyzer/branch/development/graph/badge.svg?token=JNGF6F0B7C)](https://codecov.io/gh/Hofsiedge/ChessOpeningAnalyzer)|
|[![golangci-lint](https://github.com/Hofsiedge/ChessOpeningAnalyzer/actions/workflows/golangci-lint.yml/badge.svg?branch=main)](https://github.com/Hofsiedge/ChessOpeningAnalyzer/actions/workflows/golangci-lint.yml)|[![golangci-lint](https://github.com/Hofsiedge/ChessOpeningAnalyzer/actions/workflows/golangci-lint.yml/badge.svg?branch=development)](https://github.com/Hofsiedge/ChessOpeningAnalyzer/actions/workflows/golangci-lint.yml)|
# Chess Opening Analyzer v0.1
## Installation
```sh
go install https://github.com/Hofsiedge/ChessOpeningAnalyzer
```

## Sample usage
```
$ openinganalyzer fetch chesscom Hofsiedge 2021-07-01 2021-07-10 -o openings.out -m 3
Dumping a position graph to openings.out
Successfully saved a position graph!

$ openinganalyzer print openings.out                                                 
Position graph.
Depth: 3
White positions:
└─── e4
      ├─── d5
      │     └─── exd5
      ├─── e5
      │     └─── Nf3
      ├─── d6
      │     ├─── Nf3
      │     ├─── d4
      │     ├─── Bc4
      │     └─── Nc3
      ├─── c6
      │     ├─── Nf3
      │     └─── d4
      ├─── c5
      │     └─── Nf3
      ├─── Nf6
      │     └─── e5
      ├─── b6
      │     └─── Nf3
      └─── Nc6
            └─── d4

Black positions:
├─── d4
│     └─── d6
│           └─── c4
├─── Nf3
│     └─── d6
│           └─── g3
└─── b3
      └─── d6
            └─── Bb2

```
## Help on implemented commands
```
$ openinganalyzer
Chess Opening Analyzer fetches your games from popular online chess platforms,
builds a position graph, analyzes it with a UCI engine of your choice and provides you with
information on what are your weak moves in terms of precision, not just won/drawn/lost percentage.

Usage:
  openinganalyzer [flags]
  openinganalyzer [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  fetch       fetch your games from an online chess platform
  help        Help about any command
  print       print a position graph

Flags:
  -h, --help   help for openinganalyzer

Use "openinganalyzer [command] --help" for more information about a command.
```
```
$ openinganalyzer help fetch
fetch your games from an online chess platform

Usage:
  openinganalyzer fetch [platform] [username] [start_date] [end_date] [flags]

Examples:
  openinganalyzer fetch chesscom YourUsername 2021-10-01 2021-12-31 -m 5
	Fetch from chess.com, username - YourUsername, start_date - 01.10.2021,
	end_date - 31.12.2021, number of moves - 5

Flags:
  -h, --help            help for fetch
  -m, --moves int       how deep you want a position graph to be (default 5)
  -o, --output string   output file (default "openings.out")
```
```
$ openinganalyzer help print
print a position graph

Usage:
  openinganalyzer print [path] [flags]

Examples:
  openinganalyzer print openings.out -d
	Print out a move tree of the position graph stored in openings.out
	with dates next to leaf-moves

Flags:
  -d, --dates   print out the last date for each position
  -h, --help    help for print
```

# Coming soon
* **Commands**
  * `eval` - evaluate a position graph with a UCI engine (e.g. Stockfish)
  * `merge` - merge several position graphs into one
  * `viz` - visualize a position graph with graphviz
* **Format**
  * Web