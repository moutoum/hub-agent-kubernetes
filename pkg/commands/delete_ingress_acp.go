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
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/hub-agent-kubernetes/pkg/acp/admission/reviewer"
	"github.com/traefik/hub-agent-kubernetes/pkg/platform"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (w *Watcher) deleteIngressACP(ctx context.Context, commandID string, requestedAt time.Time, data *platform.DeleteIngressACP) *platform.CommandReport {
	logger := log.Ctx(ctx).With().Str("command_type", "delete_ingress_acp").Logger()
	name, ns, ok := extractNameNamespaceFromIngressID(data.IngressID)
	if !ok {
		logger.Error().Msg("Unable to extract name and namespace from the given IngressID")
		return newErrorCommandReport(commandID, reportErrorTypeInvalidIngressID)
	}

	logger = logger.With().Str("ingress_name", name).Str("ingress_namespace", ns).Logger()

	ingresses := w.k8sClientSet.NetworkingV1().Ingresses(ns)
	ingress, err := ingresses.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if kerror.IsNotFound(err) {
			logger.Error().Err(err).Msg("Ingress not found")
			return newErrorCommandReport(commandID, reportErrorTypeIngressNotFound)
		}

		logger.Error().Err(err).Msg("Unable to find Ingress")
		return newInternalErrorCommandReport(commandID, err)
	}

	var patchedAt time.Time
	if s, ok := ingress.Annotations[AnnotationLastPatchRequestedAt]; ok {
		patchedAt, err = time.Parse(time.RFC3339, s)
		if err != nil {
			logger.Warn().Err(err).Msgf("Unexpected %q annotation format, expected RFC-3339 format. Ignoring annotation", AnnotationLastPatchRequestedAt)
		}
	}

	if requestedAt.Before(patchedAt) || requestedAt.Equal(patchedAt) {
		logger.Debug().Msg("Command already applied. Ignoring")
		return newInternalErrorCommandReport(commandID, fmt.Errorf("operation already executed"))
	}

	mergePatch := ingressPatch{
		ObjectMetadata: objectMetadata{
			Annotations: map[string]*string{
				AnnotationLastPatchRequestedAt: stringPtr(requestedAt.Format(time.RFC3339)),
				reviewer.AnnotationHubAuth:     nil,
			},
		},
	}

	if err = w.patchIngress(ctx, name, ns, mergePatch); err != nil {
		logger.Error().Err(err).Msg("Unable to delete ingress ACP")
		return newInternalErrorCommandReport(commandID, err)
	}

	return platform.NewSuccessCommandReport(commandID)
}
