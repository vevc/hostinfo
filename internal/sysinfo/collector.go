package sysinfo

import (
	"net"
	"os"
	"os/user"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

// Health holds a minimal health-check payload.
type Health struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// Overview is the aggregated server snapshot returned by /api/info.
type Overview struct {
	Hostname  string      `json:"hostname"`
	User      UserInfo    `json:"user"`
	System    SystemInfo  `json:"system"`
	Network   NetworkInfo `json:"network"`
	Memory    MemoryInfo  `json:"memory"`
	Process   ProcessInfo `json:"process"`
	EnvCount  int         `json:"env_count"`
	GoVersion string      `json:"go_version"`
}

// UserInfo describes the effective runtime user.
type UserInfo struct {
	Username      string   `json:"username"`
	Name          string   `json:"name,omitempty"`
	UID           string   `json:"uid"`
	GID           string   `json:"gid"`
	HomeDir       string   `json:"home_dir"`
	TempDir       string   `json:"temp_dir"`
	CacheDir      string   `json:"cache_dir,omitempty"`
	ConfigDir     string   `json:"config_dir,omitempty"`
	GroupIDs      []string `json:"group_ids,omitempty"`
	Shell         string   `json:"shell,omitempty"`
	EffectiveUser string   `json:"effective_user,omitempty"`
}

// SystemInfo covers OS, host, and CPU basics.
type SystemInfo struct {
	OS                 string `json:"os"`
	Arch               string `json:"arch"`
	Platform           string `json:"platform"`
	PlatformFamily     string `json:"platform_family,omitempty"`
	PlatformVersion    string `json:"platform_version,omitempty"`
	KernelVersion      string `json:"kernel_version,omitempty"`
	KernelArch         string `json:"kernel_arch,omitempty"`
	HostID             string `json:"host_id,omitempty"`
	Hostname           string `json:"hostname,omitempty"`
	Virtualization     string `json:"virtualization_system,omitempty"`
	VirtualizationRole string `json:"virtualization_role,omitempty"`
	Timezone           string `json:"timezone"`
	BootTime           int64  `json:"boot_time_unix"`
	UptimeSeconds      uint64 `json:"uptime_seconds"`
	CPUModel           string `json:"cpu_model,omitempty"`
	CPUCores           int32  `json:"cpu_cores"`
	LogicalCPUs        int    `json:"logical_cpus"`
	GoOS               string `json:"go_os"`
	GoArch             string `json:"go_arch"`
	GoVersion          string `json:"go_version"`
	MaxProcs           int    `json:"max_procs"`
	Goroutines         int    `json:"goroutines"`
}

// NetworkInfo lists hostname and interface details.
type NetworkInfo struct {
	Hostname   string      `json:"hostname"`
	FQDN       string      `json:"fqdn,omitempty"`
	PrimaryIPs []string    `json:"primary_ips"`
	Interfaces []Interface `json:"interfaces"`
}

// Interface describes one network interface.
type Interface struct {
	Name         string   `json:"name"`
	HardwareAddr string   `json:"hardware_addr,omitempty"`
	Flags        []string `json:"flags,omitempty"`
	MTU          int      `json:"mtu,omitempty"`
	Addrs        []string `json:"addrs"`
}

// MemoryInfo covers virtual and swap memory plus Go runtime stats.
type MemoryInfo struct {
	TotalBytes        uint64  `json:"total_bytes"`
	AvailableBytes    uint64  `json:"available_bytes"`
	UsedBytes         uint64  `json:"used_bytes"`
	UsedPercent       float64 `json:"used_percent"`
	FreeBytes         uint64  `json:"free_bytes"`
	SwapTotalBytes    uint64  `json:"swap_total_bytes"`
	SwapUsedBytes     uint64  `json:"swap_used_bytes"`
	SwapFreeBytes     uint64  `json:"swap_free_bytes"`
	GoAllocBytes      uint64  `json:"go_alloc_bytes"`
	GoTotalAllocBytes uint64  `json:"go_total_alloc_bytes"`
	GoSysBytes        uint64  `json:"go_sys_bytes"`
	GoNumGC           uint32  `json:"go_num_gc"`
}

// ProcessInfo describes the current process.
type ProcessInfo struct {
	PID         int      `json:"pid"`
	PPID        int32    `json:"ppid,omitempty"`
	Executable  string   `json:"executable,omitempty"`
	WorkingDir  string   `json:"working_dir"`
	Args        []string `json:"args"`
	EnvCount    int      `json:"env_count"`
	StartTime   int64    `json:"start_time_unix,omitempty"`
	CPUPercent  float64  `json:"cpu_percent,omitempty"`
	MemoryBytes uint64   `json:"memory_bytes,omitempty"`
}

// DiskPartition describes one mounted filesystem.
type DiskPartition struct {
	Device      string  `json:"device"`
	Mountpoint  string  `json:"mountpoint"`
	FSType      string  `json:"fs_type"`
	TotalBytes  uint64  `json:"total_bytes"`
	UsedBytes   uint64  `json:"used_bytes"`
	FreeBytes   uint64  `json:"free_bytes"`
	UsedPercent float64 `json:"used_percent"`
}

// EnvMap is a sorted key-value map of environment variables.
type EnvMap map[string]string

// CollectHealth returns a health payload.
func CollectHealth() Health {
	return Health{
		Status:    "ok",
		Timestamp: time.Now().UTC(),
	}
}

// CollectOverview aggregates the most useful server fields.
func CollectOverview() (Overview, error) {
	hostname, _ := os.Hostname()
	userInfo, _ := CollectUser()
	systemInfo, _ := CollectSystem()
	networkInfo, _ := CollectNetwork()
	memoryInfo, _ := CollectMemory()
	processInfo, _ := CollectProcess()

	return Overview{
		Hostname:  hostname,
		User:      userInfo,
		System:    systemInfo,
		Network:   networkInfo,
		Memory:    memoryInfo,
		Process:   processInfo,
		EnvCount:  len(os.Environ()),
		GoVersion: runtime.Version(),
	}, nil
}

// CollectEnv returns all environment variables sorted by key.
func CollectEnv() EnvMap {
	env := make(EnvMap)
	for _, entry := range os.Environ() {
		if key, val, ok := strings.Cut(entry, "="); ok {
			env[key] = val
		}
	}
	return env
}

// CollectEnvKey returns a single environment variable.
func CollectEnvKey(key string) (string, bool) {
	val, ok := os.LookupEnv(key)
	return val, ok
}

// CollectUser returns current user information.
func CollectUser() (UserInfo, error) {
	info := UserInfo{}

	u, err := user.Current()
	if err == nil {
		info.Username = u.Username
		info.Name = u.Name
		info.UID = u.Uid
		info.GID = u.Gid
		info.HomeDir = u.HomeDir
	}

	if home, err := os.UserHomeDir(); err == nil && info.HomeDir == "" {
		info.HomeDir = home
	}
	info.TempDir = os.TempDir()
	if cache, err := os.UserCacheDir(); err == nil {
		info.CacheDir = cache
	}
	if config, err := os.UserConfigDir(); err == nil {
		info.ConfigDir = config
	}

	if groups, err := os.Getgroups(); err == nil {
		info.GroupIDs = intSliceToStringSlice(groups)
	}

	info.EffectiveUser = os.Getenv("USER")
	if info.EffectiveUser == "" {
		info.EffectiveUser = os.Getenv("USERNAME")
	}
	info.Shell = os.Getenv("SHELL")
	if info.Shell == "" {
		info.Shell = os.Getenv("COMSPEC")
	}

	return info, nil
}

// CollectSystem returns OS, host, and CPU information.
func CollectSystem() (SystemInfo, error) {
	info := SystemInfo{
		OS:          runtime.GOOS,
		Arch:        runtime.GOARCH,
		GoOS:        runtime.GOOS,
		GoArch:      runtime.GOARCH,
		GoVersion:   runtime.Version(),
		MaxProcs:    runtime.GOMAXPROCS(0),
		Goroutines:  runtime.NumGoroutine(),
		LogicalCPUs: runtime.NumCPU(),
		Timezone:    timezoneName(),
	}

	if h, err := host.Info(); err == nil {
		info.Platform = h.Platform
		info.PlatformFamily = h.PlatformFamily
		info.PlatformVersion = h.PlatformVersion
		info.KernelVersion = h.KernelVersion
		info.KernelArch = h.KernelArch
		info.HostID = h.HostID
		info.Hostname = h.Hostname
		info.Virtualization = h.VirtualizationSystem
		info.VirtualizationRole = h.VirtualizationRole
		info.BootTime = int64(h.BootTime)
		info.UptimeSeconds = h.Uptime
	}

	if cpus, err := cpu.Info(); err == nil && len(cpus) > 0 {
		info.CPUModel = cpus[0].ModelName
		info.CPUCores = cpus[0].Cores
	}

	return info, nil
}

// CollectNetwork returns hostname, IPs, and interface details.
func CollectNetwork() (NetworkInfo, error) {
	info := NetworkInfo{}

	hostname, err := os.Hostname()
	if err == nil {
		info.Hostname = hostname
	}

	if addrs, err := net.LookupHost(hostname); err == nil && len(addrs) > 0 {
		info.FQDN = hostname
	}

	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			if iface.Flags&net.FlagUp == 0 {
				continue
			}
			item := Interface{
				Name:         iface.Name,
				HardwareAddr: iface.HardwareAddr.String(),
				MTU:          iface.MTU,
				Flags:        flagsToStrings(iface.Flags),
			}
			addrs, err := iface.Addrs()
			if err == nil {
				for _, addr := range addrs {
					item.Addrs = append(item.Addrs, addr.String())
				}
			}
			info.Interfaces = append(info.Interfaces, item)
		}
	}

	info.PrimaryIPs = collectPrimaryIPs()
	sort.Strings(info.PrimaryIPs)
	return info, nil
}

// CollectMemory returns system and Go runtime memory stats.
func CollectMemory() (MemoryInfo, error) {
	info := MemoryInfo{}

	if vm, err := mem.VirtualMemory(); err == nil {
		info.TotalBytes = vm.Total
		info.AvailableBytes = vm.Available
		info.UsedBytes = vm.Used
		info.UsedPercent = vm.UsedPercent
		info.FreeBytes = vm.Free
	}

	if sm, err := mem.SwapMemory(); err == nil {
		info.SwapTotalBytes = sm.Total
		info.SwapUsedBytes = sm.Used
		info.SwapFreeBytes = sm.Free
	}

	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	info.GoAllocBytes = ms.Alloc
	info.GoTotalAllocBytes = ms.TotalAlloc
	info.GoSysBytes = ms.Sys
	info.GoNumGC = ms.NumGC

	return info, nil
}

// CollectProcess returns information about the running server process.
func CollectProcess() (ProcessInfo, error) {
	info := ProcessInfo{
		PID:      os.Getpid(),
		PPID:     int32(os.Getppid()),
		Args:     os.Args,
		EnvCount: len(os.Environ()),
	}

	if wd, err := os.Getwd(); err == nil {
		info.WorkingDir = wd
	}

	if exe, err := os.Executable(); err == nil {
		info.Executable = exe
	}

	return info, nil
}

// CollectDisk returns usage for all mounted partitions.
func CollectDisk() ([]DiskPartition, error) {
	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}

	result := make([]DiskPartition, 0, len(partitions))
	for _, part := range partitions {
		usage, err := disk.Usage(part.Mountpoint)
		if err != nil {
			continue
		}
		result = append(result, DiskPartition{
			Device:      part.Device,
			Mountpoint:  part.Mountpoint,
			FSType:      part.Fstype,
			TotalBytes:  usage.Total,
			UsedBytes:   usage.Used,
			FreeBytes:   usage.Free,
			UsedPercent: usage.UsedPercent,
		})
	}
	return result, nil
}

// CollectCPU returns per-CPU details.
func CollectCPU() ([]cpu.InfoStat, error) {
	return cpu.Info()
}

// CollectHost returns raw host information from gopsutil.
func CollectHost() (*host.InfoStat, error) {
	return host.Info()
}

func collectPrimaryIPs() []string {
	var ips []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return ips
	}
	seen := make(map[string]struct{})
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}
			s := ip.String()
			if _, ok := seen[s]; !ok {
				seen[s] = struct{}{}
				ips = append(ips, s)
			}
		}
	}
	return ips
}

func flagsToStrings(flags net.Flags) []string {
	var result []string
	flagMap := map[net.Flags]string{
		net.FlagUp:           "up",
		net.FlagBroadcast:    "broadcast",
		net.FlagLoopback:     "loopback",
		net.FlagPointToPoint: "pointtopoint",
		net.FlagMulticast:    "multicast",
	}
	for f, name := range flagMap {
		if flags&f != 0 {
			result = append(result, name)
		}
	}
	sort.Strings(result)
	return result
}

func intSliceToStringSlice(in []int) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = strconv.Itoa(v)
	}
	return out
}

func timezoneName() string {
	name, offset := time.Now().Zone()
	return name + " (UTC" + formatOffset(offset) + ")"
}

func formatOffset(seconds int) string {
	sign := "+"
	if seconds < 0 {
		sign = "-"
		seconds = -seconds
	}
	h := seconds / 3600
	m := (seconds % 3600) / 60
	return sign + strconv.Itoa(h) + ":" + pad2(m)
}

func pad2(n int) string {
	if n < 10 {
		return "0" + strconv.Itoa(n)
	}
	return strconv.Itoa(n)
}
