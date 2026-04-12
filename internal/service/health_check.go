package service

import (
	"context"
)

// Статусы здоровья.
const (
	StatusHealthy   = "healthy"
	StatusUnhealthy = "unhealthy"
)

// HealthCheckResult представляет результат проверки здоровья.
type HealthCheckResult struct {
	Status     string
	Components struct {
		Database string
		Storage  string
		Kafka    string
	}
}

// HealthCheck проверяет здоровье всех компонентов.
func (s *Service) HealthCheck(ctx context.Context) *HealthCheckResult {
	result := &HealthCheckResult{
		Status: StatusHealthy,
	}

	// Проверяем database
	if err := s.repo.Ping(ctx); err != nil {
		result.Components.Database = StatusUnhealthy
		result.Status = StatusUnhealthy
	} else {
		result.Components.Database = StatusHealthy
	}

	// Проверяем storage
	if err := s.storage.Ping(ctx); err != nil {
		result.Components.Storage = StatusUnhealthy
		result.Status = StatusUnhealthy
	} else {
		result.Components.Storage = StatusHealthy
	}

	// Проверяем Kafka
	if s.producer != nil {
		if err := s.producer.Ping(ctx); err != nil {
			result.Components.Kafka = StatusUnhealthy
			result.Status = StatusUnhealthy
		} else {
			result.Components.Kafka = StatusHealthy
		}
	} else {
		result.Components.Kafka = StatusUnhealthy
		result.Status = StatusUnhealthy
	}

	return result
}
