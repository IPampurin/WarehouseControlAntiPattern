package db

import (
	"context"
	"fmt"
)

const (
	usersSchema = `CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                     name TEXT NOT NULL,
                    email TEXT UNIQUE NOT NULL,
                     role VARCHAR(20) NOT NULL DEFAULT 'viewer',
               created_at TIMESTAMPTZ NOT NULL DEFAULT NOW());`

	itemsSchema = `CREATE TABLE items (
                       id SERIAL PRIMARY KEY,
                     name TEXT NOT NULL,
                 quantity INT NOT NULL DEFAULT 0 CHECK (quantity >= 0),
                    price DECIMAL(10,2) NOT NULL CHECK (price >= 0),
               created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
               updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW());`

	itemHistorySchema = `CREATE TABLE item_history (
                             id SERIAL PRIMARY KEY,
                        item_id INT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
                         action VARCHAR(10) NOT NULL,
                       old_data JSONB,
                       new_data JSONB,
                     changed_by INT REFERENCES users(id) ON DELETE SET NULL,
                     changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW());

                         CREATE INDEX idx_item_history_item_id ON item_history(item_id);
                         CREATE INDEX idx_item_history_changed_at ON item_history(changed_at);
                         CREATE INDEX idx_item_history_changed_by ON item_history(changed_by);`

	triggerSchema = ` CREATE FUNCTION log_item_changes()
	                 RETURNS TRIGGER AS $$
                             BEGIN
                                   IF TG_OP = 'INSERT' THEN
                                       INSERT INTO item_history (item_id, action, new_data, changed_by, changed_at)
                                       VALUES (NEW.id, 'INSERT', row_to_json(NEW), current_setting('app.current_user_id', true)::int, NOW());
                                   ELSIF TG_OP = 'UPDATE' THEN
                                       INSERT INTO item_history (item_id, action, old_data, new_data, changed_by, changed_at)
                                       VALUES (OLD.id, 'UPDATE', row_to_json(OLD), row_to_json(NEW), current_setting('app.current_user_id', true)::int, NOW());
                                   ELSIF TG_OP = 'DELETE' THEN
                                       INSERT INTO item_history (item_id, action, old_data, changed_by, changed_at)
                                       VALUES (OLD.id, 'DELETE', row_to_json(OLD), current_setting('app.current_user_id', true)::int, NOW());
                                   END IF;
                             RETURN NULL;
                     END;
                     $$ LANGUAGE plpgsql;

                     CREATE TRIGGER item_changes_trigger
                      AFTER INSERT OR UPDATE OR DELETE ON items
                        FOR EACH ROW EXECUTE FUNCTION log_item_changes();`
)

// Migration создаёт таблицы БД, если они ещё не существуют, добавляет индексы
func (d *DataBase) Migration(ctx context.Context) error {

	// создаём таблицу users
	query := usersSchema
	_, err := d.Pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы users: %w", err)
	}

	// создаём таблицу items
	query = itemsSchema
	_, err = d.Pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы items: %w", err)
	}

	// создаём таблицу item_history с индексами
	query = itemHistorySchema
	_, err = d.Pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("ошибка создания таблицы item_history: %w", err)
	}

	// создаём триггеры
	if _, err := d.Pool.Exec(ctx, triggerSchema); err != nil {
		return fmt.Errorf("ошибка создания триггеров: %w", err)
	}

	return nil
}
