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

	res := cl.buildProfileCommand("my-profile", []string{"input.pdf"}, NewTimeoutArg(time.Second*60))

	assert.Equal(t, `--timeout=60 input.pdf`, strings.Join(res, " "))

	res = cl.buildProfileCommand(
		"../profiles/CLI_Example.kfpx",
		[]string{"SA41271-1UF-R-FL5EZ9QY.pdf"},
		NewSetVariableArg("trimWidth", "55"),
		NewSetVariableArg("trimHeight", "55"),
		NewTimeoutArg(time.Second*60),
	)

	assert.Equal(t, `--trimWidth=55 --trimHeight=55 --timeout=60 SA41271-1UF-R-FL5EZ9QY.pdf`, strings.Join(res, " "))
}
