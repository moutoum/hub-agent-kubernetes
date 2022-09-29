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

import "github.com/traefik/hub-agent-kubernetes/pkg/platform"

type reportErrorType string

const (
	reportErrorTypeInternalError      reportErrorType = "internal-error"
	reportErrorTypeUnsupportedCommand reportErrorType = "unsupported-command"
	reportErrorTypeInvalidIngressID   reportErrorType = "invalid-ingress-id"
	reportErrorTypeIngressNotFound    reportErrorType = "ingress-not-found"
	reportErrorTypeACPNotFound        reportErrorType = "acp-not-found"
)

func newErrorReport(commandID string, typ reportErrorType) *platform.CommandExecutionReport {
	return platform.NewErrorCommandExecutionReport(commandID, platform.CommandExecutionReportError{
		Type: string(typ),
	})
}

func newInternalErrorReport(commandID string, err error) *platform.CommandExecutionReport {
	return platform.NewErrorCommandExecutionReport(commandID, platform.CommandExecutionReportError{
		Type: string(reportErrorTypeInternalError),
		Data: err,
	})
}
