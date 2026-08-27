package models

import "time"

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type OrderItem struct {
	ID        int64    `json:"id"`
	OrderID   int64    `json:"order_id"`
	ProductID int64    `json:"product_id"`
	StoreID   int64    `json:"store_id"` 
	Quantity  int      `json:"quantity"`
	Price     float64  `json:"price"` 
	Product   *Product `json:"product,omitempty"`
}

type Order struct {
	ID         int64       `json:"id"`
	CustomerID int64       `json:"customer_id"`
	StoreID    int64       `json:"store_id"`
	TotalPrice float64     `json:"total_price"`
	Status     OrderStatus `json:"status"`
	Items      []OrderItem `json:"items"`
	CreatedAt  time.Time   `json:"created_at"`
}

