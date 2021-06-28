// Copyright 2015 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Implementation forked from github.com/prometheus/common
//
// +build windows

// windows specific
package log

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/windows"
)

var (
	kernel32         = syscall.MustLoadDLL("kernel32.dll")
	procSetStdHandle = kernel32.MustFindProc("SetStdHandle")
)

// dupFD is used to initialize OrigStderr (see stderr_redirect.go).
func dupFD(fd uintptr) (uintptr, error) {
	// Adapted from https://github.com/golang/go/blob/go1.8/src/syscall/exec_windows.go#L303.
	p, err := windows.GetCurrentProcess()
	if err != nil {
		return 0, err
	}
	var h windows.Handle
	return uintptr(h), windows.DuplicateHandle(p, windows.Handle(fd), p, &h, 0, true, windows.DUPLICATE_SAME_ACCESS)
}

// redirectStderr is used to redirect internal writes to the error
// handle to the specified file. This is needed to ensure that
// harcoded writes to the error handle by e.g. the Go runtime are
// redirected to a log file of our choosing.
//
// We also override os.Stderr for those other parts of Go which use
// that and not fd 2 directly.
func redirectStderr(f *os.File) error {
	if err := windows.SetStdHandle(windows.STD_ERROR_HANDLE, windows.Handle(f.Fd())); err != nil {
		return err
	}
	os.Stderr = f
	return nil
}

// end windows specific

func InitEventLog(serviceName string) {
	evt_log, err := NewEventLogger(serviceName, false)
	if err != nil {
		fmt.Printf("Error creating windows event logger: %+v", err)
		return
	}
	logfile, err := os.CreateTemp("", "pushprox-log-*.txt")
	if err != nil {
		s := fmt.Sprintf("Error creating log file: %+v", err)
		fmt.Println(s)
		evt_log.Error(s)
	} else {
		s := fmt.Sprintf("Logging to temp file %s", logfile.Name())
		fmt.Println(s)
		evt_log.Info(s)

		err = redirectStderr(logfile)
		if err != nil {
			s := fmt.Sprintf("Error redirecting stderr for logging: %+v", err)
			fmt.Println(s)
			evt_log.Warning(s)
		}
	}
}
