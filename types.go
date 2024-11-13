package pdftoolbox

import (
	"encoding/json"
	"time"
)

type LineOutputType string

const (
	IdentityLine LineOutputType = "identity"
	ErrorLine    LineOutputType = "error"
	DurationLine LineOutputType = "duration"
)

type CmdOutputLine interface {
	Type() LineOutputType
	String() string
}

type LineOutputTypename struct {
	json.RawMessage
	Typename LineOutputType `json:"__typename"`
}

type CmdOutputIdentityLine struct {
	Typename LineOutputType `json:"__typename"`
	Line     string         `json:"line"`
	Parts    []string       `json:"parts"`
}

func (l CmdOutputIdentityLine) Type() LineOutputType {
	return IdentityLine
}

func (l CmdOutputIdentityLine) String() string {
	return l.Line
}

func (l *CmdOutputIdentityLine) MarshalJSON() (b []byte, e error) {
	l.Typename = l.Type()
	return json.Marshal(l)
}

type CmdOutputDurationLine struct {
	Typename LineOutputType `json:"__typename"`
	Line     string         `json:"line"`
	Parts    []string       `json:"parts"`
	dur      time.Duration
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

func (l *CmdOutputDurationLine) MarshalJSON() (b []byte, e error) {
	l.Typename = l.Type()
	return json.Marshal(l)
}

type CmdOutputErrorLine struct {
	Typename LineOutputType `json:"__typename"`
	Code     int64          `json:"code"`
	Message  string         `json:"message"`
}

func (l CmdOutputErrorLine) Type() LineOutputType {
	return ErrorLine
}

func (l CmdOutputErrorLine) String() string {
	return l.Message
}

func (l *CmdOutputErrorLine) MarshalJSON() (b []byte, e error) {
	l.Typename = l.Type()
	return json.Marshal(l)
}

type CmdStepOutput struct {
	Name            string          `json:"name"`
	Lines           []CmdOutputLine `json:"lines"`
	OutputFilePaths []string        `json:"outputFilePaths"`
}

func (ce *CmdStepOutput) UnmarshalJSON(b []byte) error {
	var objMap map[string]*json.RawMessage
	err := json.Unmarshal(b, &objMap)
	if err != nil {
		return err
	}

	for k, rawMessage := range objMap {
		switch k {
		case "name":
			err := json.Unmarshal(*rawMessage, &ce.Name)
			if err != nil {
				return err
			}
		case "lines":
			var lines []LineOutputTypename
			err := json.Unmarshal(*rawMessage, &lines)
			if err != nil {
				return err
			}

			var finalLines []CmdOutputLine

			for _, line := range lines {
				switch line.Typename {
				case ErrorLine:
					var el CmdOutputErrorLine
					err := json.Unmarshal(line.RawMessage, &el)
					if err != nil {
						return err
					}

					finalLines = append(finalLines, el)
				case IdentityLine:
					var el CmdOutputIdentityLine
					err := json.Unmarshal(line.RawMessage, &el)
					if err != nil {
						return err
					}

					finalLines = append(finalLines, el)
				case DurationLine:
					var el CmdOutputDurationLine
					err := json.Unmarshal(line.RawMessage, &el)
					if err != nil {
						return err
					}

					finalLines = append(finalLines, el)
				}
			}
		case "outputFilePaths":
			if rawMessage == nil {
				continue
			}
			err := json.Unmarshal(*rawMessage, &ce.OutputFilePaths)
			if err != nil {
				return err
			}
		}
	}

	return nil
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
