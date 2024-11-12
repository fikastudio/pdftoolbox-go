package pdftoolbox

import "time"

type LineOutputType uint8

const (
	IdentityLine LineOutputType = iota
	ErrorLine
	DurationLine
)

type CmdOutputLine interface {
	Type() LineOutputType
	String() string
}

type CmdOutputIdentityLine struct {
	Line  string
	Parts []string
}

func (l CmdOutputIdentityLine) Type() LineOutputType {
	return IdentityLine
}

func (l CmdOutputIdentityLine) String() string {
	return l.Line
}

type CmdOutputDurationLine struct {
	Line  string
	Parts []string
	dur   time.Duration
}

func (l CmdOutputDurationLine) Type() LineOutputType {
	return DurationLine
}

func (l CmdOutputDurationLine) String() string {
	return l.Line
}

func (l CmdOutputDurationLine) Duration() time.Duration {
	return l.dur
}

type CmdOutputErrorLine struct {
	Code    int64
	Message string
}

func (l CmdOutputErrorLine) Type() LineOutputType {
	return ErrorLine
}

func (l CmdOutputErrorLine) String() string {
	return l.Message
}

type CmdStepOutput struct {
	Name            string
	Lines           []CmdOutputLine
	OutputFilePaths []string
}

type EnumerateProfilesResponse struct {
	Information Information `json:"information"`
	Profiles    []Profiles  `json:"profiles"`
}
type Information struct {
	Computername    string    `json:"computername"`
	DateTime        time.Time `json:"date_time"`
	OperatingSystem string    `json:"operating_system"`
	ProductName     string    `json:"product_name"`
	ProductVersion  string    `json:"product_version"`
	Username        string    `json:"username"`
}
type Variables struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Type  string `json:"type"`
	Value any    `json:"value"`
}

type Profiles struct {
	Comment          string         `json:"comment"`
	CreationDate     time.Time      `json:"creation_date"`
	ModificationDate time.Time      `json:"modification_date"`
	Name             string         `json:"name"`
	Path             string         `json:"path"`
	Size             string         `json:"size"`
	Variables        []Variables    `json:"variables"`
	Vars             map[string]any `json:"vars"`
}
