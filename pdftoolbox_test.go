package pdftoolbox_test

import (
	"testing"
	"time"

	"github.com/fikastudio/pdftoolbox-go"
	"github.com/stretchr/testify/assert"
)

func TestGeneratesCommand(t *testing.T) {
	cl := pdftoolbox.New("/tmp/pdftoolbox", nil)

	res, err := cl.RunProfile("my-profile", []string{"input.pdf"}, pdftoolbox.NewTimeoutArg(time.Second*60))
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	assert.Equal(t, `/tmp/pdftoolbox --timeout=60 input.pdf`, res.Command)
}

func TestParseOutput(t *testing.T) {
	output := `ProcessID	13913
Profile	/opt/impose/profiles/StickerIt_CLI_Example.kfpx
Input	/opt/impose/work/SA41271-1UF-R-FL5EZ9QY.pdf
Pages	1
Progress	1	%
Variable	trimHeight	70
Variable	trimWidth	70
Progress	39	%
Progress	49	%
Hit	Error	Trim box is not equal to 70 x 70 mm
Progress	100	%
Errors	1	Trim box is not equal to 70 x 70 mm
Summary	Corrections	0
Summary	Errors	1
Summary	Warnings	0
Summary	Infos	0
Finished	/opt/impose/work/SA41271-1UF-R-FL5EZ9QY.pdf
Duration	01:07
`

	parsed, err := pdftoolbox.ParseOutput(output)
	assert.NoError(t, err)
	assert.Len(t, parsed.Lines, 18)

	if !assert.Equal(t, pdftoolbox.ErrorLine, parsed.Lines[11].Type()) {
		t.FailNow()
	}

	errorLine := parsed.Lines[11].(pdftoolbox.CmdOutputErrorLine)
	assert.Equal(t, "Trim box is not equal to 70 x 70 mm", errorLine.Message)

	assert.Equal(t, time.Minute+time.Second*7, parsed.Duration)
}

func TestParseError(t *testing.T) {
	output := `ProcessID	34562
Duration	00:00
Error	1002	Could not open file /opt/impose/profilesxxx: File or folder not found`

	pe := pdftoolbox.NewParsedError(102, []byte(output))
	assert.Equal(t, 102, pe.ProcessExitCode)
	assert.Equal(t, int64(1002), pe.Code)
	assert.Equal(t, "Could not open file /opt/impose/profilesxxx: File or folder not found", pe.Message)
}
