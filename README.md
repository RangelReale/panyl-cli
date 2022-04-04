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
panyl-cli log [parameters] { <filename> | - }

panyl-cli preset <preset-name> [parameters] { <filename> | - }
```

Using "-" for the filename uses stdin.

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
        panylcli.WithProcessorProvider(func(preset string, pluginsEnabled []string, flags *pflag.FlagSet) (*panyl.Processor, error) {
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
                if preset == "default" {
                    pluginsEnabled = append(pluginsEnabled, "json")
                } else {
                    return nil, fmt.Errorf("unknown preset '%s'", preset)
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

            return ret, nil
        }),
        panylcli.WithResultProvider(func(flags *pflag.FlagSet) (panyl.ProcessResult, error) {
            return panylcli.NewOutput(), nil
        }),
    )

    err := cmd.Execute()
    if err != nil {
        _, _ = fmt.Fprintln(os.Stderr, err.Error())
        os.Exit(1)
    }
}
```

# Sample Ansi color output

```go
import (
    "bytes"
    "encoding/json"
    "fmt"
    "github.com/RangelReale/panyl"
    "github.com/fatih/color"
    "time"
)

type OutputSprintfFunc func(format string, a ...interface{}) string

type Output struct {
    ColorInformation, ColorWarning, ColorError, ColorInternalError, ColorUnknown OutputSprintfFunc
}

func NewOutput(ansi bool) *Output {
    ret := &Output{
        ColorError:         fmt.Sprintf,
        ColorWarning:       fmt.Sprintf,
        ColorInformation:   fmt.Sprintf,
        ColorInternalError: fmt.Sprintf,
        ColorUnknown:       fmt.Sprintf,
    }
    if ansi {
        ret.ColorError = color.New(color.FgRed).SprintfFunc()
        ret.ColorWarning = color.New(color.FgYellow).SprintfFunc()
        ret.ColorInformation = color.New(color.FgGreen).SprintfFunc()
        ret.ColorInternalError = color.New(color.FgHiRed).SprintfFunc()
        ret.ColorUnknown = color.New(color.FgMagenta).SprintfFunc()
    }
    return ret
}

func (o *Output) OnResult(p *panyl.Process) (cont bool) {
    var out bytes.Buffer

    // level
    var levelColor OutputSprintfFunc
    level := p.Metadata.StringValue(panyl.Metadata_Level)
    switch level {
    case panyl.MetadataLevel_TRACE, panyl.MetadataLevel_DEBUG, panyl.MetadataLevel_INFO:
        levelColor = o.ColorInformation
    case panyl.MetadataLevel_WARNING:
        levelColor = o.ColorWarning
    case panyl.MetadataLevel_CRITICAL, panyl.MetadataLevel_FATAL:
        levelColor = o.ColorError
    default:
        level = "unknown"
        levelColor = o.ColorUnknown
    }

    // timestamp
    if ts, ok := p.Metadata[panyl.Metadata_Timestamp]; ok {
        out.WriteString(fmt.Sprintf("%s ", ts.(time.Time).Local().Format("2006-01-02 15:04:05.000")))
    }

    // application
    if application := p.Metadata.StringValue(panyl.Metadata_Application); application != "" {
        out.WriteString(fmt.Sprintf("| %s | ", application))
    }

    // level
    if level != "" {
        out.WriteString(fmt.Sprintf("[%s] ", level))
    }

    // format
    if format := p.Metadata.StringValue(panyl.Metadata_Format); format != "" {
        out.WriteString(fmt.Sprintf("(%s) ", format))
    }

    // category
    if category := p.Metadata.StringValue(panyl.Metadata_Category); category != "" {
        out.WriteString(fmt.Sprintf("{{%s}} ", category))
    }

    // message
    if msg := p.Metadata.StringValue(panyl.Metadata_Message); msg != "" {
        out.WriteString(msg)
    } else if p.Line != "" {
        out.WriteString(p.Line)
    } else if len(p.Data) > 0 {
        dt, err := json.Marshal(p.Data)
        if err != nil {
            fmt.Println(o.ColorInternalError("Error marshaling data to json: %s", err.Error()))
            return
        }
        out.WriteString(fmt.Sprintf("| %s", string(dt)))
    }

    fmt.Println(levelColor(out.String()))

    return true
}

```

## Author

Rangel Reale (rangelreale@gmail.com)
