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
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	hubv1alpha1 "github.com/traefik/hub-agent-kubernetes/pkg/crd/api/hub/v1alpha1"
	hubkubemock "github.com/traefik/hub-agent-kubernetes/pkg/crd/generated/client/hub/clientset/versioned/fake"
	"github.com/traefik/hub-agent-kubernetes/pkg/platform"
	netv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubemock "k8s.io/client-go/kubernetes/fake"
)

func TestSetIngressACPCommand_Handle_success(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)

	ingress := &netv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-ingress",
			Namespace: "my-ns",
			Annotations: map[string]string{
				"something": "somewhere",
			},
		},
	}
	basicAuth := &hubv1alpha1.AccessControlPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-acp",
		},
		Spec: hubv1alpha1.AccessControlPolicySpec{
			BasicAuth: &hubv1alpha1.AccessControlPolicyBasicAuth{
				Users: []string{"user:pass"},
			},
		},
	}

	k8sClient := kubemock.NewSimpleClientset(ingress)
	hubClient := hubkubemock.NewSimpleClientset(basicAuth)

	handler := NewSetIngressACPCommand(k8sClient, hubClient)

	createdAt := now.Add(-time.Hour)
	data := []byte(`{"ingressId": "my-ingress@my-ns.ingress.networking.k8s.io", "acpName": "my-acp"}`)

	report := handler.Handle(ctx, "command-id", createdAt, data)

	updatedIngress, err := k8sClient.NetworkingV1().
		Ingresses("my-ns").
		Get(ctx, "my-ingress", metav1.GetOptions{})

	require.NoError(t, err)

	wantIngress := ingress
	wantIngress.Annotations["hub.traefik.io/access-control-policy"] = "my-acp"
	wantIngress.Annotations["hub.traefik.io/last-patch-requested-at"] = createdAt.Format(time.RFC3339)

	assert.Equal(t, platform.NewSuccessCommandExecutionReport("command-id"), report)
	assert.Equal(t, wantIngress, updatedIngress)
}

func TestSetIngressACPCommand_Handle_ingressNotFound(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)

	k8sClient := kubemock.NewSimpleClientset()
	hubClient := hubkubemock.NewSimpleClientset()

	createdAt := now.Add(-time.Hour)
	data := []byte(`{"ingressId": "my-ingress@my-ns.ingress.networking.k8s.io", "acpName": "my-acp"}`)

	handler := NewSetIngressACPCommand(k8sClient, hubClient)

	report := handler.Handle(ctx, "command-id", createdAt, data)

	assert.Equal(t, platform.NewErrorCommandExecutionReport("command-id", platform.CommandExecutionReportError{
		Type: "ingress-not-found",
	}), report)
}

func TestSetIngressACPCommand_Handle_acpNotFound(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Millisecond)

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
	hubClient := hubkubemock.NewSimpleClientset()

	createdAt := now.Add(-time.Hour)
	data := []byte(`{"ingressId": "my-ingress@my-ns.ingress.networking.k8s.io", "acpName": "my-acp"}`)

	handler := NewSetIngressACPCommand(k8sClient, hubClient)

	report := handler.Handle(ctx, "command-id", createdAt, data)

	assert.Equal(t, platform.NewErrorCommandExecutionReport("command-id", platform.CommandExecutionReportError{
		Type: "acp-not-found",
	}), report)
}

func TestSetIngressACPCommand_Handle_replace(t *testing.T) {
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

	basicAuth := &hubv1alpha1.AccessControlPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-acp-2",
		},
		Spec: hubv1alpha1.AccessControlPolicySpec{
			BasicAuth: &hubv1alpha1.AccessControlPolicyBasicAuth{
				Users: []string{"user:pass"},
			},
		},
	}

	k8sClient := kubemock.NewSimpleClientset(ingress)
	hubClient := hubkubemock.NewSimpleClientset(basicAuth)

	createdAt := now
	data := []byte(`{"ingressId": "my-ingress@my-ns.ingress.networking.k8s.io", "acpName": "my-acp-2"}`)

	handler := NewSetIngressACPCommand(k8sClient, hubClient)

	report := handler.Handle(ctx, "command-id", createdAt, data)

	updatedIngress, err := k8sClient.NetworkingV1().
		Ingresses("my-ns").
		Get(ctx, "my-ingress", metav1.GetOptions{})

	require.NoError(t, err)

	wantIngress := ingress
	wantIngress.Annotations["hub.traefik.io/access-control-policy"] = "my-acp-2"
	wantIngress.Annotations["hub.traefik.io/last-patch-requested-at"] = createdAt.Format(time.RFC3339)

	assert.Equal(t, platform.NewSuccessCommandExecutionReport("command-id"), report)
	assert.Equal(t, wantIngress, updatedIngress)
}

func TestSetIngressACPCommand_Handle_oldCommand(t *testing.T) {
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

	createdAt := now.Add(-2 * time.Hour)
	data := []byte(`{"ingressId": "my-ingress@my-ns.ingress.networking.k8s.io", "acpName": "my-acp-2"}`)

	handler := NewSetIngressACPCommand(k8sClient, nil)

	report := handler.Handle(ctx, "command-id", createdAt, data)

	updatedIngress, err := k8sClient.NetworkingV1().
		Ingresses("my-ns").
		Get(ctx, "my-ingress", metav1.GetOptions{})

	require.NoError(t, err)

	assert.Equal(t, ingress, updatedIngress)
	assert.Equal(t, platform.NewErrorCommandExecutionReport("command-id", platform.CommandExecutionReportError{
		Type: "internal-error",
		Data: errors.New("operation already executed"),
	}), report)
}

func TestSetIngressACPCommand_Handle_invalidPayload(t *testing.T) {
	ctx := context.Background()

	now := time.Now().UTC().Truncate(time.Millisecond)

	handler := NewSetIngressACPCommand(nil, nil)

	createdAt := now
	data := []byte(`invalid payload`)

	report := handler.Handle(ctx, "command-id", createdAt, data)

	assert.Equal(t, platform.CommandExecutionStatusFailure, report.Status)
	assert.NotNil(t, report.Error)
	assert.Equal(t, "internal-error", report.Error.Type)
	assert.NotEmpty(t, report.Error.Data)
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
