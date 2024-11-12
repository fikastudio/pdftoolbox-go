package pdftoolbox_test

import (
	"os/exec"
	"testing"
	"time"

	"github.com/fikastudio/pdftoolbox-go"
	"github.com/stretchr/testify/assert"
)

func TestParseErrorOutput(t *testing.T) {
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

func TestParseOutput(t *testing.T) {
	output := `ProcessID	35218
Profile	/opt/impose/profiles/StickerIt_CLI_Example.kfpx
Input	/opt/impose/work/SA41271-1UF-R-FL5EZ9QY.pdf
Pages	1
Progress	1	%
Variable	trimHeight	55
Variable	trimWidth	55
Progress	39	%
Progress	49	%
Progress	100	%
Summary	Corrections	0
Summary	Errors	0
Summary	Warnings	0
Summary	Infos	0
Finished	/opt/impose/work/SA41271-1UF-R-FL5EZ9QY.pdf
Duration	00:01`

	parsed, err := pdftoolbox.ParseOutput(output)
	assert.NoError(t, err)
	assert.Len(t, parsed.Lines, 16)
	assert.Equal(t, output, parsed.Raw)
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

func TestFullOutput(t *testing.T) {
	output := `ProcessID	104089
Profile	/opt/impose/profiles/Indigo-MotionCutter.kfpx
Input	/opt/impose/work/SA41271-1UF-R-FL5EZ9QY.pdf
Pages	1
Progress	1	%
Step	Variable	Indigo-MotionCutter_Job
Step	Action	Indigo-MotionCutter_2
Step	Fixup	Indigo-MotionCutter_MediaBox
Variable	imposeAddBotomZund	11
Variable	imposeAddLeft	15
Variable	imposeAddRight	15
Variable	imposeAddTop	50
Progress	12	%
Progress	40	%
Progress	45	%
Fix	Indigo-MotionCutter_MediaBox
Progress	79	%
Progress	80	%
Progress	81	%
Progress	82	%
Progress	91	%
Progress	100	%
Summary	Corrections	1
Summary	Errors	0
Summary	Warnings	0
Summary	Infos	0
Step	Fixup	Indigo-MotionCutter_DotMark
Variable	DotMarkX	17.5
Variable	DotMarkY	36.5
Variable	imposeAddRight	15
Variable	imposeTopZund	29
Progress	12	%
Progress	40	%
Progress	45	%
Fix	Indigo-MotionCutter_DotMark
Progress	79	%
Progress	80	%
Progress	91	%
Progress	100	%
Summary	Corrections	1
Summary	Errors	0
Summary	Warnings	0
Summary	Infos	0
Step	Fixup	Indigo-MotionCutter_QRCodeBox
Variable	QRCodeBoxX	75
Variable	QRCodeBoxY	40
Variable	imposeAddLeft	15
Variable	imposeTopZund	29
Progress	12	%
Progress	40	%
Progress	45	%
Fix	Indigo-MotionCutter_QRCodeBox
Progress	79	%
Progress	80	%
Progress	91	%
Progress	100	%
Summary	Corrections	1
Summary	Errors	0
Summary	Warnings	0
Summary	Infos	0
Step	Fixup	Indigo-MotionCutter_QRCode
Variable	QRCodeData	Data for QR code...
Variable	QRCodeX	21.75
Variable	QRCodeY	33
Variable	imposeAddLeft	15
Variable	imposeTopZund	29
Variable	qrData	Data for QR code...
Progress	12	%
Progress	40	%
Progress	45	%
Progress	51	%
Fix	Indigo-MotionCutter_QRCode
Progress	79	%
Progress	80	%
Progress	91	%
Progress	100	%
Summary	Corrections	1
Summary	Errors	0
Summary	Warnings	0
Summary	Infos	0
Step	Fixup	Indigo-MotionCutter_QRCodeJobInfo
Variable	QRCodeData	Data for QR code...
Variable	QRCodeJobInfoX	36.75
Variable	QRCodeJobInfoY	41
Variable	imposeAddLeft	15
Variable	imposeTopZund	29
Variable	qrData	Data for QR code...
Progress	12	%
Progress	40	%
Progress	45	%
Progress	51	%
Fix	Indigo-MotionCutter_QRCodeJobInfo
Progress	79	%
Progress	80	%
Progress	91	%
Progress	100	%
Summary	Corrections	1
Summary	Errors	0
Summary	Warnings	0
Summary	Infos	0
Step	Fixup	Extract cutline
Variable	cutlineName	Die Cut
Progress	12	%
Progress	40	%
Progress	45	%
Fix	Extract cutline
Progress	79	%
Progress	80	%
Progress	91	%
Progress	100	%
Summary	Corrections	64
Summary	Errors	0
Summary	Warnings	0
Summary	Infos	0
Step	Create PDF copy
Output	/opt/impose/work/output/Output_File.pdf_cut_x2_0001.pdf
Step	File Pickup
Step	Fixup	Indigo-MotionCutter_ZundSheetingLines
Variable	imposeAddBotomZund	11
Variable	imposeAddLeft	15
Variable	imposeAddRight	15
Variable	imposeTopZund	29
Variable	maxWidth	340
Progress	12	%
Progress	40	%
Progress	45	%
Fix	Indigo-MotionCutter_ZundSheetingLines
Progress	79	%
Progress	80	%
Progress	91	%
Progress	100	%
Summary	Corrections	1
Summary	Errors	0
Summary	Warnings	0
Summary	Infos	0
Step	Fixup	Indigo-MotionCutter_ZundSheetingCamera
Variable	ZundCamera_LeftX	22.5
Variable	imposeAddLeft	15
Progress	12	%
Progress	40	%
Progress	45	%
Fix	Indigo-MotionCutter_ZundSheetingCamera
Progress	79	%
Progress	80	%
Progress	91	%
Progress	100	%
Summary	Corrections	1
Summary	Errors	0
Summary	Warnings	0
Summary	Infos	0
Step	Fixup	Extract Zund layer
Progress	12	%
Progress	40	%
Progress	45	%
Fix	Extract Zund layer
Progress	79	%
Progress	91	%
Progress	100	%
Summary	Corrections	176
Summary	Errors	0
Summary	Warnings	0
Summary	Infos	0
Step	Create PDF copy
Output	/opt/impose/work/output/Output_File.pdf_sheeting_x2_0001.pdf
Output	/opt/impose/work/output/Output_File.pdf_sheeting_x2_0002.pdf
Step	File Pickup
Step	Rename PDF
Output	/opt/impose/work/output/Output_File.pdf_x2_0001.pdf
Finished	/opt/impose/work/SA41271-1UF-R-FL5EZ9QY.pdf
Duration	00:03`

	exe := &FakeExecutor{
		cmd: &exec.Cmd{
			Path: "/tmp/fakepdftoolbox",
			Args: []string{"myarg"},
		},
		output:   output,
		exitCode: 1002,
	}

	cli, err := pdftoolbox.New("/tmp/fakepdftoolbox", &pdftoolbox.ClientOpts{
		Executor: exe,
	})
	assert.NoError(t, err)
	res, err := cli.RunProfile("myprofile", []string{"inputfile.pdf"})

	assert.NoError(t, err)
	assert.Len(t, res.Lines, 170)
	assert.Equal(t, output, res.Raw)
	assert.Len(t, res.Steps, 16)

	pdfCopyStep := res.Steps[13]
	assert.Equal(t, pdfCopyStep.Name, "Create PDF copy")
	assert.Len(t, pdfCopyStep.Lines, 2)
	assert.Len(t, pdfCopyStep.OutputFilePaths, 2)
	assert.Equal(t, "/opt/impose/work/output/Output_File.pdf_sheeting_x2_0001.pdf", pdfCopyStep.OutputFilePaths[0])
	assert.Equal(t, "/opt/impose/work/output/Output_File.pdf_sheeting_x2_0002.pdf", pdfCopyStep.OutputFilePaths[1])

	lastStep := res.Steps[15]
	assert.Equal(t, lastStep.Name, "Rename PDF")
	assert.Len(t, lastStep.Lines, 1)
	assert.Len(t, lastStep.OutputFilePaths, 1)
	assert.Equal(t, "/opt/impose/work/output/Output_File.pdf_x2_0001.pdf", lastStep.OutputFilePaths[0])
}

type FakeExecutor struct {
	cmd      *exec.Cmd
	output   string
	exitCode int
}

func NewFakeExecutor() (*FakeExecutor, error) {

	return &FakeExecutor{}, nil
}

func (e FakeExecutor) Command(name string, args ...string) *exec.Cmd {
	return e.cmd

}

func (e FakeExecutor) CombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	return []byte(e.output), nil
}

func (e FakeExecutor) ExitCode(cmd *exec.Cmd) int {
	return e.exitCode
}

func TestWithExec(t *testing.T) {
	output := `ProcessID	34562
Duration	00:00
Error	1002	Could not open file /opt/impose/profilesxxx: File or folder not found`

	exe := &FakeExecutor{
		cmd: &exec.Cmd{
			Path: "/tmp/fakepdftoolbox",
			Args: []string{"myarg"},
		},
		output:   output,
		exitCode: 1002,
	}

	cli, err := pdftoolbox.New("/tmp/fakepdftoolbox", &pdftoolbox.ClientOpts{
		Executor: exe,
	})
	assert.NoError(t, err)
	res, err := cli.RunProfile("myprofile", []string{"inputfile.pdf"})
	assert.NoError(t, err)
	assert.Equal(t, 1002, res.ExitCode)
}
