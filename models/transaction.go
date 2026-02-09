package models

import "time"

type Transaction struct {
	ID          int                 `json:"id"`
	TotalAmount int                 `json:"total_amount"`
	CreatedAt   time.Time           `json:"created_at"`
	Details     []TransactionDetail `json:"details"`
}

type TransactionDetail struct {
	ID            int    `json:"id"`
	TransactionID int    `json:"transaction_id"`
	ProductID     int    `json:"product_id"`
	ProductName   string `json:"product_name,omitempty"`
	Quantity      int    `json:"quantity"`
	Subtotal      int    `json:"subtotal"`
}

type CheckoutItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type CheckoutRequest struct {
	Items   []CheckoutItem `json:"items"`
}

type TodayReport struct {
    Date             string                      `json:"date"` // "2026-02-09"
    TotalRevenue     int                         `json:"total_revenue"`
    TransactionCount int                         `json:"transaction_count"`
    TotalItemsSold   int                         `json:"total_items_sold"`
    Transactions     []TodayTransactionSummary   `json:"transactions"`
}

type TodayTransactionSummary struct {
    ID          int       `json:"id"`
    TotalAmount int       `json:"total_amount"`
    CreatedAt   time.Time `json:"created_at"`
    ItemCount   int       `json:"item_count"`
}