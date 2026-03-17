package service

import (
	"context"
	"time"
)

// ExportCSV возвращает данные в формате CSV как []byte
func (s *Service) ExportCSV(ctx context.Context, from, to time.Time) ([]byte, error) {

}
