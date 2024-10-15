package pdftoolbox

import (
	"fmt"
	"math"
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
