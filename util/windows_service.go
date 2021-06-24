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

	"github.com/prometheus-community/pushprox/log"
	"golang.org/x/sys/windows/svc"
)

type windowsService struct {
	stopCh chan<- bool
	server *http.Server
}

func InitService(serviceName string, server *http.Server, stopCh chan<- bool) {
	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("failed to determine if we are running as service: %v", err)
	}
	if isService {
		go func() {
			log.Info("Running as a service")
			err = svc.Run(serviceName, &windowsService{stopCh: stopCh, server: server})
			if err != nil {
				log.Infof("Failed to start service: %v\n", err)
			}
		}()
	}
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
				log.Info("Stop or Shutdown signal received")
				s.stopCh <- true
				ctx, cancel := context.WithTimeout(context.Background(), 1000*time.Millisecond)
				defer func() {
					cancel()
				}()
				changes <- svc.Status{State: svc.StopPending}
				if err := s.server.Shutdown(ctx); err != nil {
					log.Fatalf("server shutdown error: %+s", err)
				}
				break loop
			default:
				log.Info(fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	return
}
