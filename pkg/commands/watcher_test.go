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

	"github.com/traefik/hub-agent-kubernetes/pkg/crd/api/hub/v1alpha1"
	hubmock "github.com/traefik/hub-agent-kubernetes/pkg/crd/generated/client/hub/clientset/versioned/fake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"

	"github.com/traefik/hub-agent-kubernetes/pkg/platform"
	kubemock "k8s.io/client-go/kubernetes/fake"
)

func TestWatcher_applyCommands_skipsUnknownCommands(t *testing.T) {
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Millisecond)

	k8sClient := kubemock.NewSimpleClientset()
	commands := newCommandStoreMock(t)

	pendingCommands := []platform.Command{
		{
			ID:        "command-1",
			CreatedAt: now.Add(-time.Minute),
		},
		{
			ID:        "command-2",
			CreatedAt: now,
			DeleteIngressACP: &platform.DeleteIngressACP{
				IngressID: "my-ingress@my-ns.ingress.networking.k8s.io",
			},
		},
	}
	commands.OnListPendingCommands().TypedReturns(pendingCommands, nil).Once()

	commands.OnSendCommandReports([]platform.CommandReport{
		*platform.NewErrorCommandReport("command-1", platform.CommandReportError{
			Type: string(reportErrorTypeUnsupportedCommand),
		}),
		*platform.NewErrorCommandReport("command-2", platform.CommandReportError{
			Type: string(reportErrorTypeIngressNotFound),
		}),
	}).TypedReturns(nil).Once()

	w := NewWatcher(commands, k8sClient, nil)

	w.applyPendingCommands(ctx)
}

func TestWatcher_applyCommands_appliedByDate(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)

	k8sClient := kubemock.NewSimpleClientset(&netv1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        "my-ingress",
			Namespace:   "my-ns",
			Annotations: map[string]string{},
		},
	})

	hubClient := hubmock.NewSimpleClientset(
		&v1alpha1.AccessControlPolicy{ObjectMeta: metav1.ObjectMeta{Name: "my-acp-1"}},
		&v1alpha1.AccessControlPolicy{ObjectMeta: metav1.ObjectMeta{Name: "my-acp-2"}})

	pendingCommands := []platform.Command{
		{
			ID:        "command-2",
			CreatedAt: now,
			SetIngressACP: &platform.SetIngressACP{
				IngressID: "my-ingress@my-ns.ingress.networking.k8s.io",
				ACPName:   "my-acp-2",
			},
		},
		{
			ID:        "command-1",
			CreatedAt: now.Add(-2 * time.Hour),
			SetIngressACP: &platform.SetIngressACP{
				IngressID: "my-ingress@my-ns.ingress.networking.k8s.io",
				ACPName:   "my-acp-1",
			},
		},
	}

	commands := newCommandStoreMock(t)
	commands.OnListPendingCommands().TypedReturns(pendingCommands, nil).Once()

	commands.OnSendCommandReports([]platform.CommandReport{
		*platform.NewSuccessCommandReport("command-1"),
		*platform.NewSuccessCommandReport("command-2"),
	}).TypedReturns(nil).Once()

	wantAnnotationUpdates := []map[string]string{
		{
			"hub.traefik.io/access-control-policy":   "my-acp-1",
			"hub.traefik.io/last-patch-requested-at": now.Add(-2 * time.Hour).Format(time.RFC3339),
		},
		{
			"hub.traefik.io/access-control-policy":   "my-acp-2",
			"hub.traefik.io/last-patch-requested-at": now.Format(time.RFC3339),
		},
	}

	w := NewWatcher(commands, k8sClient, hubClient)

	go w.applyPendingCommands(ctx)

	// Watch for ingresses events and make sure updates are made in the right order.
	ingressWatcher, err := k8sClient.NetworkingV1().
		Ingresses("my-ns").
		Watch(ctx, metav1.ListOptions{})
	require.NoError(t, err)

	eventCh := ingressWatcher.ResultChan()

	var updateCount int
	for {
		select {
		case event := <-eventCh:
			assert.Equal(t, watch.Modified, event.Type)
			metadata, err := meta.Accessor(event.Object)
			require.NoError(t, err)

			assert.Equal(t, "my-ingress", metadata.GetName())
			assert.Equal(t, "my-ns", metadata.GetNamespace())
			assert.Equal(t, wantAnnotationUpdates[updateCount], metadata.GetAnnotations())

			updateCount++
			if updateCount == len(pendingCommands) {
				return
			}

		case <-time.After(100 * time.Millisecond):
			require.Fail(t, "timed out waiting for a command to be executed")
		}
	}
}
