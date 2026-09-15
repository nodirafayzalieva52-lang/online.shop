package models

import "time"

type Role string

const (
	RoleUnassigned Role = "unassigned"
	RoleCustomer   Role = "customer"
	RoleClient     Role = "client"
	RoleSeller     Role = "seller"
	RoleAdmin      Role = "admin"
)

type User struct {
	ID           int64     `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}