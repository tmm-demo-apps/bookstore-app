package models

import (
	"golang.org/x/crypto/bcrypt"
)

type Product struct {
	ID          int
	Name        string
	Description string
	Price       float64
}
