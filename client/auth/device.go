package auth

import "github.com/kernel-ai/koscore/utils/types"

type DeviceInfo struct {
	GUID          types.Bytes `json:"guid"`
	DeviceName    string      `json:"device_name"`
	SystemKernel  string      `json:"system_kernel"`
	KernelVersion string      `json:"kernel_version"`
}
