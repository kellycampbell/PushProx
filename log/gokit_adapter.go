// Copyright 2020 The Prometheus Authors
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

package log

import (
	"github.com/go-kit/kit/log/level"
)

// Returns an adapter implementing the go-kit/kit/log.Logger interface on our
// logrus logger
func NewToolkitAdapter() *logAdapter {
	return &logAdapter{}
}

type logAdapter struct{}

func (*logAdapter) Log(keyvals ...interface{}) error {
	var lvl level.Value
	var msg string
	for i := 0; i < len(keyvals); i += 2 {
		switch keyvals[i] {
		case "level":
			tlvl, ok := keyvals[i+1].(level.Value)
			if !ok {
				Warnf("Could not cast level of type %T", keyvals[i+1])
			} else {
				lvl = tlvl
			}
		case "msg":
			msg = keyvals[i+1].(string)
		}
	}

	switch lvl {
	case level.ErrorValue():
		Errorln(msg)
	case level.WarnValue():
		Warnln(msg)
	case level.InfoValue():
		Infoln(msg)
	case level.DebugValue():
		Debugln(msg)
	default:
		Warnf("Unmatched log level: '%v' for message %q", lvl, msg)
	}

	return nil
}
