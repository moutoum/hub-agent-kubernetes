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
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	kubemock "k8s.io/client-go/kubernetes/fake"
)

func TestWatcher_applyCommands_setIngressACP(t *testing.T) {
	ctx := context.Background()

	ingress := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-ingress",
			Namespace: "my-ns",
			Annotations: map[string]string{
				"something": "somewhere",
			},
		},
	}
	k8sClient := kubemock.NewSimpleClientset(ingress)

	w := NewWatcher(nil, k8sClient)

	now := time.Now().UTC().Truncate(time.Millisecond)
	command := platform.Command{
		CreatedAt: now.Add(-time.Hour),
		SetIngressACP: &platform.SetIngressACP{
			IngressID: "my-ingress@my-ns.ingress.networking.k8s.io",
			ACPName:   "my-acp",
		},
	}

	w.applyCommands(ctx, []platform.Command{command})

	updatedIngress, err := k8sClient.NetworkingV1().
		Ingresses("my-ns").
		Get(ctx, "my-ingress", metav1.GetOptions{})

	require.NoError(t, err)

	wantIngress := ingress
	wantIngress.Annotations["hub.traefik.io/access-control-policy"] = "my-acp"
	wantIngress.Annotations["hub.traefik.io/last-patch-requested-at"] = command.CreatedAt.Format(time.RFC3339)

	assert.Equal(t, wantIngress, updatedIngress)
}

func TestWatcher_applyCommands_setIngressACP_replace(t *testing.T) {
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

	w := NewWatcher(nil, k8sClient)

	command := platform.Command{
		CreatedAt: now,
		SetIngressACP: &platform.SetIngressACP{
			IngressID: "my-ingress@my-ns.ingress.networking.k8s.io",
			ACPName:   "my-acp-2",
		},
	}

	w.applyCommands(ctx, []platform.Command{command})

	updatedIngress, err := k8sClient.NetworkingV1().
		Ingresses("my-ns").
		Get(ctx, "my-ingress", metav1.GetOptions{})

	require.NoError(t, err)

	wantIngress := ingress
	wantIngress.Annotations["hub.traefik.io/access-control-policy"] = "my-acp-2"
	wantIngress.Annotations["hub.traefik.io/last-patch-requested-at"] = command.CreatedAt.Format(time.RFC3339)

	assert.Equal(t, wantIngress, updatedIngress)
}

func TestWatcher_applyCommands_setIngressACP_oldCommand(t *testing.T) {
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

	w := NewWatcher(nil, k8sClient)

	command := platform.Command{
		CreatedAt: now.Add(-2 * time.Hour),
		SetIngressACP: &platform.SetIngressACP{
			IngressID: "my-ingress@my-ns.ingress.networking.k8s.io",
			ACPName:   "my-acp-2",
		},
	}

	w.applyCommands(ctx, []platform.Command{command})

	updatedIngress, err := k8sClient.NetworkingV1().
		Ingresses("my-ns").
		Get(ctx, "my-ingress", metav1.GetOptions{})

	require.NoError(t, err)

	assert.Equal(t, ingress, updatedIngress)
}

func TestWatcher_applyCommands_deleteIngressACP(t *testing.T) {
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

	w := NewWatcher(nil, k8sClient)

	command := platform.Command{
		CreatedAt: now,
		DeleteIngressACP: &platform.DeleteIngressACP{
			IngressID: "my-ingress@my-ns.ingress.networking.k8s.io",
		},
	}

	w.applyCommands(ctx, []platform.Command{command})

	updatedIngress, err := k8sClient.NetworkingV1().
		Ingresses("my-ns").
		Get(ctx, "my-ingress", metav1.GetOptions{})

	require.NoError(t, err)

	wantIngress := ingress
	delete(wantIngress.Annotations, "hub.traefik.io/access-control-policy")
	wantIngress.Annotations["hub.traefik.io/last-patch-requested-at"] = command.CreatedAt.Format(time.RFC3339)

	assert.Equal(t, wantIngress, updatedIngress)
}

func TestWatcher_applyCommands_deleteIngressACP_nothingDoDelete(t *testing.T) {
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

	w := NewWatcher(nil, k8sClient)

	command := platform.Command{
		CreatedAt: now,
		DeleteIngressACP: &platform.DeleteIngressACP{
			IngressID: "my-ingress@my-ns.ingress.networking.k8s.io",
		},
	}

	w.applyCommands(ctx, []platform.Command{command})

	updatedIngress, err := k8sClient.NetworkingV1().
		Ingresses("my-ns").
		Get(ctx, "my-ingress", metav1.GetOptions{})

	require.NoError(t, err)

	wantIngress := ingress
	wantIngress.Annotations["hub.traefik.io/last-patch-requested-at"] = command.CreatedAt.Format(time.RFC3339)

	assert.Equal(t, wantIngress, updatedIngress)
}

func TestWatcher_applyCommands_skipsUnknownCommands(t *testing.T) {
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Millisecond)

	w := NewWatcher(nil, nil)

	// Simulating a command that failed to unmarshal: No data attached to the command.
	command := platform.Command{CreatedAt: now}

	w.applyCommands(ctx, []platform.Command{command})
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

	commands := []platform.Command{
		{
			CreatedAt: now,
			SetIngressACP: &platform.SetIngressACP{
				IngressID: "my-ingress@my-ns.ingress.networking.k8s.io",
				ACPName:   "my-acp-2",
			},
		},
		{
			CreatedAt: now.Add(-2 * time.Hour),
			SetIngressACP: &platform.SetIngressACP{
				IngressID: "my-ingress@my-ns.ingress.networking.k8s.io",
				ACPName:   "my-acp-1",
			},
		},
	}
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

	w := NewWatcher(nil, k8sClient)

	go w.applyCommands(ctx, commands)

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
			if updateCount == len(commands) {
				return
			}

		case <-time.After(100 * time.Second):
			require.Fail(t, "timed out waiting for a command to be executed")
		}
	}
}

func TestExtractNameNamespaceFromIngressID(t *testing.T) {
	tests := []struct {
		desc          string
		ingressID     string
		wantName      string
		wantNamespace string
		wantOK        bool
	}{
		{
			desc:          "group contains more than one dot",
			ingressID:     "whoami-2@default.ingress.networking.k8s.io",
			wantName:      "whoami-2",
			wantNamespace: "default",
			wantOK:        true,
		},
		{
			desc:          "simple group",
			ingressID:     "whoami-2@default.ingress.group",
			wantName:      "whoami-2",
			wantNamespace: "default",
			wantOK:        true,
		},
		{
			desc:      "missing group",
			ingressID: "whoami-2@default.ingress",
			wantOK:    false,
		},
		{
			desc:      "missing namespace",
			ingressID: "whoami-2.ingress.networking.k8s.io",
			wantOK:    false,
		},
		{
			desc:      "not an ingress ID",
			ingressID: "hello",
			wantOK:    false,
		},
		{
			desc:      "empty",
			ingressID: "",
			wantOK:    false,
		},
	}

	for _, test := range tests {
		test := test

		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			gotName, gotNamespace, gotOK := extractNameNamespaceFromIngressID(test.ingressID)
			assert.Equal(t, test.wantName, gotName)
			assert.Equal(t, test.wantNamespace, gotNamespace)
			assert.Equal(t, test.wantOK, gotOK)
		})
	}
}
