// SPDX-License-Identifier: GPL-3.0-only

// Package svcmgr provides init system detection and command templates.
package svcmgr

// Backend holds command templates for a service manager.
// Each template contains a single %s verb for the service name.
type Backend struct {
	Name         string
	IsActive     string // exit 0 = running
	IsEnabled    string // exit 0 = enabled at boot
	Start        string
	Stop         string
	Enable       string
	Disable      string
	DaemonReload string // "" if not applicable
	NeedsRoot    bool
}

var backends = map[string]Backend{
	"systemd": {
		Name:         "systemd",
		IsActive:     "systemctl is-active %s",
		IsEnabled:    "systemctl is-enabled %s",
		Start:        "systemctl start %s",
		Stop:         "systemctl stop %s",
		Enable:       "systemctl enable %s",
		Disable:      "systemctl disable %s",
		DaemonReload: "systemctl daemon-reload",
		NeedsRoot:    true,
	},
	"openrc": {
		Name:      "openrc",
		IsActive:  "rc-service %s status",
		IsEnabled: "rc-update show default | grep -q %s",
		Start:     "rc-service %s start",
		Stop:      "rc-service %s stop",
		Enable:    "rc-update add %s default",
		Disable:   "rc-update del %s default",
		NeedsRoot: true,
	},
}
