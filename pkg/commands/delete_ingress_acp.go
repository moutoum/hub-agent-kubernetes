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
	"time"

	"github.com/rs/zerolog/log"
	"github.com/traefik/hub-agent-kubernetes/pkg/acp/admission/reviewer"
	"github.com/traefik/hub-agent-kubernetes/pkg/platform"
	kerror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "k8s.io/apimachinery/pkg/types"
	clientset "k8s.io/client-go/kubernetes"
)

// DeleteIngressACPCommand removes the ACP of a given Ingress.
type DeleteIngressACPCommand struct {
	k8sClientSet clientset.Interface
}

// NewDeleteIngressACPCommand creates a new DeleteIngressACPCommand.
func NewDeleteIngressACPCommand(k8sClientSet clientset.Interface) *DeleteIngressACPCommand {
	return &DeleteIngressACPCommand{
		k8sClientSet: k8sClientSet,
	}
}

type deleteIngressACPPayload struct {
	IngressID string `json:"ingressId"`
}

// Handle handles the ACP deletion on the given Ingress.
func (c *DeleteIngressACPCommand) Handle(ctx context.Context, id string, requestedAt time.Time, data json.RawMessage) *platform.CommandReport {
	var payload deleteIngressACPPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("Unable to unmarshal command payload")
		return newInternalErrorCommandReport(id, err)
	}

	logger := log.Ctx(ctx).With().Str("ingress_id", payload.IngressID).Logger()

	name, ns, ok := extractNameNamespaceFromIngressID(payload.IngressID)
	if !ok {
		logger.Error().Msg("Unable to extract name and namespace from the given IngressID")
		return newErrorCommandReport(id, reportErrorTypeInvalidIngressID)
	}

	logger = logger.With().Str("ingress_name", name).Str("ingress_namespace", ns).Logger()

	ingresses := c.k8sClientSet.NetworkingV1().Ingresses(ns)
	ingress, err := ingresses.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if kerror.IsNotFound(err) {
			logger.Error().Err(err).Msg("Ingress not found")
			return newErrorCommandReport(id, reportErrorTypeIngressNotFound)
		}

		logger.Error().Err(err).Msg("Unable to find Ingress")
		return newInternalErrorCommandReport(id, err)
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
		return newInternalErrorCommandReport(id, fmt.Errorf("operation already executed"))
	}

	mergePatch := ingressPatch{
		ObjectMetadata: objectMetadata{
			Annotations: map[string]*string{
				AnnotationLastPatchRequestedAt: stringPtr(requestedAt.Format(time.RFC3339)),
				reviewer.AnnotationHubAuth:     nil,
			},
		},
	}

	if err = c.patchIngress(ctx, name, ns, mergePatch); err != nil {
		logger.Error().Err(err).Msg("Unable to delete ingress ACP")
		return newInternalErrorCommandReport(id, err)
	}

	return platform.NewSuccessCommandReport(id)
}

func (c *DeleteIngressACPCommand) patchIngress(ctx context.Context, name, ns string, patch ingressPatch) error {
	rawPatch, err := json.Marshal(patch)
	if err != nil {
		return err
	}

	ingresses := c.k8sClientSet.NetworkingV1().Ingresses(ns)

	_, err = ingresses.Patch(ctx, name, ktypes.MergePatchType, rawPatch, metav1.PatchOptions{})
	if err != nil {
		return err
	}

	return nil
}
