package document

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// MemoryMonitor monitors memory usage during document processing
type MemoryMonitor struct {
	// Configuration
	warningThreshold  uint64 // Memory threshold for warning (bytes)
	criticalThreshold uint64 // Memory threshold for critical alert (bytes)
	checkInterval     time.Duration
	
	// State
	isRunning    bool
	stopChan     chan struct{}
	alertHandler func(level string, usage uint64, limit uint64)
	
	// Statistics
	peakUsage    uint64
	avgUsage     uint64
	sampleCount  int64
	totalSamples uint64
	
	mu sync.RWMutex
}

// MemoryStats contains memory usage statistics
type MemoryStats struct {
	CurrentUsage uint64    `json:"current_usage"`
	PeakUsage    uint64    `json:"peak_usage"`
	AvgUsage     uint64    `json:"avg_usage"`
	SystemTotal  uint64    `json:"system_total"`
	SystemFree   uint64    `json:"system_free"`
	HeapAlloc    uint64    `json:"heap_alloc"`
	HeapSys      uint64    `json:"heap_sys"`
	NumGC        uint32    `json:"num_gc"`
	Timestamp    time.Time `json:"timestamp"`
}

// NewMemoryMonitor creates a new memory monitor
func NewMemoryMonitor() *MemoryMonitor {
	return &MemoryMonitor{
		warningThreshold:  500 * 1024 * 1024,  // 500MB default warning
		criticalThreshold: 1024 * 1024 * 1024, // 1GB default critical
		checkInterval:     1 * time.Second,
		stopChan:         make(chan struct{}),
		alertHandler: func(level string, usage uint64, limit uint64) {
			// Default alert handler - can be overridden
			fmt.Printf("[%s] Memory usage: %s / %s (%.1f%%)\n",
				level,
				FormatBytes(usage),
				FormatBytes(limit),
				float64(usage)/float64(limit)*100)
		},
	}
}

// SetThresholds sets memory warning and critical thresholds
func (m *MemoryMonitor) SetThresholds(warning, critical uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.warningThreshold = warning
	m.criticalThreshold = critical
}

// SetAlertHandler sets a custom alert handler function
func (m *MemoryMonitor) SetAlertHandler(handler func(level string, usage uint64, limit uint64)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alertHandler = handler
}

// Start starts monitoring memory usage
func (m *MemoryMonitor) Start() {
	m.mu.Lock()
	if m.isRunning {
		m.mu.Unlock()
		return
	}
	m.isRunning = true
	m.mu.Unlock()
	
	go m.monitorLoop()
}

// Stop stops monitoring memory usage
func (m *MemoryMonitor) Stop() {
	m.mu.Lock()
	if !m.isRunning {
		m.mu.Unlock()
		return
	}
	m.isRunning = false
	m.mu.Unlock()
	
	close(m.stopChan)
}

// monitorLoop is the main monitoring loop
func (m *MemoryMonitor) monitorLoop() {
	ticker := time.NewTicker(m.checkInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			m.checkMemory()
		case <-m.stopChan:
			return
		}
	}
}

// checkMemory checks current memory usage and triggers alerts if needed
func (m *MemoryMonitor) checkMemory() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	currentUsage := memStats.Alloc
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Update statistics
	if currentUsage > m.peakUsage {
		m.peakUsage = currentUsage
	}
	
	m.sampleCount++
	m.totalSamples += currentUsage
	m.avgUsage = m.totalSamples / uint64(m.sampleCount)
	
	// Check thresholds
	if currentUsage > m.criticalThreshold {
		if m.alertHandler != nil {
			m.alertHandler("CRITICAL", currentUsage, m.criticalThreshold)
		}
		// Force garbage collection when critical
		runtime.GC()
	} else if currentUsage > m.warningThreshold {
		if m.alertHandler != nil {
			m.alertHandler("WARNING", currentUsage, m.warningThreshold)
		}
	}
}

// GetStats returns current memory statistics
func (m *MemoryMonitor) GetStats() *MemoryStats {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return &MemoryStats{
		CurrentUsage: memStats.Alloc,
		PeakUsage:    m.peakUsage,
		AvgUsage:     m.avgUsage,
		SystemTotal:  memStats.Sys,
		SystemFree:   memStats.Frees,
		HeapAlloc:    memStats.HeapAlloc,
		HeapSys:      memStats.HeapSys,
		NumGC:        memStats.NumGC,
		Timestamp:    time.Now(),
	}
}

// FormatBytes formats bytes into human-readable string
func FormatBytes(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// GetSystemMemoryInfo returns system memory information
func GetSystemMemoryInfo() (total, available uint64) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	// Note: This is a simplified version
	// In production, you might want to use system-specific calls
	// to get actual system memory (not just Go runtime memory)
	return memStats.Sys, memStats.Sys - memStats.Alloc
}

// ShouldProcessInMemory determines if a file should be processed in memory
// based on its size and available memory
func ShouldProcessInMemory(fileSize int64) bool {
	_, available := GetSystemMemoryInfo()
	
	// Conservative estimate: file may expand 10x when processed
	estimatedMemoryNeeded := uint64(fileSize * 10)
	
	// Use in-memory processing if we have enough available memory
	// with a safety margin of 20%
	return estimatedMemoryNeeded < (available * 80 / 100)
}

// MemoryPool manages a pool of reusable byte slices
type MemoryPool struct {
	pool *sync.Pool
	size int
}

// NewMemoryPool creates a new memory pool with specified chunk size
func NewMemoryPool(chunkSize int) *MemoryPool {
	return &MemoryPool{
		size: chunkSize,
		pool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, chunkSize)
			},
		},
	}
}

// Get retrieves a byte slice from the pool
func (p *MemoryPool) Get() []byte {
	return p.pool.Get().([]byte)
}

// Put returns a byte slice to the pool
func (p *MemoryPool) Put(b []byte) {
	if cap(b) >= p.size {
		// Only return to pool if it's the right size
		p.pool.Put(b[:p.size])
	}
}

// Reset clears the byte slice and returns it to the pool
func (p *MemoryPool) Reset(b []byte) {
	// Clear the slice for security
	for i := range b {
		b[i] = 0
	}
	p.Put(b)
}