// Package service implements the business logic between handlers and the database
package service

import (
	"github.com/jacoboneill/SecureBin/internal/db"
)

type Service struct {
	queries db.Querier
}

func NewService(queries db.Querier) *Service {
	return &Service{queries: queries}
}
