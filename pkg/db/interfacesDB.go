package db

// Storage - интерфейсы БД
type Storage interface {
	StorageUsers
	StorageItems
	StorageItemsHistory
}

// StorageUsers определяет методы работы с таблицей users
type StorageUsers interface {
}

// StorageItems определяет методы работы с таблицей items
type StorageItems interface {
}

// StorageItemsHistory определяет методы работы с таблицей item_history
type StorageItemsHistory interface {
}
