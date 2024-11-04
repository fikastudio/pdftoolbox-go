package pdftoolbox

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGeneratesCommand(t *testing.T) {
	cl, err := New("/tmp/pdftoolbox", nil)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	res := cl.buildProfileCommand("my-profile.kpfx", []string{"input.pdf"}, NewTimeoutArg(time.Second*60))

	assert.Equal(t, `--timeout=60 my-profile.kpfx input.pdf`, strings.Join(res, " "))

	res = cl.buildProfileCommand(
		"../profiles/CLI_Example.kfpx",
		[]string{"SA41271-1UF-R-FL5EZ9QY.pdf"},
		NewSetVariableArg("trimWidth", "55"),
		NewSetVariableArg("trimHeight", "55"),
		NewTimeoutArg(time.Second*60),
	)

	assert.Equal(t, `--setvariable=trimWidth:55 --setvariable=trimHeight:55 --timeout=60 ../profiles/CLI_Example.kfpx SA41271-1UF-R-FL5EZ9QY.pdf`, strings.Join(res, " "))
}

func TestGeneratesCommandProfileFolder(t *testing.T) {
	pf := "/tmp/profiles"
	cl, err := New("/tmp/pdftoolbox", &ClientOpts{
		ProfileFolder: &pf,
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	res := cl.buildProfileCommand("my-profile.kpfx", []string{"input.pdf"}, NewTimeoutArg(time.Second*60))

	assert.Equal(t, `--timeout=60 /tmp/profiles/my-profile.kpfx input.pdf`, strings.Join(res, " "))

	res = cl.buildProfileCommand(
		"../profiles/CLI_Example.kfpx",
		[]string{"SA41271-1UF-R-FL5EZ9QY.pdf"},
		NewSetVariableArg("trimWidth", "55"),
		NewSetVariableArg("trimHeight", "55"),
		NewTimeoutArg(time.Second*60),
	)

	assert.Equal(t, `--setvariable=trimWidth:55 --setvariable=trimHeight:55 --timeout=60 ../profiles/CLI_Example.kfpx SA41271-1UF-R-FL5EZ9QY.pdf`, strings.Join(res, " "))
}
