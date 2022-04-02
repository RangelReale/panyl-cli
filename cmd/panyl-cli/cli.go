package main

import (
	"fmt"
	"os"

	"github.com/RangelReale/panyl"
	panylcli "github.com/RangelReale/panyl-cli"
	"github.com/RangelReale/panyl/plugins/clean"
	"github.com/RangelReale/panyl/plugins/metadata"
	"github.com/spf13/pflag"
)

func main() {
	cmd := panylcli.New(
		panylcli.WithDeclareGlobalFlags(func(flags *pflag.FlagSet) {
			flags.StringP("application", "a", "", "set application name")
			flags.IntP("start-line", "s", 0, "start line (0 = first line, 1 = second line)")
			flags.IntP("line-amount", "m", 0, "amount of lines to process (0 = all)")
		}),
		panylcli.WithPluginOptions([]panylcli.PluginOption{
			panylcli.PluginOption{
				Name:    "ansiescape",
				Enabled: true,
			},
			panylcli.PluginOption{
				Name:    "dockercompose",
				Enabled: true,
			},
		}),
		panylcli.WithProcessorlProvider(func(preset string, pluginsEnabled []string, flags *pflag.FlagSet) (*panyl.Processor, error) {
			parseflags := struct {
				Application string `flag:"application"`
				StartLine   int    `flag:"start-line"`
				LineAmount  int    `flag:"line-amount"`
			}{}

			err := panylcli.ParseFlags(flags, &parseflags)
			if err != nil {
				return nil, err
			}

			ret := panyl.NewProcessor(panyl.WithLineLimit(parseflags.StartLine, parseflags.LineAmount))
			if preset != "" {
				if preset == "all" {
					pluginsEnabled = []string{"ansiescape", "dockercompose"}
				} else {
					return nil, fmt.Errorf("Preset '%s' not supported", preset)
				}
			}

			if parseflags.Application != "" {
				ret.RegisterPlugin(&metadata.ForceApplication{Application: parseflags.Application})
			}

			for _, plugin := range pluginsEnabled {
				switch plugin {
				case "ansiescape":
					ret.RegisterPlugin(&clean.AnsiEscape{})
				case "dockercompose":
					ret.RegisterPlugin(&metadata.DockerCompose{})
				}
			}

			return ret, nil
		}),
		panylcli.WithResultProvider(func(flags *pflag.FlagSet) (panyl.ProcessResult, error) {
			return panylcli.NewOutput(), nil
		}),
	)

	err := cmd.Execute()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
