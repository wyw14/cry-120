package weather

import (
	"sync"
	"time"
)

type Reading struct {
	WindMetersPerSecond float64   `json:"wind_meters_per_second"`
	LightningDistanceKM float64   `json:"lightning_distance_km"`
	ValidUntil          time.Time `json:"valid_until"`
	ObservedAt          time.Time `json:"observed_at"`
}

type Service struct {
	mu      sync.RWMutex
	holds   *HoldService
	reading Reading
}

func NewService(holds *HoldService) *Service {
	return &Service{holds: holds}
}

func (s *Service) Update(reading Reading, now time.Time) Reading {
	s.mu.Lock()
	reading.ObservedAt = now.UTC()
	s.reading = reading
	s.mu.Unlock()
	if reading.LightningDistanceKM > 0 && reading.LightningDistanceKM < 20 {
		s.holds.Raise("区域不安全", now)
	} else {
		s.holds.Clear("区域不安全")
	}
	if reading.WindMetersPerSecond > 18 {
		s.holds.Raise("high wind", now)
	} else {
		s.holds.Clear("high wind")
	}
	return reading
}

func (s *Service) Current(now time.Time) map[string]any {
	s.Expire(now)
	s.mu.RLock()
	reading := s.reading
	s.mu.RUnlock()
	return map[string]any{"reading": reading, "valid": !reading.ValidUntil.IsZero() && now.Before(reading.ValidUntil), "holds": s.holds.Active()}
}

func (s *Service) Expire(now time.Time) bool {
	s.mu.RLock()
	expired := !s.reading.ValidUntil.IsZero() && !now.Before(s.reading.ValidUntil)
	s.mu.RUnlock()
	if expired {
		s.holds.Raise("weather stale", now)
	}
	return expired
}
