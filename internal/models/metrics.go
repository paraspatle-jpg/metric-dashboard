package models

import "time"

type Metrics struct {
	ID            int64     `json:"id,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
	CPUPercent    float64   `json:"cpu_percent"`
	MemoryPercent float64   `json:"memory_percent"`
	DiskPercent   float64   `json:"disk_percent"`
	NetBytesSent  uint64    `json:"net_bytes_sent"`
	NetBytesRecv  uint64    `json:"net_bytes_recv"`
	LoadAvg       float64   `json:"load_avg"`
}
