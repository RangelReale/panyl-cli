package panylcli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/RangelReale/panyl"
)

func ExecProcessFinished(job *panyl.Job) error {
	return job.ProcessLine(&panyl.Process{
		LineNo:    0,
		LineCount: 0,
		Metadata: map[string]interface{}{
			panyl.Metadata_Timestamp: time.Now(),
			panyl.Metadata_Level:     panyl.MetadataLevel_DEBUG,
			panyl.Metadata_Message:   "process finished",
		},
		Line: "process finished",
	})
}

type execReader struct {
	name     string
	arg      []string
	isKill   atomic.Bool
	execCmd  *exec.Cmd
	source   io.Reader
	finished chan struct{}
}

func newExecReader(name string, arg ...string) (*execReader, error) {
	ret := &execReader{
		name:     name,
		arg:      arg,
		finished: make(chan struct{}),
	}
	err := ret.initReader()
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func (e *execReader) kill(s os.Signal) {
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
	if e.source == nil {
		return 0, errors.New("execReader: source is nil")
	}
	n, err = e.source.Read(p)
	if errors.Is(err, io.EOF) {
		if e.isKill.Load() {
			return n, err
		}
		fmt.Println("exec process disconnecting, running again...")
		err = e.initReader()
		if err != nil {
			return 0, err
		}
		return 0, nil
	}
	return n, err
}

func (e *execReader) initReader() error {
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

func (e *execReader) Wait() error {
	for {
		if e.execCmd != nil {
			err := e.execCmd.Wait()
			if e.isKill.Load() {
				return err
			} else if err != nil {
				fmt.Printf("error executing command: %v\n", err)
			}
			time.Sleep(5 * time.Second)
		} else {
			return fmt.Errorf("no exec process running")
		}
	}
}