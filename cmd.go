package panylcli

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

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
	ret.cmd.PersistentFlags().Bool("restart", true, "restart on close")

	executeFunc := func(cmd *cobra.Command, preset string, isExec bool, args []string) error {
		if opts.processorProvider == nil {
			return errors.New("provider was not set")
		}

		restartOnClose, err := cmd.Flags().GetBool("restart")
		if err != nil {
			return err
		}

		ctx := SLogCLIToContext(cmd.Context(), slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})))

		// check enabled plugins
		var pluginsEnabled []string
		for _, po := range opts.pluginOptions {
			if preset == "" || po.Preset {
				enabled, err := cmd.Flags().GetBool(fmt.Sprintf("enable-%s", po.Name))
				if err != nil {
					return err
				}
				if enabled {
					pluginsEnabled = append(pluginsEnabled, po.Name)
				}
			}
		}

		// create panyl processor
		ctx, processor, jobOptions, err := opts.processorProvider(ctx, preset, pluginsEnabled, cmd.Flags())
		if err != nil {
			return err
		}

		var source io.Reader
		// var execCmd *exec.Cmd
		var execHandler *execReader
		if isExec {
			// run the passed command
			execHandler, err = newExecReader(ctx, restartOnClose, args[0], args[1:]...)
			if err != nil {
				return err
			}
			source = execHandler

			c := make(chan os.Signal)
			signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
			go func() {
				s := <-c
				SLogCLIFromContext(ctx).Warn("received signal", "signal", s.String())
				execHandler.kill(s)
			}()
		} else {
			// open source file or stdin
			if args[0] == "-" {
				source = os.Stdin
			} else {
				file, err := os.Open(args[0])
				if err != nil {
					return err
				}
				defer file.Close()
				source = file
			}
		}

		// create the output provider
		output, err := opts.outputProvider(ctx, cmd.Flags())
		if err != nil {
			return err
		}

		// process
		err = processor.Process(ctx, source, output, jobOptions...)
		if err != nil {
			SLogCLIFromContext(ctx).Error("error running processor", "error", err)
		} else if execHandler != nil {
			execHandler.Wait()
		}

		return nil
	}

	ret.presetCmd = &cobra.Command{
		Use:     "preset <preset-name> [flags]",
		Short:   "run using preset plugins",
		Aliases: []string{"p"},
		Args:    cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().ArgsLenAtDash() == 0 {
				return errors.New("missing preset name")
			}
			if cmd.Flags().ArgsLenAtDash() > 1 {
				return errors.New("command to execute must be the last parameter")
			}
			return executeFunc(cmd, args[0], cmd.Flags().ArgsLenAtDash() != -1, args[1:])
		},
	}

	ret.logCmd = &cobra.Command{
		Use:     "log",
		Short:   "run using configurable plugin",
		Aliases: []string{"l"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().ArgsLenAtDash() > 0 {
				return errors.New("command to execute must be the last parameter")
			}
			return executeFunc(cmd, "", cmd.Flags().ArgsLenAtDash() != -1, args)
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
		if pluginOption.Preset {
			ret.presetCmd.Flags().Bool(fmt.Sprintf("enable-%s", pluginOption.Name), pluginOption.PresetEnabled,
				fmt.Sprintf("Enable '%s' plugin", pluginOption.Name))
		}
	}

	ret.cmd.AddCommand(ret.presetCmd, ret.logCmd)

	return ret
}

// Execute executes the command and returns the exit code and error if available
func (c *Cmd) Execute() (int, error) {
	err := c.cmd.Execute()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			waitStatus := exitError.Sys().(syscall.WaitStatus)
			return waitStatus.ExitStatus(), err
		}
		return 1, err
	}
	return 0, nil
}
