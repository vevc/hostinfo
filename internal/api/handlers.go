package api

import (
	"encoding/json"
	"net/http"
	"os"
	"runtime"

	"hostinfo/internal/sysinfo"
)

type errorResponse struct {
	Error string `json:"error"`
}

type envKeyResponse struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type indexResponse struct {
	Service     string            `json:"service"`
	Description string            `json:"description"`
	Endpoints   map[string]string `json:"endpoints"`
}

// Handler holds HTTP handlers for server info APIs.
type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, indexResponse{
		Service:     "hostinfo",
		Description: "Host introspection API — environment, user, network, system, memory, disk, and process info.",
		Endpoints: map[string]string{
			"GET /api/health":    "Health check",
			"GET /api/info":      "Aggregated server overview",
			"GET /api/env":       "All environment variables",
			"GET /api/env/{key}": "Single environment variable",
			"GET /api/user":      "Current runtime user (username, uid, gid, home)",
			"GET /api/network":    "Hostname, local IPs, network interfaces",
			"GET /api/public-ip":  "Public IP via outbound probe (requires internet)",
			"GET /api/system":    "OS, platform, kernel, CPU, uptime",
			"GET /api/host":      "Detailed host information",
			"GET /api/cpu":       "CPU details",
			"GET /api/memory":    "System and Go runtime memory",
			"GET /api/disk":      "Disk partitions and usage",
			"GET /api/process":   "Current process info",
			"GET /api/workdir":   "Process working directory",
			"GET /api/hostname":  "Hostname only",
			"GET /api/runtime":   "Go runtime statistics",
		},
	})
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, sysinfo.CollectHealth())
}

func (h *Handler) Info(w http.ResponseWriter, r *http.Request) {
	data, err := sysinfo.CollectOverview()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) Env(w http.ResponseWriter, r *http.Request) {
	env := sysinfo.CollectEnv()
	writeJSON(w, http.StatusOK, env)
}

func (h *Handler) EnvKey(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		writeError(w, http.StatusBadRequest, "missing env key")
		return
	}
	val, ok := sysinfo.CollectEnvKey(key)
	if !ok {
		writeError(w, http.StatusNotFound, "environment variable not found")
		return
	}
	writeJSON(w, http.StatusOK, envKeyResponse{Key: key, Value: val})
}

func (h *Handler) User(w http.ResponseWriter, r *http.Request) {
	data, err := sysinfo.CollectUser()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) Network(w http.ResponseWriter, r *http.Request) {
	data, err := sysinfo.CollectNetwork()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) PublicIP(w http.ResponseWriter, r *http.Request) {
	data := sysinfo.CollectPublicIP()
	if !data.Reachable {
		writeJSON(w, http.StatusServiceUnavailable, data)
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) System(w http.ResponseWriter, r *http.Request) {
	data, err := sysinfo.CollectSystem()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) Host(w http.ResponseWriter, r *http.Request) {
	data, err := sysinfo.CollectHost()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) CPU(w http.ResponseWriter, r *http.Request) {
	data, err := sysinfo.CollectCPU()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) Memory(w http.ResponseWriter, r *http.Request) {
	data, err := sysinfo.CollectMemory()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) Disk(w http.ResponseWriter, r *http.Request) {
	data, err := sysinfo.CollectDisk()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) Process(w http.ResponseWriter, r *http.Request) {
	data, err := sysinfo.CollectProcess()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, data)
}

func (h *Handler) Workdir(w http.ResponseWriter, r *http.Request) {
	wd, err := os.Getwd()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"working_dir": wd})
}

func (h *Handler) Hostname(w http.ResponseWriter, r *http.Request) {
	hostname, err := os.Hostname()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"hostname": hostname})
}

type runtimeStats struct {
	GoVersion    string `json:"go_version"`
	GOOS         string `json:"goos"`
	GOARCH       string `json:"goarch"`
	NumCPU       int    `json:"num_cpu"`
	GOMAXPROCS   int    `json:"gomaxprocs"`
	NumGoroutine int    `json:"num_goroutine"`
	Compiler     string `json:"compiler"`
}

func (h *Handler) Runtime(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, runtimeStats{
		GoVersion:    runtime.Version(),
		GOOS:         runtime.GOOS,
		GOARCH:       runtime.GOARCH,
		NumCPU:       runtime.NumCPU(),
		GOMAXPROCS:   runtime.GOMAXPROCS(0),
		NumGoroutine: runtime.NumGoroutine(),
		Compiler:     runtime.Compiler,
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}
