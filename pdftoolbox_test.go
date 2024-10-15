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
