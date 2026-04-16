package backup

import (
	"sync"
	"time"
)

type speedSample struct {
	bytes    int64
	duration time.Duration
}

// SpeedTracker 滑動平均速度追蹤器
type SpeedTracker struct {
	samples    []speedSample
	windowSize int
	mu         sync.Mutex
}

// NewSpeedTracker 建立速度追蹤器（窗口 20 筆）
func NewSpeedTracker() *SpeedTracker {
	return &SpeedTracker{windowSize: 20}
}

// Add 加入新樣本
func (s *SpeedTracker) Add(bytes int64, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.samples = append(s.samples, speedSample{bytes, duration})
	if len(s.samples) > s.windowSize {
		s.samples = s.samples[1:]
	}
}

// Average 計算滑動平均速度（bytes/sec）
func (s *SpeedTracker) Average() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.samples) == 0 {
		return 0
	}
	var totalBytes int64
	var totalDuration time.Duration
	for _, sample := range s.samples {
		totalBytes += sample.bytes
		totalDuration += sample.duration
	}
	if totalDuration == 0 {
		return 0
	}
	return float64(totalBytes) / totalDuration.Seconds()
}

// ETA 估算剩餘時間
func (s *SpeedTracker) ETA(remainingBytes int64) time.Duration {
	avg := s.Average()
	if avg <= 0 {
		return 0
	}
	return time.Duration(float64(remainingBytes) / avg * float64(time.Second))
}
