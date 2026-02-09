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

func (repo *TransactionRepository) GetTodayReport() (models.TodayReport, error) {
    today := time.Now().Truncate(24 * time.Hour) // mulai dari 00:00 hari ini
    tomorrow := today.Add(24 * time.Hour)

    var report models.TodayReport

    // 1. Total amount & jumlah transaksi
    err := repo.db.QueryRow(`
        SELECT 
            COALESCE(SUM(total_amount), 0) AS total_revenue,
            COUNT(*) AS transaction_count
        FROM transactions
        WHERE created_at >= $1 AND created_at < $2
    `, today, tomorrow).Scan(
        &report.TotalRevenue,
        &report.TransactionCount,
    )
    if err != nil {
        return models.TodayReport{}, err
    }

    // 2. Total item terjual
    err = repo.db.QueryRow(`
        SELECT COALESCE(SUM(quantity), 0) AS total_items_sold
        FROM transaction_details td
        JOIN transactions t ON td.transaction_id = t.id
        WHERE t.created_at >= $1 AND t.created_at < $2
    `, today, tomorrow).Scan(&report.TotalItemsSold)
    if err != nil {
        return models.TodayReport{}, err
    }

    // 3. Daftar transaksi hari ini (ringkas)
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
        LIMIT 50
    `, today, tomorrow)
    if err != nil {
        return models.TodayReport{}, err
    }
    defer rows.Close()

    report.Transactions = make([]models.TodayTransactionSummary, 0)
    for rows.Next() {
        var tx models.TodayTransactionSummary
        err := rows.Scan(
            &tx.ID,
            &tx.TotalAmount,
            &tx.CreatedAt,
            &tx.ItemCount,
        )
        if err != nil {
            return models.TodayReport{}, err
        }
        report.Transactions = append(report.Transactions, tx)
    }

    if err = rows.Err(); err != nil {
        return models.TodayReport{}, err
    }

    return report, nil
}