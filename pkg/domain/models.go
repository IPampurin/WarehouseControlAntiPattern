package domain

import (
	"time"
)

// User представляет пользователя системы
type User struct {
	ID        int       // идентификатор пользователя
	Name      string    // имя пользователя
	Email     string    // уникальный email пользователя
	Role      string    // роль: admin, manager, viewer
	CreatedAt time.Time // дата регистрации пользователя
}

// Item представляет товар на складе
type Item struct {
	ID        int       // идентификатор товара
	Name      string    // название товара
	Quantity  int       // количество (неотрицательное)
	Price     float64   // цена товара
	CreatedAt time.Time // дата создания записи о товаре
	UpdatedAt time.Time // дата обновления данных о товаре
}

// ItemHistory представляет запись в истории изменений товара
type ItemHistory struct {
	ID        int       // идентификатор записи истории
	ItemID    int       // идентификатор товара
	Action    string    // действие (INSERT, UPDATE, DELETE)
	OldData   []byte    // предыдущее состояние товара в формате JSON (может быть nil)
	NewData   []byte    // новое состояние товара в формате JSON (может быть nil)
	ChangedBy *int      // идентификатор пользователя, совершившего изменение (nil если пользователь удалён)
	ChangedAt time.Time // время изменений
}

// HistoryFilter содержит параметры фильтрации при запросе истории
// (используется для выборки записей)
type HistoryFilter struct {
	ItemID   int       // если не 0, то только для указанного товара
	FromDate time.Time // если не zero, то только записи после этой даты
	ToDate   time.Time // если не zero, то только записи до этой даты
	UserID   int       // если не 0, то только изменения, сделанные этим пользователем
	Action   string    // если не пустая, то только указанное действие (INSERT, UPDATE, DELETE)
	Limit    int       // максимальное количество записей (0 - использовать значение по умолчанию)
	Offset   int       // смещение для пагинации
}
