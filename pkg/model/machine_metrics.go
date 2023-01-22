// Copyright Konstantin Bakanov 2023

package model

// MachineMetrics contains metrics reported to us
type MachineMetrics struct {
  ID            string        `json:"id"`
  MachineID     int           `json:"machineId"`
  Stats         MetricsStats  `json:"stats"`
  LastLoggedIn  string        `json:"lastLoggedIn"`
  SysTime       string        `json:"sysTime"`
}

type MetricsStats struct {
  CPUTemp       int   `json:"cpuTemp"`
  FanSpeed      int   `json:"fanSpeed"`
  HDDSpace      int   `json:"HDDSpace"`
  InternalTemp  *int  `json:"internalTemp,omitempty"` // optional field
}

