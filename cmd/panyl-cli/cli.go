package main

import (
	"context"
	"fmt"
	"os"

	panylcli "github.com/RangelReale/panyl-cli/v2"
	"github.com/RangelReale/panyl/v2"
	"github.com/RangelReale/panyl/v2/plugins/clean"
	"github.com/RangelReale/panyl/v2/plugins/consolidate"
	"github.com/RangelReale/panyl/v2/plugins/metadata"
	"github.com/RangelReale/panyl/v2/plugins/structure"
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
			{
				Name:          "ansiescape",
				Enabled:       true,
				Preset:        true,
				PresetEnabled: true,
			},
			{
				Name:    "json",
				Enabled: true,
			},
			{
				Name:    "consolidate-lines",
				Enabled: false,
			},
		}),
		panylcli.WithProcessorProvider(func(ctx context.Context, preset string, pluginsEnabled []string,
			flags *pflag.FlagSet) (context.Context, *panyl.Processor, []panyl.JobOption, error) {
			parseflags := struct {
				Application string `flag:"application"`
				StartLine   int    `flag:"start-line"`
				LineAmount  int    `flag:"line-amount"`
			}{}

			err := panylcli.ParseFlags(flags, &parseflags)
			if err != nil {
				return ctx, nil, nil, err
			}

			ret := panyl.NewProcessor()
			if preset != "" {
				if preset == "default" {
					pluginsEnabled = append(pluginsEnabled, "json")
				} else {
					return ctx, nil, nil, fmt.Errorf("unknown preset '%s'", preset)
				}
			}

			if parseflags.Application != "" {
				ret.RegisterPlugin(&metadata.ForceApplication{Application: parseflags.Application})
			}

			for _, plugin := range panylcli.PluginsEnabledUnique(pluginsEnabled) {
				switch plugin {
				case "ansiescape":
					ret.RegisterPlugin(&clean.AnsiEscape{})
				case "json":
					ret.RegisterPlugin(&structure.JSON{})
				case "consolidate-lines":
					ret.RegisterPlugin(&consolidate.JoinAllLines{})
				}
			}

			return ctx, ret, []panyl.JobOption{panyl.WithLineLimit(parseflags.StartLine, parseflags.LineAmount)}, nil
		}),
		panylcli.WithOutputProvider(func(ctx context.Context, flags *pflag.FlagSet) (panyl.Output, error) {
			return panylcli.NewDefaultOutput(), nil
		}),
	)

	exitCode, err := cmd.Execute()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
	}
	os.Exit(exitCode)
}
