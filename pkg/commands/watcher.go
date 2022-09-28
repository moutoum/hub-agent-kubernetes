/*
Copyright (C) 2022 Traefik Labs

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/

package commands

import (
	"context"
	"sort"
	"time"

	"github.com/rs/zerolog/log"
	hubclientset "github.com/traefik/hub-agent-kubernetes/pkg/crd/generated/client/hub/clientset/versioned"
	"github.com/traefik/hub-agent-kubernetes/pkg/platform"
	clientset "k8s.io/client-go/kubernetes"
)

// AnnotationLastPatchRequestedAt is specifies the date at which an update
// has been requested in the RFC-3339 format.
const AnnotationLastPatchRequestedAt = "hub.traefik.io/last-patch-requested-at"

// CommandStore is capable of fetching commands and sending command reports.
type CommandStore interface {
	ListPendingCommands(ctx context.Context) ([]platform.Command, error)
	SendCommandReports(ctx context.Context, reports []platform.CommandReport) error
}

// Watcher watches and applies the patch commands from the platform.
type Watcher struct {
	commands     CommandStore
	k8sClientSet clientset.Interface
	hubClientSet hubclientset.Interface
}

// NewWatcher creates a Watcher.
func NewWatcher(commands CommandStore, k8sClientSet clientset.Interface, hubClientSet hubclientset.Interface) *Watcher {
	return &Watcher{
		commands:     commands,
		k8sClientSet: k8sClientSet,
		hubClientSet: hubClientSet,
	}
}

// Start starts watching commands.
func (w *Watcher) Start(ctx context.Context) {
	tick := time.NewTicker(5 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Ctx(ctx).Info().Msg("Stopping command watcher")
			return
		case <-tick.C:
			w.applyPendingCommands(ctx)
		}
	}
}

func (w *Watcher) applyPendingCommands(ctx context.Context) {
	logger := log.Ctx(ctx)

	commands, err := w.commands.ListPendingCommands(ctx)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to list commands")
		return
	}

	if len(commands) == 0 {
		return
	}

	// Sort commands from the oldest to the newest.
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].CreatedAt.Before(commands[j].CreatedAt)
	})

	var reports []platform.CommandReport

	for _, command := range commands {
		var report *platform.CommandReport

		switch {
		case command.SetIngressACP != nil:
			report = w.setIngressACP(ctx, command.ID, command.CreatedAt, command.SetIngressACP)
		case command.DeleteIngressACP != nil:
			report = w.deleteIngressACP(ctx, command.ID, command.CreatedAt, command.DeleteIngressACP)
		default:
			logger.Error().Msg("Command unsupported on this agent version")
			report = newErrorCommandReport(command.ID, reportErrorTypeUnsupportedCommand)
		}

		if report != nil {
			reports = append(reports, *report)
		}
	}

	if len(reports) == 0 {
		return
	}

	if err = w.commands.SendCommandReports(ctx, reports); err != nil {
		logger.Error().Err(err).Msg("Failed to send command reports")
	}
}
