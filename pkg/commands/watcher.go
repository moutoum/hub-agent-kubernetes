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
	"encoding/json"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/hub-agent-kubernetes/pkg/acp/admission/reviewer"
	"github.com/traefik/hub-agent-kubernetes/pkg/platform"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

// AnnotationLastPatchRequestedAt is specifies the date at which an update
// has been requested in the RFC-3339 format.
const AnnotationLastPatchRequestedAt = "hub.traefik.io/last-patch-requested-at"

// CommandFetcher is capable of fetching commands.
type CommandFetcher interface {
	ListCommands(ctx context.Context) ([]platform.Command, error)
}

// Watcher watches and applies the patch commands from the platform.
type Watcher struct {
	fetcher CommandFetcher
	k8s     kubernetes.Interface
}

// NewWatcher creates a Watcher.
func NewWatcher(fetcher CommandFetcher, k8s kubernetes.Interface) *Watcher {
	return &Watcher{fetcher: fetcher, k8s: k8s}
}

// Start starts watching commands.
func (w *Watcher) Start(ctx context.Context) {
	logger := log.Ctx(ctx)

	tick := time.NewTicker(5 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("Stopping command watcher")
			return
		case <-tick.C:
			commands, err := w.fetcher.ListCommands(ctx)
			if err != nil {
				logger.Warn().Err(err).Msg("Failed to list commands")
				continue
			}

			for _, command := range commands {
				switch {
				case command.SetIngressACP != nil:
					w.setIngressACPCommand(ctx, command.CreatedAt, command.SetIngressACP)
				case command.DeleteIngressACP != nil:
					w.deleteIngressACPCommand(ctx, command.CreatedAt, command.DeleteIngressACP)
				default:
					logger.Error().Msg("Command unsupported on this agent version")
				}
			}
		}
	}
}

type patch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func (w *Watcher) setIngressACPCommand(ctx context.Context, requestedAt time.Time, data *platform.SetIngressACP) {
	logger := log.Ctx(ctx).With().Str("command_type", "set_ingress_acp").Logger()

	name, ns, ok := extractNameNamespaceFromIngressID(data.IngressID)
	if !ok {
		logger.Error().Str("ingress_id", data.IngressID).Msg("Invalid ingress ID format")
		return
	}

	logger = logger.With().Str("ingress_name", name).Str("ingress_namespace", ns).Logger()

	ingresses := w.k8s.NetworkingV1().Ingresses(ns)
	ingress, err := ingresses.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		logger.Error().Err(err).Msg("Unable to get ingress")
		return
	}

	var patchedAt time.Time
	s, ok := ingress.Annotations[AnnotationLastPatchRequestedAt]
	if ok {
		patchedAt, err = time.Parse(time.RFC3339, s)
		if err != nil {
			logger.Warn().Err(err).Msgf("Unexpected %q annotation format, expected RFC-3339 format. Ignoring annotation", AnnotationLastPatchRequestedAt)
		}
	}

	if requestedAt.Before(patchedAt) || requestedAt.Equal(patchedAt) {
		logger.Debug().Msg("Command already applied. Ignoring")
		return
	}

	patches := []patch{
		{
			Op:    "replace",
			Path:  getAnnotationPatchPath(AnnotationLastPatchRequestedAt),
			Value: requestedAt.Format(time.RFC3339),
		},
		{
			Op:    "replace",
			Path:  getAnnotationPatchPath(reviewer.AnnotationHubAuth),
			Value: data.ACPName,
		},
	}

	logger = logger.With().Str("acp_name", data.ACPName).Logger()

	rawPatch, err := json.Marshal(patches)
	if err != nil {
		logger.Error().Err(err).Msg("Unable to set ACP on ingress")
		return
	}

	_, err = ingresses.Patch(ctx, name, types.JSONPatchType, rawPatch, metav1.PatchOptions{})
	if err != nil {
		logger.Error().Err(err).Msg("Unable to set ACP on ingress")
		return
	}
}

func (w *Watcher) deleteIngressACPCommand(ctx context.Context, requestedAt time.Time, data *platform.DeleteIngressACP) {
	logger := log.Ctx(ctx).With().Str("command_type", "delete_ingress_acp").Logger()

	name, ns, ok := extractNameNamespaceFromIngressID(data.IngressID)
	if !ok {
		logger.Error().Str("ingress_id", data.IngressID).Msg("Invalid ingress ID format")
		return
	}

	logger = logger.With().Str("ingress_name", name).Str("ingress_namespace", ns).Logger()

	ingresses := w.k8s.NetworkingV1().Ingresses(ns)
	ingress, err := ingresses.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		logger.Error().Err(err).Msg("Unable to get ingress")
		return
	}

	var patchedAt time.Time
	s, ok := ingress.Annotations[AnnotationLastPatchRequestedAt]
	if ok {
		patchedAt, err = time.Parse(time.RFC3339, s)
		if err != nil {
			logger.Warn().Err(err).Msgf("Unexpected %q annotation format, expected RFC-3339 format. Ignoring annotation", AnnotationLastPatchRequestedAt)
		}
	}

	if requestedAt.Before(patchedAt) || requestedAt.Equal(patchedAt) {
		logger.Debug().Msg("Command already applied. Ignoring")
		return
	}

	patches := []patch{
		{
			Op:    "replace",
			Path:  getAnnotationPatchPath(AnnotationLastPatchRequestedAt),
			Value: requestedAt.Format(time.RFC3339),
		},
		{
			Op:   "remove",
			Path: getAnnotationPatchPath(reviewer.AnnotationHubAuth),
		},
	}

	rawPatch, err := json.Marshal(patches)
	if err != nil {
		logger.Error().Err(err).Msg("Unable to remove ACP from ingress")
		return
	}

	_, err = ingresses.Patch(ctx, name, types.JSONPatchType, rawPatch, metav1.PatchOptions{})
	if err != nil {
		logger.Error().Err(err).Msg("Unable to remove ACP from ingress")
		return
	}
}

// getAnnotationPatchPath returns the json-patch path for resources annotation.
// More info: https://www.rfc-editor.org/rfc/rfc6901#page-2.
func getAnnotationPatchPath(annotation string) string {
	return "/metadata/annotations/" + strings.ReplaceAll(annotation, "/", "~1")
}

func extractNameNamespaceFromIngressID(ingressID string) (name string, namespace string, ok bool) {
	parts := strings.Split(ingressID, ".")
	if len(parts) != 3 {
		return "", "", false
	}

	keyParts := strings.Split(parts[0], "@")
	if len(keyParts) != 2 {
		return "", "", false
	}

	return keyParts[0], keyParts[1], true
}
