package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/IPampurin/WarehouseControlAntiPattern/pkg/domain"
)

// getCurrentUserID извлекает id пользователя из контекста (устанавливается middleware)
func getCurrentUserID(ctx context.Context) int {

	if userID, ok := ctx.Value("userID").(int); ok {
		return userID
	}

	return 0 // 0 означает отсутствие пользователя (не должно случаться после middleware)
}

// getCurrentUserRole извлекает роль пользователя из контекста
func getCurrentUserRole(ctx context.Context) string {

	if role, ok := ctx.Value("role").(string); ok {
		return role
	}

	return "" // пустая строка – отсутствие роли
}

// checkRole проверяет, имеет ли текущий пользователь одну из разрешённых ролей
func (s *Service) checkRole(ctx context.Context, allowedRoles ...string) error {

	role := getCurrentUserRole(ctx)
	for _, r := range allowedRoles {
		if role == r {
			return nil
		}
	}

	return domain.ErrForbidden
}

// товары

// CreateItem создаёт запись о новом товаре
func (s *Service) CreateItem(ctx context.Context, item *domain.Item) (*domain.Item, error) {

	// проверка прав
	if err := s.checkRole(ctx, "admin", "manager"); err != nil {
		return nil, err
	}

	// валидация
	if item == nil || item.Name == "" || item.Quantity < 0 || item.Price < 0 {
		return nil, domain.ErrInvalidInput
	}

	// сохранение
	if err := s.storage.CreateItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

// GetItemByID возвращает товар по его идентификатору (если товар не найден, возвращает nil, nil)
func (s *Service) GetItemByID(ctx context.Context, id int) (*domain.Item, error) {

	// просмотр доступен всем
	item, err := s.storage.GetItemByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if item == nil {
		return nil, domain.ErrNotFound
	}

	return item, nil
}

// GetAllItems возвращает список всех товаров
func (s *Service) GetAllItems(ctx context.Context) ([]*domain.Item, error) {

	// просмотр доступен всем
	items, err := s.storage.GetAllItems(ctx)
	if err != nil {
		return nil, err
	}

	return items, nil
}

// UpdateItem обновляет существующий товар
func (s *Service) UpdateItem(ctx context.Context, item *domain.Item) (*domain.Item, error) {

	// проверка прав
	if err := s.checkRole(ctx, "admin", "manager"); err != nil {
		return nil, err
	}

	// валидация
	if item == nil || item.Name == "" || item.Quantity < 0 || item.Price < 0 {
		return nil, domain.ErrInvalidInput
	}

	// проверяем, существует ли товар
	existing, err := s.storage.GetItemByID(ctx, item.ID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, domain.ErrNotFound
	}

	if err := s.storage.UpdateItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

// DeleteItem удаляет товар по id
func (s *Service) DeleteItem(ctx context.Context, id int) error {

	// проверка прав
	if err := s.checkRole(ctx, "admin"); err != nil {
		return err
	}

	existing, err := s.storage.GetItemByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return domain.ErrNotFound
	}

	return s.storage.DeleteItem(ctx, id)
}

// история

// GetItemHistory возвращает историю изменений для конкретного товара (с фильтрацией и пагинацией)
func (s *Service) GetItemHistory(ctx context.Context, itemID int, filter *domain.HistoryFilter) ([]*domain.ItemHistory, error) {

	// проверка прав
	if err := s.checkRole(ctx, "admin"); err != nil {
		return nil, err
	}
	if filter == nil {
		filter = &domain.HistoryFilter{}
	}

	filter.ItemID = itemID // принудительно устанавливаем itemID

	history, err := s.storage.GetHistory(ctx, filter)
	if err != nil {
		return nil, err
	}

	return history, nil
}

// GetGlobalHistory возвращает общую историю изменений всех товаров (с фильтрацией и пагинацией)
func (s *Service) GetGlobalHistory(ctx context.Context, filter *domain.HistoryFilter) ([]*domain.ItemHistory, error) {

	// проверка прав
	if err := s.checkRole(ctx, "admin"); err != nil {
		return nil, err
	}

	if filter == nil {
		filter = &domain.HistoryFilter{}
	}

	history, err := s.storage.GetHistory(ctx, filter)
	if err != nil {
		return nil, err
	}

	return history, nil
}

// GetHistoryByID возвращает одну запись истории по её id
func (s *Service) GetHistoryByID(ctx context.Context, id int) (*domain.ItemHistory, error) {

	// проверка прав
	if err := s.checkRole(ctx, "admin"); err != nil {
		return nil, err
	}

	history, err := s.storage.GetHistoryByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if history == nil {
		return nil, domain.ErrNotFound
	}

	return history, nil
}

// CompareHistoryVersions сравнивает две записи истории и возвращает структуру с различиями
func (s *Service) CompareHistoryVersions(ctx context.Context, id1, id2 int) (*domain.DiffResponse, error) {

	// прверка прав
	if err := s.checkRole(ctx, "admin"); err != nil {
		return nil, err
	}

	h1, err := s.storage.GetHistoryByID(ctx, id1)
	if err != nil {
		return nil, err
	}

	if h1 == nil {
		return nil, domain.ErrNotFound
	}

	h2, err := s.storage.GetHistoryByID(ctx, id2)
	if err != nil {
		return nil, err
	}

	if h2 == nil {
		return nil, domain.ErrNotFound
	}

	// парсим JSON данные из записей
	data1 := make(map[string]interface{})
	data2 := make(map[string]interface{})

	if len(h1.NewData) > 0 { // используем NewData как состояние после действия
		if err := json.Unmarshal(h1.NewData, &data1); err != nil {
			return nil, fmt.Errorf("не удалось разобрать данные первой версии: %w", err)
		}
	} else if len(h1.OldData) > 0 {
		if err := json.Unmarshal(h1.OldData, &data1); err != nil {
			return nil, fmt.Errorf("не удалось разобрать данные первой версии: %w", err)
		}
	}

	if len(h2.NewData) > 0 {
		if err := json.Unmarshal(h2.NewData, &data2); err != nil {
			return nil, fmt.Errorf("не удалось разобрать данные второй версии: %w", err)
		}
	} else if len(h2.OldData) > 0 {
		if err := json.Unmarshal(h2.OldData, &data2); err != nil {
			return nil, fmt.Errorf("не удалось разобрать данные второй версии: %w", err)
		}
	}

	// сравниваем поля name, quantity, price
	changes := []domain.FieldChange{}
	fields := []string{"name", "quantity", "price"}

	for _, field := range fields {
		oldVal, oldOk := data1[field]
		newVal, newOk := data2[field]
		if oldOk && newOk {
			// если значения различаются
			if fmt.Sprint(oldVal) != fmt.Sprint(newVal) {
				changes = append(changes, domain.FieldChange{
					Field:    field,
					OldValue: oldVal,
					NewValue: newVal,
				})
			}
		} else if oldOk && !newOk {
			// было, а стало нет (удаление)
			changes = append(changes, domain.FieldChange{
				Field:    field,
				OldValue: oldVal,
				NewValue: nil,
			})
		} else if !oldOk && newOk {
			// не было, появилось (создание)
			changes = append(changes, domain.FieldChange{
				Field:    field,
				OldValue: nil,
				NewValue: newVal,
			})
		}
	}

	return &domain.DiffResponse{
		FromID:  id1,
		ToID:    id2,
		Changes: changes,
	}, nil
}

// экспорт

// ExportItemsCSV генерирует CSV-файл со всеми товарами
func (s *Service) ExportItemsCSV(ctx context.Context) ([]byte, error) {

	// доступно всем
	items, err := s.storage.GetAllItems(ctx)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	// заголовки
	if err := writer.Write([]string{"ID", "Имя", "Количество", "Цена", "Создано", "Обновлено"}); err != nil {
		return nil, err
	}

	for _, item := range items {
		record := []string{
			strconv.Itoa(item.ID),
			item.Name,
			strconv.Itoa(item.Quantity),
			fmt.Sprintf("%.2f", item.Price),
			item.CreatedAt.Format(time.RFC3339),
			item.UpdatedAt.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ExportHistoryCSV генерирует CSV-файл с записями истории, отфильтрованными по заданным критериям
func (s *Service) ExportHistoryCSV(ctx context.Context, filter *domain.HistoryFilter) ([]byte, error) {

	// проверка прав
	if err := s.checkRole(ctx, "admin"); err != nil {
		return nil, err
	}
	if filter == nil {
		filter = &domain.HistoryFilter{}
	}

	// для экспорта не используем пагинацию, поэтому устанавливаем Limit=0 (вернёт всё)
	filter.Limit = 0
	filter.Offset = 0

	history, err := s.storage.GetHistory(ctx, filter)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	writer := csv.NewWriter(buf)

	// заголовки
	if err := writer.Write([]string{"ID", "ItemID", "Действие", "Было", "Стало", "Кто менял", "Когда менял"}); err != nil {
		return nil, err
	}

	for _, h := range history {
		changedBy := ""
		if h.ChangedBy != nil {
			changedBy = strconv.Itoa(*h.ChangedBy)
		}
		record := []string{
			strconv.Itoa(h.ID),
			strconv.Itoa(h.ItemID),
			h.Action,
			string(h.OldData),
			string(h.NewData),
			changedBy,
			h.ChangedAt.Format(time.RFC3339),
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
