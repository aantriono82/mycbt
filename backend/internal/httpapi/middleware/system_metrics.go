package middleware

import (
	"math"
	"runtime"
	"runtime/metrics"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

type systemMetricsState struct {
	activeSSEConnections int64
	apiLatencyTotalNs    int64
	apiLatencyCount      int64
	cpuMu                sync.Mutex
	lastCPUSeconds       float64
	lastCPUWallTime      time.Time
	queueMu              sync.Mutex
	queue                queueState
	queueHistory         []queueHistoryItem
	queueBatchSeq        int64
}

var systemMetrics = &systemMetricsState{}

type queueState struct {
	BatchID     int64
	Status      string
	BatchType   string
	Pending     int
	Processing  int
	Success     int
	Failed      int
	LastError   string
	LastUpdated time.Time
}

type queueHistoryItem struct {
	BatchID     int64
	BatchType   string
	Status      string
	TotalJobs   int
	Success     int
	Failed      int
	LastError   string
	StartedAt   time.Time
	CompletedAt time.Time
}

func RequestMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start).Nanoseconds()
		if latency < 0 {
			latency = 0
		}
		atomic.AddInt64(&systemMetrics.apiLatencyTotalNs, latency)
		atomic.AddInt64(&systemMetrics.apiLatencyCount, 1)
	}
}

func TrackSSEConnection() func() {
	atomic.AddInt64(&systemMetrics.activeSSEConnections, 1)
	return func() {
		atomic.AddInt64(&systemMetrics.activeSSEConnections, -1)
	}
}

type RuntimeSnapshot struct {
	ServerTimeUTC         string  `json:"server_time_utc"`
	ActiveSSEConnections  int64   `json:"active_sse_connections"`
	AverageAPILatencyMS   float64 `json:"average_api_latency_ms"`
	Goroutines            int     `json:"goroutines"`
	GOMAXPROCS            int     `json:"gomaxprocs"`
	NumCPU                int     `json:"num_cpu"`
	HeapAllocBytes        uint64  `json:"heap_alloc_bytes"`
	HeapSysBytes          uint64  `json:"heap_sys_bytes"`
	HeapIdleBytes         uint64  `json:"heap_idle_bytes"`
	HeapInUseBytes        uint64  `json:"heap_inuse_bytes"`
	StackInUseBytes       uint64  `json:"stack_inuse_bytes"`
	StackSysBytes         uint64  `json:"stack_sys_bytes"`
	TotalAllocBytes       uint64  `json:"total_alloc_bytes"`
	SysBytes              uint64  `json:"sys_bytes"`
	NumGC                 uint32  `json:"num_gc"`
	LastGCTimeUTC         string  `json:"last_gc_time_utc,omitempty"`
	ProcessCPUUsagePct    float64 `json:"process_cpu_usage_pct"`
	ProcessCPUSeconds     float64 `json:"process_cpu_seconds_total"`
	CPUUsageWindowSeconds float64 `json:"cpu_usage_window_seconds"`
}

type QueueSnapshot struct {
	BatchID     int64  `json:"batch_id,omitempty"`
	Status      string `json:"status"`
	BatchType   string `json:"batch_type,omitempty"`
	Pending     int    `json:"pending"`
	Processing  int    `json:"processing"`
	Success     int    `json:"success"`
	Failed      int    `json:"failed"`
	LastError   string `json:"last_error,omitempty"`
	LastUpdated string `json:"last_updated_utc,omitempty"`
}

type QueueHistorySnapshot struct {
	BatchID        int64  `json:"batch_id"`
	BatchType      string `json:"batch_type"`
	Status         string `json:"status"`
	TotalJobs      int    `json:"total_jobs"`
	Success        int    `json:"success"`
	Failed         int    `json:"failed"`
	LastError      string `json:"last_error,omitempty"`
	StartedAtUTC   string `json:"started_at_utc,omitempty"`
	CompletedAtUTC string `json:"completed_at_utc,omitempty"`
}

const queueHistoryMaxRetention = 100

func ReadRuntimeSnapshot(now time.Time) RuntimeSnapshot {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	totalLatencyNs := atomic.LoadInt64(&systemMetrics.apiLatencyTotalNs)
	latencyCount := atomic.LoadInt64(&systemMetrics.apiLatencyCount)
	avgLatencyMS := 0.0
	if latencyCount > 0 {
		avgLatencyMS = float64(totalLatencyNs) / float64(latencyCount) / 1_000_000
	}

	cpuSeconds := readProcessCPUSeconds()
	cpuUsagePct, cpuWindowSeconds := systemMetrics.computeCPUUsagePct(now, cpuSeconds)

	snapshot := RuntimeSnapshot{
		ServerTimeUTC:         now.UTC().Format(time.RFC3339),
		ActiveSSEConnections:  atomic.LoadInt64(&systemMetrics.activeSSEConnections),
		AverageAPILatencyMS:   roundFloat(avgLatencyMS, 2),
		Goroutines:            runtime.NumGoroutine(),
		GOMAXPROCS:            runtime.GOMAXPROCS(0),
		NumCPU:                runtime.NumCPU(),
		HeapAllocBytes:        m.HeapAlloc,
		HeapSysBytes:          m.HeapSys,
		HeapIdleBytes:         m.HeapIdle,
		HeapInUseBytes:        m.HeapInuse,
		StackInUseBytes:       m.StackInuse,
		StackSysBytes:         m.StackSys,
		TotalAllocBytes:       m.TotalAlloc,
		SysBytes:              m.Sys,
		NumGC:                 m.NumGC,
		ProcessCPUUsagePct:    roundFloat(cpuUsagePct, 2),
		ProcessCPUSeconds:     roundFloat(cpuSeconds, 4),
		CPUUsageWindowSeconds: roundFloat(cpuWindowSeconds, 2),
	}
	if m.LastGC > 0 {
		snapshot.LastGCTimeUTC = time.Unix(0, int64(m.LastGC)).UTC().Format(time.RFC3339)
	}
	return snapshot
}

func (m *systemMetricsState) computeCPUUsagePct(now time.Time, cpuSeconds float64) (float64, float64) {
	m.cpuMu.Lock()
	defer m.cpuMu.Unlock()

	if m.lastCPUWallTime.IsZero() {
		m.lastCPUSeconds = cpuSeconds
		m.lastCPUWallTime = now
		return 0, 0
	}

	wallDelta := now.Sub(m.lastCPUWallTime).Seconds()
	cpuDelta := cpuSeconds - m.lastCPUSeconds
	m.lastCPUSeconds = cpuSeconds
	m.lastCPUWallTime = now

	if wallDelta <= 0 || cpuDelta < 0 {
		return 0, 0
	}
	return (cpuDelta / wallDelta) * 100, wallDelta
}

func readProcessCPUSeconds() float64 {
	samples := []metrics.Sample{{Name: "/cpu/classes/total:cpu-seconds"}}
	metrics.Read(samples)
	if len(samples) == 0 {
		return 0
	}
	return samples[0].Value.Float64()
}

func roundFloat(v float64, precision int) float64 {
	if precision < 0 {
		return v
	}
	f := math.Pow10(precision)
	return math.Round(v*f) / f
}

func QueueBeginBatch(batchType string, totalJobs int) {
	systemMetrics.queueMu.Lock()
	defer systemMetrics.queueMu.Unlock()
	if totalJobs < 0 {
		totalJobs = 0
	}
	systemMetrics.queueBatchSeq++
	now := time.Now().UTC()
	batchID := systemMetrics.queueBatchSeq
	systemMetrics.queue = queueState{
		BatchID:     batchID,
		Status:      "running",
		BatchType:   batchType,
		Pending:     totalJobs,
		Processing:  0,
		Success:     0,
		Failed:      0,
		LastError:   "",
		LastUpdated: now,
	}
	systemMetrics.queueHistory = append([]queueHistoryItem{{
		BatchID:   batchID,
		BatchType: batchType,
		Status:    "running",
		TotalJobs: totalJobs,
		StartedAt: now,
	}}, systemMetrics.queueHistory...)
	if len(systemMetrics.queueHistory) > queueHistoryMaxRetention {
		systemMetrics.queueHistory = systemMetrics.queueHistory[:queueHistoryMaxRetention]
	}
}

func QueueStartJob() {
	systemMetrics.queueMu.Lock()
	defer systemMetrics.queueMu.Unlock()
	if systemMetrics.queue.Pending > 0 {
		systemMetrics.queue.Pending--
	}
	systemMetrics.queue.Processing++
	systemMetrics.queue.LastUpdated = time.Now().UTC()
}

func QueueFinishJob(err error) {
	systemMetrics.queueMu.Lock()
	defer systemMetrics.queueMu.Unlock()
	if systemMetrics.queue.Processing > 0 {
		systemMetrics.queue.Processing--
	}
	if err != nil {
		systemMetrics.queue.Failed++
		systemMetrics.queue.LastError = err.Error()
	} else {
		systemMetrics.queue.Success++
	}
	systemMetrics.queue.LastUpdated = time.Now().UTC()
	for i := range systemMetrics.queueHistory {
		if systemMetrics.queueHistory[i].BatchID != systemMetrics.queue.BatchID {
			continue
		}
		systemMetrics.queueHistory[i].Success = systemMetrics.queue.Success
		systemMetrics.queueHistory[i].Failed = systemMetrics.queue.Failed
		systemMetrics.queueHistory[i].LastError = systemMetrics.queue.LastError
		break
	}
}

func QueueCompleteBatch() {
	systemMetrics.queueMu.Lock()
	defer systemMetrics.queueMu.Unlock()
	now := time.Now().UTC()
	if systemMetrics.queue.Processing == 0 && systemMetrics.queue.Pending == 0 {
		systemMetrics.queue.Status = "idle"
	}
	systemMetrics.queue.LastUpdated = now
	for i := range systemMetrics.queueHistory {
		if systemMetrics.queueHistory[i].BatchID != systemMetrics.queue.BatchID {
			continue
		}
		systemMetrics.queueHistory[i].Success = systemMetrics.queue.Success
		systemMetrics.queueHistory[i].Failed = systemMetrics.queue.Failed
		systemMetrics.queueHistory[i].LastError = systemMetrics.queue.LastError
		systemMetrics.queueHistory[i].Status = systemMetrics.queue.Status
		systemMetrics.queueHistory[i].CompletedAt = now
		break
	}
}

func ReadQueueSnapshot() QueueSnapshot {
	systemMetrics.queueMu.Lock()
	defer systemMetrics.queueMu.Unlock()
	out := QueueSnapshot{
		BatchID:    systemMetrics.queue.BatchID,
		Status:     systemMetrics.queue.Status,
		BatchType:  systemMetrics.queue.BatchType,
		Pending:    systemMetrics.queue.Pending,
		Processing: systemMetrics.queue.Processing,
		Success:    systemMetrics.queue.Success,
		Failed:     systemMetrics.queue.Failed,
		LastError:  systemMetrics.queue.LastError,
	}
	if out.Status == "" {
		out.Status = "idle"
	}
	if !systemMetrics.queue.LastUpdated.IsZero() {
		out.LastUpdated = systemMetrics.queue.LastUpdated.UTC().Format(time.RFC3339)
	}
	return out
}

func ReadQueueHistorySnapshot(limit int) []QueueHistorySnapshot {
	systemMetrics.queueMu.Lock()
	defer systemMetrics.queueMu.Unlock()
	if limit <= 0 {
		limit = 20
	}
	if limit > queueHistoryMaxRetention {
		limit = queueHistoryMaxRetention
	}
	history := systemMetrics.queueHistory
	if len(history) > limit {
		history = history[:limit]
	}
	out := make([]QueueHistorySnapshot, 0, len(history))
	for _, item := range history {
		row := QueueHistorySnapshot{
			BatchID:   item.BatchID,
			BatchType: item.BatchType,
			Status:    item.Status,
			TotalJobs: item.TotalJobs,
			Success:   item.Success,
			Failed:    item.Failed,
			LastError: item.LastError,
		}
		if !item.StartedAt.IsZero() {
			row.StartedAtUTC = item.StartedAt.UTC().Format(time.RFC3339)
		}
		if !item.CompletedAt.IsZero() {
			row.CompletedAtUTC = item.CompletedAt.UTC().Format(time.RFC3339)
		}
		out = append(out, row)
	}
	return out
}

func ClearQueueHistory() {
	systemMetrics.queueMu.Lock()
	defer systemMetrics.queueMu.Unlock()
	systemMetrics.queueHistory = nil
}
