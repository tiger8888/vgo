package system

import (
	"fmt"

	"github.com/corego/vgo/mecury/agent"
)

type DiskStats struct {
	ps PS

	// Legacy support
	Mountpoints []string

	MountPoints []string
	IgnoreFS    []string `toml:"ignore_fs"`
}

func (_ *DiskStats) Description() string {
	return "Read metrics about disk usage by mount point"
}

var diskSampleConfig = `
  ## By default, telegraf gather stats for all mountpoints.
  ## Setting mountpoints will restrict the stats to the specified mountpoints.
  # mount_points = ["/"]

  ## Ignore some mountpoints by filesystem type. For example (dev)tmpfs (usually
  ## present on /run, /var/run, /dev/shm or /dev).
  ignore_fs = ["tmpfs", "devtmpfs"]
`

func (_ *DiskStats) SampleConfig() string {
	return diskSampleConfig
}

func (s *DiskStats) Gather(acc agent.Accumulator) error {
	// Legacy support:
	if len(s.Mountpoints) != 0 {
		s.MountPoints = s.Mountpoints
	}

	disks, err := s.ps.DiskUsage(s.MountPoints, s.IgnoreFS)
	if err != nil {
		return fmt.Errorf("error getting disk usage info: %s", err)
	}

	for _, du := range disks {
		if du.Total == 0 {
			// Skip dummy filesystem (procfs, cgroupfs, ...)
			continue
		}
		tags := map[string]string{
			"path":   du.Path,
			"fstype": du.Fstype,
		}
		var used_percent float64
		if du.Used+du.Free > 0 {
			used_percent = float64(du.Used) /
				(float64(du.Used) + float64(du.Free)) * 100
		}

		fields := map[string]interface{}{
			"total":        du.Total,
			"free":         du.Free,
			"used":         du.Used,
			"used_percent": used_percent,
			"inodes_total": du.InodesTotal,
			"inodes_free":  du.InodesFree,
			"inodes_used":  du.InodesUsed,
		}
		acc.AddFields("disk", fields, tags)
	}

	return nil
}

func init() {
	agent.AddInput("disk", &DiskStats{ps: &systemPS{}})

}
