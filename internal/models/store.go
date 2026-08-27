package models

import "time"

type Store struct {
	ID          int64     `json:"id"`
	SellerID    int64     `json:"seller_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}
