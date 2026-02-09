package services

import (
	"time"
	"categories-api/models"
	"categories-api/repositories"
)

type TransactionService struct {
	repo *repositories.TransactionRepository
}

func NewTransactionService(repo *repositories.TransactionRepository) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) Checkout(items []models.CheckoutItem) (*models.Transaction, error) {
		return s.repo.CreateTransaction(items)
}

func (s *TransactionService) GetTodayReport() (models.TodayReport, error) {
    report, err := s.repo.GetTodayReport()
    if err != nil {
        return models.TodayReport{}, err
    }

    report.Date = time.Now().Format("2006-01-02")

    return report, nil
}