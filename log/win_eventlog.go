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

package log

import (
	"fmt"
	"os"

	"golang.org/x/sys/windows/svc/eventlog"
)

type eventlogger struct {
	log         *eventlog.Log
	debugAsInfo bool
}

func NewEventLogger(name string, debugAsInfo bool) (*eventlogger, error) {
	logHandle, err := eventlog.Open(name)
	if err != nil {
		return nil, err
	}
	return &eventlogger{log: logHandle, debugAsInfo: debugAsInfo}, nil
}

func (s *eventlogger) Info(msg string) error {
	err := s.log.Info(100, msg)

	if err != nil {
		fmt.Fprintf(os.Stderr, "eventlogger: can't send log to eventlog: %v\n", err)
	}

	return err
}

func (s *eventlogger) Warning(msg string) error {
	err := s.log.Warning(101, msg)

	if err != nil {
		fmt.Fprintf(os.Stderr, "eventlogger: can't send log to eventlog: %v\n", err)
	}

	return err
}

func (s *eventlogger) Error(msg string) error {
	err := s.log.Error(102, msg)

	if err != nil {
		fmt.Fprintf(os.Stderr, "eventlogger: can't send log to eventlog: %v\n", err)
	}

	return err
}
