package pdftoolbox

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type PDFToolboxClient interface {
	RunProfile(profile string, inputFiles []string, args ...Arg) (*Result, error)
	EnumerateProfiles(profileFolder string) (*EnumerateProfilesResponse, error)
}

type Client struct {
	exePath       string
	cacheFolder   *string
	profileFolder *string
	logger        *slog.Logger
}

var _ PDFToolboxClient = &Client{}

type ClientOpts struct {
	CacheFolder   *string
	ProfileFolder *string
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
		if opts.ProfileFolder != nil {
			cl.profileFolder = opts.ProfileFolder
		}
	}

	return cl, nil
}

type ArgString interface {
	ArgString() Arg
}

type Arg struct {
	Arg   string  `json:"arg"`
	Value *string `json:"value"`
}

func (a Arg) ArgString() string {
	if a.Value == nil {
		return a.Arg
	}

	return fmt.Sprintf("%s=%s", a.Arg, *a.Value)
}

func NewTimeoutArg(dur time.Duration) Arg {
	s := fmt.Sprintf("%.0f", math.Ceil(dur.Seconds()))
	return Arg{Arg: "--timeout", Value: &s}
}

func NewSetVariableArg(name string, value any) Arg {
	s := fmt.Sprintf("%s:%v", name, value)
	return Arg{Arg: "--setvariable", Value: &s}
}

func NewOutputFolderArg(dir string) Arg {
	return Arg{Arg: "--outputfolder", Value: &dir}
}

func (cl *Client) buildProfileCommand(profile string, inputFiles []string, args ...Arg) []string {
	cmd := []string{}

	for _, a := range args {
		cmd = append(cmd, a.ArgString())
	}

	if cl.profileFolder != nil && filepath.IsLocal(profile) {
		cmd = append(cmd, path.Join(*cl.profileFolder, profile))
	} else {
		cmd = append(cmd, profile)
	}

	for _, inputFile := range inputFiles {
		cmd = append(cmd, inputFile)
	}

	return cmd
}

// RunProfile uses profile in the form of myprofile.kpfx (though the file extension is not checked for)
func (cl *Client) RunProfile(profile string, inputFiles []string, args ...Arg) (*Result, error) {
	cmd := cl.buildProfileCommand(profile, inputFiles, args...)
	res, err := cl.runCmd(cmd...)

	return &Result{
		Duration:     res.Duration,
		Command:      res.Command,
		RawOutput:    res.Raw,
		ParsedOutput: res.Lines,
	}, err
}

func (cl *Client) runCmd(args ...string) (CmdOutput, error) {
	startedAt := time.Now()
	cmd := exec.Command(cl.exePath, args...)

	cl.logger.Debug("running command", slog.String("cmd", cmd.String()))

	out, err := cmd.CombinedOutput()
	if err != nil {
		return CmdOutput{
			Raw: string(out),
		}, NewParsedError(cmd.ProcessState.ExitCode(), out)
	}

	if cmd.ProcessState.ExitCode() != 0 {
		return CmdOutput{
			Raw: string(out),
		}, NewParsedError(cmd.ProcessState.ExitCode(), out)
	}

	output, err := ParseOutput(string(out))
	if err != nil {
		return CmdOutput{
			Raw: string(out),
		}, err
	}
	output.Duration = time.Since(startedAt)
	output.ExitCode = cmd.ProcessState.ExitCode()

	return output, nil
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
	RawOutput       string
}

func (p *ParsedError) Error() string {

	return ""
}

func NewParsedError(exitCode int, output []byte) *ParsedError {
	pe := &ParsedError{
		ProcessExitCode: exitCode,
		b:               output,
		s:               string(output),
		RawOutput:       string(output),
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
	Raw      string
	ExitCode int
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
	cmdOutput.Raw = s

	return cmdOutput, nil
}
