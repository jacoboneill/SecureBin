package service

import (
	"github.com/jacoboneill/SecureBin/internal/db"
)

type Service struct {
	queries *db.Queries
}

func NewService(queries *db.Queries) *Service {
	return &Service{queries: queries}
}
