package pdftoolbox

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	exePath     string
	cacheFolder *string
	logger      *slog.Logger
}

type ClientOpts struct {
	CacheFolder *string
}

func New(exePath string, opts *ClientOpts) (*Client, error) {
	absPath, err := filepath.Abs(exePath)
	if err != nil {
		return nil, err
	}

	cl := &Client{
		exePath: absPath,
		logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
	}

	if opts != nil {
		if opts.CacheFolder != nil {
			cl.cacheFolder = opts.CacheFolder
		}
	}

	return cl, nil
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

func NewSetVariableArg(name string, value any) Arg {
	s := fmt.Sprintf("%s:%v", name, value)
	return Arg{arg: "--setvariable", value: &s}
}

func (cl *Client) buildProfileCommand(profile string, inputFiles []string, args ...Arg) []string {
	cmd := []string{
		profile,
	}

	for _, a := range args {
		cmd = append(cmd, a.ArgString())
	}

	for _, inputFile := range inputFiles {
		cmd = append(cmd, inputFile)
	}

	return cmd
}

func (cl *Client) RunProfile(profile string, inputFiles []string, args ...Arg) (*Result, error) {
	cmd := cl.buildProfileCommand(profile, inputFiles, args...)
	res, err := cl.runCmd(cmd...)
	if err != nil {
		return nil, err
	}

	fmt.Println(res.Command, res.Duration)

	return &Result{}, nil
}

func (cl *Client) runCmd(args ...string) (CmdOutput, error) {
	cmd := exec.Command(cl.exePath, args...)

	cl.logger.Debug("running command", slog.String("cmd", cmd.String()))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return CmdOutput{}, err
	}

	if cmd.ProcessState.ExitCode() != 0 {
		return CmdOutput{}, NewParsedError(cmd.ProcessState.ExitCode(), out)
	}

	return ParseOutput(string(out))
}

func (cl *Client) EnumerateProfiles(profileFolder string) (*EnumerateProfilesResponse, error) {
	tmpFile, err := os.CreateTemp("", "enumprofile")
	if err != nil {
		return nil, err
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	_, err = cl.runCmd(
		"--format=json",
		"--enumprofiles",
		profileFolder,
		tmpFile.Name(),
	)
	if err != nil {
		return nil, err
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
	Command  string
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
