// SPDX-License-Identifier: GPL-3.0-only

package lxc

import (
	"fmt"
	"strconv"
	"strings"
)

// pctListEntry represents one row from `pct list` output.
type pctListEntry struct {
	VMID   int
	Status string
	Name   string
}

// parsePctList parses the tabular output of `pct list`.
//
//	VMID       Status     Lock         Name
//	100        running                 pihole
//	101        stopped                 dns
func parsePctList(output string) map[int]pctListEntry {
	entries := make(map[int]pctListEntry)
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "VMID") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		vmid, err := strconv.Atoi(fields[0])
		if err != nil {
			continue
		}
		// Fields: VMID, Status, [Lock], Name
		// Lock column may be empty — name is always last.
		entries[vmid] = pctListEntry{
			VMID:   vmid,
			Status: fields[1],
			Name:   fields[len(fields)-1],
		}
	}
	return entries
}

// parsePctStatus parses `pct status <id>` output.
// Expected format: "status: running" or "status: stopped"
func parsePctStatus(output string) string {
	_, status, ok := strings.Cut(strings.TrimSpace(output), ": ")
	if !ok {
		return ""
	}
	return status
}

// formatNet0 builds the --net0 value for pct create/set.
//
//	name=eth0,bridge=vmbr0,ip=10.10.10.10/24,gw=10.10.10.1,type=veth
func formatNet0(net LxcNet) string {
	var b strings.Builder
	b.WriteString("name=eth0")

	bridge := net.Bridge
	if bridge == "" {
		bridge = "vmbr0"
	}
	b.WriteString(",bridge=")
	b.WriteString(bridge)

	b.WriteString(",ip=")
	b.WriteString(net.IP)

	if net.Gw != "" {
		b.WriteString(",gw=")
		b.WriteString(net.Gw)
	}

	b.WriteString(",type=veth")
	return b.String()
}

// normalizeSizeGiB strips a trailing "G" or "g" suffix so "4G" becomes "4".
// PVE rootfs sizes are always in GiB without a unit suffix.
func normalizeSizeGiB(s string) string {
	return strings.TrimRight(s, "Gg")
}

// buildCreateCmd builds the full `pct create` command.
// Template storage and rootfs storage are independent pools.
func buildCreateCmd(cfg lxcAction) string {
	return fmt.Sprintf("pct create %d %s"+
		" --hostname %s"+
		" --cores %d"+
		" --memory %d"+
		" --rootfs %s:%s"+
		" --net0 %s"+
		" --unprivileged 1"+
		" --password yolo123",
		cfg.id, cfg.template.templatePath(),
		cfg.hostname,
		cfg.cores,
		cfg.memory,
		cfg.storage, normalizeSizeGiB(cfg.size),
		formatNet0(cfg.network),
	)
}

// buildDownloadCmd builds the `pveam download` command from a parsed template.
func buildDownloadCmd(storage, filename string) string {
	return fmt.Sprintf("pveam download %s %s", storage, filename)
}
