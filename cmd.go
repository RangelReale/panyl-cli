package panylcli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/RangelReale/panyl"
	"github.com/spf13/cobra"
)

type Cmd struct {
	cmd       *cobra.Command
	presetCmd *cobra.Command
	logCmd    *cobra.Command
}

func New(opt ...Option) *Cmd {
	var opts options
	for _, o := range opt {
		o(&opts)
	}

	ret := &Cmd{}

	ret.cmd = &cobra.Command{}

	executeFunc := func(cmd *cobra.Command, processor *panyl.Processor, filename string) error {
		var source io.Reader
		if filename == "-" {
			source = os.Stdin
		} else {
			file, err := os.Open(filename)
			if err != nil {
				return err
			}
			defer file.Close()
			source = file
		}

		result, err := opts.resultProvider(cmd.Flags())
		if err != nil {
			return err
		}

		return processor.Process(source, result)
	}

	ret.presetCmd = &cobra.Command{
		Use:     "preset",
		Short:   "run using preset plugins",
		Aliases: []string{"p"},
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.panylProvider == nil {
				return errors.New("Panyl provider was not set")
			}
			processor, err := opts.panylProvider(args[0], nil, cmd.Flags())
			if err != nil {
				return err
			}
			return executeFunc(cmd, processor, args[1])
		},
	}

	ret.logCmd = &cobra.Command{
		Use:     "log",
		Short:   "run using configurable plugin",
		Aliases: []string{"l"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.panylProvider == nil {
				return errors.New("Panyl provider was not set")
			}
			var pluginsEnabled []string
			for _, po := range opts.pluginOptions {
				enabled, err := cmd.Flags().GetBool(fmt.Sprintf("enable-%s", po.Name))
				if err != nil {
					return err
				}
				if enabled {
					pluginsEnabled = append(pluginsEnabled, po.Name)
				}
			}

			processor, err := opts.panylProvider("", pluginsEnabled, cmd.Flags())
			if err != nil {
				return err
			}
			return executeFunc(cmd, processor, args[0])
		},
	}

	if opts.globalFlags != nil {
		opts.globalFlags(ret.cmd.PersistentFlags())
	}
	if opts.presetFlags != nil {
		opts.presetFlags(ret.presetCmd.Flags())
	}
	if opts.logFlags != nil {
		opts.logFlags(ret.logCmd.Flags())
	}

	for _, pluginOption := range opts.pluginOptions {
		ret.logCmd.Flags().Bool(fmt.Sprintf("enable-%s", pluginOption.Name), pluginOption.Enabled,
			fmt.Sprintf("Enable '%s' plugin", pluginOption.Name))
	}

	ret.cmd.AddCommand(ret.presetCmd, ret.logCmd)

	return ret
}

func (c *Cmd) Execute() error {
	return c.cmd.Execute()
}
