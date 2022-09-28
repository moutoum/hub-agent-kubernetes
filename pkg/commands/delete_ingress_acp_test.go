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

func TestWatcher_deleteIngressACP(t *testing.T) {
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

	w := NewWatcher(nil, k8sClient, nil)

	createdAt := now
	data := &platform.DeleteIngressACP{
		IngressID: "my-ingress@my-ns.ingress.networking.k8s.io",
	}

	report := w.deleteIngressACP(ctx, "command-id", createdAt, data)

	updatedIngress, err := k8sClient.NetworkingV1().
		Ingresses("my-ns").
		Get(ctx, "my-ingress", metav1.GetOptions{})

	require.NoError(t, err)

	wantIngress := ingress
	delete(wantIngress.Annotations, "hub.traefik.io/access-control-policy")
	wantIngress.Annotations["hub.traefik.io/last-patch-requested-at"] = createdAt.Format(time.RFC3339)

	assert.Equal(t, platform.NewSuccessCommandReport("command-id"), report)
	assert.Equal(t, wantIngress, updatedIngress)
}

func TestWatcher_deleteIngressACP_ingressNotFound(t *testing.T) {
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Millisecond)
	k8sClient := kubemock.NewSimpleClientset()

	w := NewWatcher(nil, k8sClient, nil)

	createdAt := now
	data := &platform.DeleteIngressACP{
		IngressID: "my-ingress@my-ns.ingress.networking.k8s.io",
	}

	report := w.deleteIngressACP(ctx, "command-id", createdAt, data)

	assert.Equal(t, platform.NewErrorCommandReport("command-id", platform.CommandReportError{
		Type: "ingress-not-found",
	}), report)
}

func TestWatcher_deleteIngressACP_nothingDoDelete(t *testing.T) {
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

	w := NewWatcher(nil, k8sClient, nil)

	createdAt := now
	data := &platform.DeleteIngressACP{
		IngressID: "my-ingress@my-ns.ingress.networking.k8s.io",
	}

	report := w.deleteIngressACP(ctx, "command-id", createdAt, data)

	updatedIngress, err := k8sClient.NetworkingV1().
		Ingresses("my-ns").
		Get(ctx, "my-ingress", metav1.GetOptions{})

	require.NoError(t, err)

	wantIngress := ingress
	wantIngress.Annotations["hub.traefik.io/last-patch-requested-at"] = createdAt.Format(time.RFC3339)

	assert.Equal(t, platform.NewSuccessCommandReport("command-id"), report)
	assert.Equal(t, wantIngress, updatedIngress)
}
