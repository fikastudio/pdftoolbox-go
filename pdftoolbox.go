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
	RunProfile(profile string, inputFiles []string, args ...Arg) (CmdOutput, error)
	EnumerateProfiles(profileFolder string) (*EnumerateProfilesResponse, error)
}

type PDFToolboxExecutor interface {
	Command(name string, arg ...string) *exec.Cmd
	CombinedOutput(cmd *exec.Cmd) ([]byte, error)
	ExitCode(cmd *exec.Cmd) int
}

type Executor struct {
}

func NewExecutor() (*Executor, error) {
	return &Executor{}, nil
}

func (e Executor) Command(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	return cmd

}

func (e Executor) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	return cmd.CombinedOutput()
}

func (e Executor) ExitCode(cmd *exec.Cmd) int {
	return cmd.ProcessState.ExitCode()
}

type Client struct {
	executor      PDFToolboxExecutor
	exePath       string
	cacheFolder   *string
	profileFolder *string
	logger        *slog.Logger
}

var _ PDFToolboxClient = &Client{}

type ClientOpts struct {
	CacheFolder   *string
	ProfileFolder *string
	Executor      PDFToolboxExecutor
	Logger        *slog.Logger
}

func New(exePath string, opts *ClientOpts) (*Client, error) {
	absPath, err := filepath.Abs(exePath)
	if err != nil {
		return nil, err
	}

	exe, err := NewExecutor()
	if err != nil {
		return nil, err
	}

	cl := &Client{
		executor: exe,
		exePath:  absPath,
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
		if opts.Executor != nil {
			cl.executor = opts.Executor
		}
		if opts.Logger != nil {
			cl.logger = opts.Logger
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
func (cl *Client) RunProfile(profile string, inputFiles []string, args ...Arg) (CmdOutput, error) {
	cmd := cl.buildProfileCommand(profile, inputFiles, args...)
	return cl.runCmd(cmd...)
}

func (cl *Client) runCmd(args ...string) (CmdOutput, error) {
	startedAt := time.Now()
	cmd := cl.executor.Command(cl.exePath, args...)

	cl.logger.Debug("running command", slog.String("cmd", cmd.String()))

	out, err := cl.executor.CombinedOutput(cmd)
	if len(out) == 0 || cl.executor.ExitCode(cmd) >= 100 {
		return CmdOutput{}, NewParsedError(cl.executor.ExitCode(cmd), out)
	}

	elapsedTime := time.Since(startedAt)
	output, err := ParseOutput(string(out))
	if err != nil {
		return CmdOutput{
			Raw: string(out),
		}, err
	}
	output.ExitCode = cl.executor.ExitCode(cmd)
	output.Duration = elapsedTime

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
	return p.Message
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
	Steps    []CmdStepOutput
	Duration time.Duration
	Command  string
	Raw      string
	ExitCode int
}

func ParseOutput(s string) (CmdOutput, error) {
	lines := strings.Split(s, "\n")
	var outLines []CmdOutputLine
	var cmdOutput CmdOutput

	var step *CmdStepOutput

	for _, line := range lines {
		items := strings.Split(line, "\t")
		fmt.Println(strings.Join(items, "|"), len(items))

		if len(line) == 0 {
			continue
		}

		var il CmdOutputIdentityLine
		il.Line = line
		il.Parts = items

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
		case "Step":
			if step != nil {
				cmdOutput.Steps = append(cmdOutput.Steps, *step)
			}

			step = &CmdStepOutput{
				Name: items[1],
			}

			outLines = append(outLines, il)
		case "Output":
			if step != nil {
				step.Lines = append(step.Lines, il)
				step.OutputFilePaths = append(step.OutputFilePaths, items[1])
			}
			outLines = append(outLines, il)
		default:
			outLines = append(outLines, il)
		}
	}

	if step != nil {
		cmdOutput.Steps = append(cmdOutput.Steps, *step)
	}

	cmdOutput.Lines = outLines
	cmdOutput.Raw = s

	return cmdOutput, nil
}
