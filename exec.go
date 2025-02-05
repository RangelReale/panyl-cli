package panylcli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/RangelReale/panyl/v2"
)

func ExecProcessFinished(ctx context.Context, job *panyl.Job) error {
	SLogCLIFromContext(ctx).Info("process finished")
	return nil
}

type execReader struct {
	ctx            context.Context
	name           string
	arg            []string
	restartOnClose bool
	isKill         atomic.Bool
	execCmd        *exec.Cmd
	source         io.Reader
	outputCache    []byte
}

func newExecReader(ctx context.Context, restartOnClose bool, name string, arg ...string) (*execReader, error) {
	ret := &execReader{
		ctx:            ctx,
		name:           name,
		arg:            arg,
		restartOnClose: restartOnClose,
	}
	err := ret.initReader()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (e *execReader) kill(s os.Signal) {
	SLogCLIFromContext(e.ctx).Warn("killing process")
	e.isKill.Store(true)
	if e.execCmd != nil {
		if runtime.GOOS != "windows" {
			e.execCmd.Process.Signal(s)
		} else {
			e.execCmd.Process.Kill()
		}
	}
}

func (e *execReader) Read(p []byte) (n int, err error) {
loop:
	for {
		// the first time it was already initialized
		for {
			n, err := e.source.Read(p)
			if errors.Is(err, io.EOF) {
				break
			}
			if len(e.outputCache) < 200 {
				e.outputCache = append(e.outputCache, p[:min(n, 200-len(e.outputCache))]...)
			}
			select {
			case <-e.ctx.Done():
				break loop
			default:
				return n, err
			}
		}

		if e.isKill.Load() {
			break
		}

		SLogCLIFromContext(e.ctx).Warn("exec process exited, waiting for exit code...")

		// check for exit errors
		err := e.execCmd.Wait()
		if err != nil {
			var ee *exec.ExitError
			if errors.As(err, &ee) {
				if ee.ExitCode() > 0 {
					var outputCache string
					if len(e.outputCache) < 199 {
						outputCache = string(e.outputCache)
					}
					return 0, fmt.Errorf("error executing command: %s (exit code: %d)(stderr: '%s')",
						ee.Error(), ee.ExitCode(), outputCache)
				}
				if e.restartOnClose {
					SLogCLIFromContext(e.ctx).Warn("exec process exited, running again...",
						"error", err.Error(),
						"exitCode", ee.ExitCode())
				}
			} else {
				SLogCLIFromContext(e.ctx).Error("error executing command", "error", err.Error())
			}
		} else {
			if e.restartOnClose {
				SLogCLIFromContext(e.ctx).Warn("exec process exited, running again...")
			}
		}
		if !e.restartOnClose {
			break
		}

		select {
		case <-e.ctx.Done():
			break loop
		case <-time.After(5 * time.Second):
		}

		if e.isKill.Load() {
			break
		}

		// initialize a new reader
		err = e.initReader()
		if err != nil {
			return 0, err
		}
	}
	return 0, io.EOF
}

func (e *execReader) initReader() error {
	e.outputCache = nil
	var err error
	// run the passed command
	e.execCmd = exec.Command(e.name, e.arg...)
	e.source, err = e.execCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("error creating stdout pipe: %v", err)
	}
	e.execCmd.Stderr = e.execCmd.Stdout

	err = e.execCmd.Start()
	if err != nil {
		return fmt.Errorf("error starting command: %v", err)
	}
	return nil
}

func (e *execReader) Wait() {}
