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
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/hub-agent-kubernetes/pkg/acp/admission/reviewer"
	"github.com/traefik/hub-agent-kubernetes/pkg/platform"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
)

type ingressPatch struct {
	ObjectMetadata objectMetadata `json:"metadata"`
}

type objectMetadata struct {
	Annotations map[string]*string `json:"annotations"`
}

func (w *Watcher) setIngressACP(ctx context.Context, commandID string, requestedAt time.Time, data *platform.SetIngressACP) *platform.CommandReport {
	logger := log.Ctx(ctx).With().Str("command_type", "set_ingress_acp").Logger()

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
				reviewer.AnnotationHubAuth:     stringPtr(data.ACPName),
			},
		},
	}

	logger = logger.With().Str("acp_name", data.ACPName).Logger()

	exists, err := w.acpExists(ctx, data.ACPName)
	if err != nil {
		logger.Error().Err(err).Msg("Unable to find ACP")
		return newInternalErrorCommandReport(commandID, err)
	}
	if !exists {
		logger.Error().Err(err).Msg("ACP not found")
		return newErrorCommandReport(commandID, reportErrorTypeACPNotFound)
	}

	if err = w.patchIngress(ctx, name, ns, mergePatch); err != nil {
		logger.Error().Err(err).Msg("Unable to set ACP on ingress")
		return newInternalErrorCommandReport(commandID, err)
	}

	return platform.NewSuccessCommandReport(commandID)
}

func (w *Watcher) acpExists(ctx context.Context, acpName string) (bool, error) {
	_, err := w.hubClientSet.HubV1alpha1().AccessControlPolicies().Get(ctx, acpName, metav1.GetOptions{})
	if err != nil {
		if kerror.IsNotFound(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (w *Watcher) patchIngress(ctx context.Context, name, ns string, patch ingressPatch) error {
	rawPatch, err := json.Marshal(patch)
	if err != nil {
		return err
	}

	ingresses := w.k8sClientSet.NetworkingV1().Ingresses(ns)

	_, err = ingresses.Patch(ctx, name, ktypes.MergePatchType, rawPatch, metav1.PatchOptions{})
	if err != nil {
		return err
	}

	return nil
}

func extractNameNamespaceFromIngressID(ingressID string) (name, namespace string, ok bool) {
	parts := strings.Split(ingressID, ".")
	if len(parts) < 3 {
		return "", "", false
	}

	keyParts := strings.Split(parts[0], "@")
	if len(keyParts) != 2 {
		return "", "", false
	}

	return keyParts[0], keyParts[1], true
}

func stringPtr(s string) *string {
	return &s
}
