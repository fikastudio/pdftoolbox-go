package pdftoolbox

import "time"

type Result struct {
	Command      string
	RawOutput    string
	ParsedOutput []CmdOutputLine
	Duration     time.Duration
	ExitCode     int
}
