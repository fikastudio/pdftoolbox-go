package main

import (
	"encoding/json"
	"fmt"

	"github.com/fikastudio/pdftoolbox-go"
)

func main() {
	cli, err := pdftoolbox.New("../callas/callas_pdfToolboxCLI15_x64_Linux_15-1-639/pdfToolbox", &pdftoolbox.ClientOpts{})
	if err != nil {
		panic(err)
	}

	resp, err := cli.EnumerateProfiles("../profiles")
	if err != nil {
		panic(err)
	}

	b, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(b))

	runResp, err := cli.RunProfile(
		"../profiles/CLI_Example.kfpx",
		[]string{"SA41271-1UF-R-FL5EZ9QY.pdf"},
		pdftoolbox.NewSetVariableArg("trimWidth", "55"),
		pdftoolbox.NewSetVariableArg("trimHeight", "55"),
	)
	if err != nil {
		panic(err)
	}

	b, _ = json.MarshalIndent(runResp, "", "  ")
	fmt.Println(string(b))
}
