package service

import (
	"time"

	"github.com/melihgurlek/backend-path/internal/domain"
)

type BalanceServiceImpl struct {
	repo domain.BalanceRepository
}

func NewBalanceService(repo domain.BalanceRepository) domain.BalanceService {
	return &BalanceServiceImpl{repo: repo}
}

func (s *BalanceServiceImpl) GetCurrentBalance(userID int) (*domain.Balance, error) {
	return s.repo.GetByUserID(userID)
}

func (s *BalanceServiceImpl) GetHistoricalBalance(userID int, limit int) ([]*domain.Balance, error) {
	return s.repo.GetHistoricalBalance(userID, limit)
}

func (s *BalanceServiceImpl) GetBalanceAtTime(userID int, t time.Time) (*domain.Balance, error) {
	return s.repo.GetBalanceAtTime(userID, t)
}
