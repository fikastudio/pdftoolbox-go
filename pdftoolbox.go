package pdftoolbox

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	exePath     string
	cacheFolder *string
}

type ClientOpts struct {
	CacheFolder *string
}

func New(exePath string, opts *ClientOpts) *Client {
	cl := &Client{
		exePath: exePath,
	}

	if opts != nil {
		if opts.CacheFolder != nil {
			cl.cacheFolder = opts.CacheFolder
		}
	}

	return cl
}

type ArgString interface {
	ArgString() Arg
}

type Arg struct {
	arg   string
	value *string
}

func (a Arg) ArgString() string {
	if a.value == nil {
		return a.arg
	}

	return fmt.Sprintf("%s=%s", a.arg, *a.value)
}

func NewTimeoutArg(dur time.Duration) Arg {
	s := fmt.Sprintf("%.0f", math.Ceil(dur.Seconds()))
	return Arg{arg: "--timeout", value: &s}
}

func (cl *Client) RunProfile(profile string, inputFiles []string, args ...Arg) (*Result, error) {
	cmd := []string{
		cl.exePath,
	}

	for _, a := range args {
		cmd = append(cmd, a.ArgString())
	}

	for _, inputFile := range inputFiles {
		cmd = append(cmd, inputFile)
	}

	fullCommand := strings.Join(cmd, " ")

	return &Result{
		Command: fullCommand,
	}, nil
}

func (cl *Client) EnumerateProfiles(profileFolder string) (*EnumerateProfilesResponse, error) {
	tmpFile, err := os.CreateTemp("", "enumprofile")
	if err != nil {
		return nil, err
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	cmd := exec.Command(cl.exePath, "--format=json", "--enumprofles", profileFolder, tmpFile.Name())

	if cmd.ProcessState.ExitCode() != 0 {
		out, err := cmd.CombinedOutput()
		if err != nil {
			return nil, err
		}

		return nil, NewParsedError(cmd.ProcessState.ExitCode(), out)
	}

	tmpFile, err = os.Open(tmpFile.Name())
	if err != nil {
		return nil, err
	}

	var resp EnumerateProfilesResponse
	if err = json.NewDecoder(tmpFile).Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

type ParsedError struct {
	b []byte
	s string

	ProcessExitCode int
	Code            int64
	Message         string
}

func (p *ParsedError) Error() string {

	return ""
}

func NewParsedError(exitCode int, output []byte) *ParsedError {
	pe := &ParsedError{
		ProcessExitCode: exitCode,
		b:               output,
		s:               string(output),
	}

	parsed, err := ParseOutput(string(output))
	if err != nil {
		fmt.Println(pe)
		return pe
	}

	for _, line := range parsed.Lines {
		switch line.Type() {
		case ErrorLine:
			l := line.(CmdOutputErrorLine)

			pe.Code = l.Code
			pe.Message = l.Message
		}
	}

	return pe
}

type CmdOutput struct {
	Lines    []CmdOutputLine
	Duration time.Duration
}

func ParseOutput(s string) (CmdOutput, error) {
	lines := strings.Split(s, "\n")
	var outLines []CmdOutputLine
	var cmdOutput CmdOutput

	for _, line := range lines {
		items := strings.Split(line, "\t")
		fmt.Println(strings.Join(items, "|"), len(items))

		if len(line) == 0 {
			continue
		}

		switch items[0] {
		case "Error", "Errors":
			var l CmdOutputErrorLine

			code, err := strconv.ParseInt(items[1], 10, 64)
			if err != nil {
				fmt.Println(err)
			}
			l.Code = code
			l.Message = items[2]

			outLines = append(outLines, l)
		case "Duration":
			var l CmdOutputDurationLine

			parsed, err := time.Parse("04:05", items[1])
			if err != nil {
				fmt.Println(err)
			}
			l.dur = time.Duration(parsed.Minute())*time.Minute + time.Duration(parsed.Second())*time.Second
			cmdOutput.Duration = l.dur

			outLines = append(outLines, l)
		default:
			var l CmdOutputIdentityLine
			l.Line = line
			l.Parts = items

			outLines = append(outLines, l)
		}
	}

	cmdOutput.Lines = outLines

	return cmdOutput, nil
}
