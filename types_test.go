package pdftoolbox_test

import (
	"encoding/json"
	"testing"

	"github.com/fikastudio/pdftoolbox-go"
	"github.com/stretchr/testify/assert"
)

func TestDeserialize(t *testing.T) {
	t.Skip()
	stepOutput := pdftoolbox.CmdStepOutput{
		Name: "x",
		Lines: []pdftoolbox.CmdOutputLine{
			pdftoolbox.CmdOutputIdentityLine{Line: "my line", Typename: "identity"},
		},
	}

	expected := `{"name":"x","lines":[{"__typename":"identity","line":"my line","parts":null}],"outputFilePaths":null}`

	b, err := json.Marshal(stepOutput)
	assert.NoError(t, err)
	assert.Equal(t, expected, string(b))

	var so pdftoolbox.CmdStepOutput
	err = json.Unmarshal(b, &so)
	assert.NoError(t, err)

}
