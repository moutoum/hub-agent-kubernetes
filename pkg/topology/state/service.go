package state

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

func (f *Fetcher) getServices(apps map[string]*App) (map[string]*Service, map[string]string, error) {
	services, err := f.k8s.Core().V1().Services().Lister().List(labels.Everything())
	if err != nil {
		return nil, nil, err
	}

	svcs := make(map[string]*Service)
	traefikNames := make(map[string]string)
	for _, service := range services {
		svcName := objectKey(service.Name, service.Namespace)
		svcs[svcName] = &Service{
			Name:      service.Name,
			Namespace: service.Namespace,
			Selector:  service.Spec.Selector,
			Apps:      selectApps(apps, service),
			Type:      service.Spec.Type,
			status:    service.Status,
		}

		for _, key := range traefikServiceNames(service) {
			traefikNames[key] = svcName
		}
	}

	return svcs, traefikNames, nil
}

func traefikServiceNames(svc *corev1.Service) []string {
	var result []string
	for _, port := range svc.Spec.Ports {
		result = append(result,
			fmt.Sprintf("%s-%s-%d", svc.Namespace, svc.Name, port.Port),
			fmt.Sprintf("%s-%s-%s", svc.Namespace, svc.Name, port.Name),
		)
	}
	return result
}

func selectApps(apps map[string]*App, service *corev1.Service) []string {
	if service.Spec.Type == corev1.ServiceTypeExternalName {
		return nil
	}

	var result []string
	for key, app := range apps {
		if app.Namespace != service.Namespace {
			continue
		}

		var match bool
		for k, v := range service.Spec.Selector {
			if app.podLabels[k] != v {
				match = false
				break
			}
			match = true
		}

		if match {
			result = append(result, key)
		}
	}

	sort.Strings(result)

	return result
}

// GetServiceLogs returns the logs from a service.
func (f *Fetcher) GetServiceLogs(ctx context.Context, namespace, name string, lines, maxLen int) ([]byte, error) {
	service, err := f.k8s.Core().V1().Services().Lister().Services(namespace).Get(name)
	if err != nil {
		return nil, fmt.Errorf("invalid service %s/%s: %w", name, namespace, err)
	}

	pods, err := f.k8s.Core().V1().Pods().Lister().Pods(namespace).List(labels.SelectorFromSet(service.Spec.Selector))
	if err != nil {
		return nil, fmt.Errorf("list pods for %s/%s: %w", namespace, name, err)
	}

	if len(pods) == 0 {
		return nil, nil
	}
	if len(pods) > lines {
		pods = pods[:lines]
	}

	buf := bytes.NewBuffer(make([]byte, 0, maxLen*lines))
	podLogOpts := corev1.PodLogOptions{Previous: false, TailLines: int64Ptr(int64(lines / len(pods)))}
	for _, pod := range pods {
		req := f.clientSet.CoreV1().Pods(service.Namespace).GetLogs(pod.Name, &podLogOpts)
		podLogs, logErr := req.Stream(ctx)
		if logErr != nil {
			return nil, fmt.Errorf("opening pod log stream: %w", logErr)
		}

		r := bufio.NewReader(podLogs)
		for {
			b, readErr := r.ReadBytes('\n')
			if readErr != nil {
				if errors.Is(readErr, io.EOF) {
					writeBytes(buf, b, maxLen)
					break
				}
				return nil, err
			}

			writeBytes(buf, b, maxLen)
		}

		if err = podLogs.Close(); err != nil {
			return nil, fmt.Errorf("closing pod log stream: %w", err)
		}
	}

	return buf.Bytes(), nil
}

func writeBytes(buf *bytes.Buffer, b []byte, maxLen int) {
	switch {
	case len(b) == 0:
		return
	case len(b) > maxLen:
		b = b[:maxLen-1]
		b = append(b, '\n')
	case b[len(b)-1] != byte('\n'):
		b = append(b, '\n')
	}

	buf.Write(b)
}

func int64Ptr(v int64) *int64 {
	return &v
}
