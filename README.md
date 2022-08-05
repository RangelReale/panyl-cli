# panyl-cli

Panyl-cli is a library to create command-line applications using the [Panyl](https://github.com/RangelReale/panyl) library.

It allows creating a customized command for the log types you use in your services, which is the
recommended way of using Panyl.

A sample cli is provided that can be used as a starting point or proof of concept.

# Usage (sample cli)

```shell
go install github.com/RangelReale/panyl-cli/cmd/panyl-cli@latest
```

```shell
panyl-cli log [parameters] { <filename> | - | -- <shell command> }

panyl-cli preset <preset-name> [parameters] { <filename> | - | -- <shell command> }
```

Using "-" for the filename uses stdin.

Using "--" for the filename executes the command after "--" and pipes its stdout/stderr.

# Creating your own cli

```shell
go get github.com/RangelReale/panyl-cli
```

Each plugin option becomes a command line option in the format `--enable-<pluginname>=<true|false>`.

```go
package main

import (
    "fmt"
    "os"

    "github.com/RangelReale/panyl"
    panylcli "github.com/RangelReale/panyl-cli"
    "github.com/RangelReale/panyl/plugins/clean"
    "github.com/RangelReale/panyl/plugins/consolidate"
    "github.com/RangelReale/panyl/plugins/metadata"
    "github.com/RangelReale/panyl/plugins/structure"
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
		panylcli.WithProcessorProvider(func(preset string, pluginsEnabled []string, flags *pflag.FlagSet) (*panyl.Processor, []panyl.JobOption, error) {
			parseflags := struct {
				Application string `flag:"application"`
				StartLine   int    `flag:"start-line"`
				LineAmount  int    `flag:"line-amount"`
			}{}

			err := panylcli.ParseFlags(flags, &parseflags)
			if err != nil {
				return nil, nil, err
			}

			ret := panyl.NewProcessor()
			if preset != "" {
				if preset == "default" {
					pluginsEnabled = append(pluginsEnabled, "json")
				} else {
					return nil, nil, fmt.Errorf("unknown preset '%s'", preset)
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

			return ret, []panyl.JobOption{panyl.WithLineLimit(parseflags.StartLine, parseflags.LineAmount)}, nil
		}),
        panylcli.WithResultProvider(func(flags *pflag.FlagSet) (panyl.ProcessResult, error) {
            return panylcli.NewOutput(), nil
        }),
    )

	exitCode, err := cmd.Execute()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err.Error())
	}
	os.Exit(exitCode)
}
```

## Author

Rangel Reale (rangelreale@gmail.com)
