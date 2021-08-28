package util

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/cosmos/cosmos-sdk/client"
)

func LineBreak() *cobra.Command {
	return &cobra.Command{Run: func(*cobra.Command, []string) {}}
}

func PrintConsoleOutput(ctx client.Context, toPrint interface{}) error {
	var marshal func(interface{}) ([]byte, error)

	if ctx.OutputFormat == "text" {
		marshal = yaml.Marshal
	} else {
		marshal = json.Marshal
	}

	bz, err := marshal(toPrint)
	if err != nil {
		return err
	}

	writer := ctx.Output
	if writer == nil {
		writer = os.Stdout
	}

	_, err = writer.Write(bz)
	if err != nil {
		return err
	}

	if ctx.OutputFormat != "text" {
		// append new-line for formats besides YAML
		_, err = writer.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}

	return nil
}
