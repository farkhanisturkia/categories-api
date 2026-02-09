package repositories

import (
	"fmt"
	"time"
	"database/sql"
	"categories-api/models"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (repo *TransactionRepository) CreateTransaction(items []models.CheckoutItem) (*models.Transaction, error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	totalAmount := 0
	details := make([]models.TransactionDetail, 0)

	for _, item := range items {
		var productPrice, stock int
		var productName string

		err := tx.QueryRow("SELECT name, price, stock FROM products WHERE id = $1", item.ProductID).Scan(&productName, &productPrice, &stock)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product id %d not found", item.ProductID)
		}
		if err != nil {
			return nil, err
		}
  
		subtotal := productPrice * item.Quantity
		totalAmount += subtotal

		_, err = tx.Exec("UPDATE products SET stock = stock - $1 WHERE id = $2", item.Quantity, item.ProductID)
		if err != nil {
			return nil, err
		}

		details = append(details, models.TransactionDetail{
			ProductID:   item.ProductID,
			ProductName: productName,
			Quantity:    item.Quantity,
			Subtotal:    subtotal,
		})
	}

	var transactionID int
	err = tx.QueryRow("INSERT INTO transactions (total_amount) VALUES ($1) RETURNING id", totalAmount).Scan(&transactionID)
	if err != nil {
		return nil, err
	}

	for i := range details {
		details[i].TransactionID = transactionID
		_, err = tx.Exec("INSERT INTO transaction_details (transaction_id, product_id, quantity, subtotal) VALUES ($1, $2, $3, $4)",
			transactionID, details[i].ProductID, details[i].Quantity, details[i].Subtotal)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.Transaction{
		ID:          transactionID,
		TotalAmount: totalAmount,
		Details:     details,
	}, nil
}

func (repo *TransactionRepository) GetReport(start, end time.Time) (models.Report, error) {
    var report models.Report

    err := repo.db.QueryRow(`
        SELECT 
            COALESCE(SUM(t.total_amount), 0) AS total_revenue,
            COUNT(DISTINCT t.id) AS transaction_count,
            COALESCE(SUM(td.quantity), 0) AS total_items_sold
        FROM transactions t
        LEFT JOIN transaction_details td ON t.id = td.transaction_id
        WHERE t.created_at >= $1 AND t.created_at < $2
    `, start, end).Scan(
        &report.TotalRevenue,
        &report.TransactionCount,
        &report.TotalItemsSold,
    )
    if err != nil {
        return models.Report{}, err
    }

    rows, err := repo.db.Query(`
        SELECT 
            t.id,
            t.total_amount,
            t.created_at,
            COUNT(td.id) AS item_count
        FROM transactions t
        LEFT JOIN transaction_details td ON t.id = td.transaction_id
        WHERE t.created_at >= $1 AND t.created_at < $2
        GROUP BY t.id, t.total_amount, t.created_at
        ORDER BY t.created_at DESC
        LIMIT 100
    `, start, end)
    if err != nil {
        return models.Report{}, err
    }
    defer rows.Close()

    report.Transactions = []models.ReportTransactionSummary{}
    for rows.Next() {
        var tx models.ReportTransactionSummary
        err := rows.Scan(
            &tx.ID,
            &tx.TotalAmount,
            &tx.CreatedAt,
            &tx.ItemCount,
        )
        if err != nil {
            return models.Report{}, err
        }
        report.Transactions = append(report.Transactions, tx)
    }

    if err = rows.Err(); err != nil {
        return models.Report{}, err
    }

    return report, nil
}