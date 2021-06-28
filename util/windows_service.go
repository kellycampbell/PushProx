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

// +build windows

package util

// Provides necessary functionality to run wireguard_exporter as a
// Windows service.

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"golang.org/x/sys/windows/svc"
)

type windowsService struct {
	stopCh             chan<- bool
	finishedCh         <-chan bool
	shutdownCompleteCh chan<- bool
	server             *http.Server
	logger             log.Logger
}

func InitService(serviceName string, server *http.Server, logger log.Logger, stopCh chan<- bool, finishedCh <-chan bool) chan bool {
	isService, err := svc.IsWindowsService()
	if err != nil {
		level.Error(logger).Log("msg", "Failed to determine if we are running as Windows Service", "err", err)
	}
	shutdownCompleteCh := make(chan bool)
	if isService {
		go func() {
			level.Info(logger).Log("msg", "Running as a Windows Service")
			err = svc.Run(serviceName, &windowsService{stopCh: stopCh, server: server, logger: logger, shutdownCompleteCh: shutdownCompleteCh})
			if err != nil {
				level.Info(logger).Log("msg", "Failed to start Service", "err", err)
			}
		}()
	}
	return shutdownCompleteCh
}

func (s *windowsService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				level.Info(s.logger).Log("msg", "Stop or Shutdown signal received")
				s.stopCh <- true
				ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
				defer func() {
					cancel()
				}()
				changes <- svc.Status{State: svc.StopPending}
				if err := s.server.Shutdown(ctx); err != nil {
					level.Error(s.logger).Log("msg", "server shutdown error", "err", err)
				}
				level.Info(s.logger).Log("msg", "Exiting Windows Service loop")
				break loop
			default:
				level.Info(s.logger).Log("msg", fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	close(s.shutdownCompleteCh)
	return
}
