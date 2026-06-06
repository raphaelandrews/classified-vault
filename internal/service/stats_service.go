package service

import (
	"classified-vault/internal/repository"
)

type StatsService struct {
	repo *repository.StatsRepository
}

func NewStatsService(repo *repository.StatsRepository) *StatsService {
	return &StatsService{repo: repo}
}

func (s *StatsService) GetStats() (*repository.StatsResponse, error) {
	return s.repo.GetStats()
}
