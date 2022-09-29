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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/hub-agent-kubernetes/pkg/platform"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubemock "k8s.io/client-go/kubernetes/fake"
)

func TestDeleteIngressACPCommand_Handle_success(t *testing.T) {
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Millisecond)

	ingress := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-ingress",
			Namespace: "my-ns",
			Annotations: map[string]string{
				"something":                              "somewhere",
				"hub.traefik.io/access-control-policy":   "my-acp",
				"hub.traefik.io/last-patch-requested-at": now.Add(-time.Hour).Format(time.RFC3339),
			},
		},
	}

	k8sClient := kubemock.NewSimpleClientset(ingress)
	handler := NewDeleteIngressACPCommand(k8sClient)

	createdAt := now
	data := []byte(`{"ingressId": "my-ingress@my-ns.ingress.networking.k8s.io"}`)

	report := handler.Handle(ctx, "command-id", createdAt, data)

	updatedIngress, err := k8sClient.NetworkingV1().
		Ingresses("my-ns").
		Get(ctx, "my-ingress", metav1.GetOptions{})

	require.NoError(t, err)

	wantIngress := ingress
	delete(wantIngress.Annotations, "hub.traefik.io/access-control-policy")
	wantIngress.Annotations["hub.traefik.io/last-patch-requested-at"] = createdAt.Format(time.RFC3339)

	assert.Equal(t, platform.NewSuccessCommandExecutionReport("command-id"), report)
	assert.Equal(t, wantIngress, updatedIngress)
}

func TestDeleteIngressACPCommand_Handle_ingressNotFound(t *testing.T) {
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Millisecond)
	k8sClient := kubemock.NewSimpleClientset()

	handler := NewDeleteIngressACPCommand(k8sClient)

	createdAt := now
	data := []byte(`{"ingressId": "my-ingress@my-ns.ingress.networking.k8s.io"}`)

	report := handler.Handle(ctx, "command-id", createdAt, data)

	assert.Equal(t, platform.NewErrorCommandExecutionReport("command-id", platform.CommandExecutionReportError{
		Type: "ingress-not-found",
	}), report)
}

func TestDeleteIngressACPCommand_Handle_nothingDoDelete(t *testing.T) {
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Millisecond)

	ingress := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-ingress",
			Namespace: "my-ns",
			Annotations: map[string]string{
				"something":                              "somewhere",
				"hub.traefik.io/last-patch-requested-at": now.Add(-time.Hour).Format(time.RFC3339),
			},
		},
	}
	k8sClient := kubemock.NewSimpleClientset(ingress)
	handler := NewDeleteIngressACPCommand(k8sClient)

	createdAt := now
	data := []byte(`{"ingressId": "my-ingress@my-ns.ingress.networking.k8s.io"}`)

	report := handler.Handle(ctx, "command-id", createdAt, data)

	updatedIngress, err := k8sClient.NetworkingV1().
		Ingresses("my-ns").
		Get(ctx, "my-ingress", metav1.GetOptions{})

	require.NoError(t, err)

	wantIngress := ingress
	wantIngress.Annotations["hub.traefik.io/last-patch-requested-at"] = createdAt.Format(time.RFC3339)

	assert.Equal(t, platform.NewSuccessCommandExecutionReport("command-id"), report)
	assert.Equal(t, wantIngress, updatedIngress)
}

func TestDeleteIngressACPCommand_Handle_invalidPayload(t *testing.T) {
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Millisecond)

	handler := NewDeleteIngressACPCommand(nil)

	createdAt := now
	data := []byte(`invalid payload`)

	report := handler.Handle(ctx, "command-id", createdAt, data)

	assert.Equal(t, platform.CommandExecutionStatusFailure, report.Status)
	assert.NotNil(t, report.Error)
	assert.Equal(t, "internal-error", report.Error.Type)
	assert.NotEmpty(t, report.Error.Data)
}
