package service

// ServiceMethods определяет методы бизнес-логики
type ServiceMethods interface {

	// CSVExporter возвращает файл для скачивания
	CSVExporter()
}
