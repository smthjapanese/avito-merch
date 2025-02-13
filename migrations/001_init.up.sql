BEGIN;

-- Создаем таблицу пользователей
CREATE TABLE IF NOT EXISTS users (
                                     id SERIAL PRIMARY KEY,
                                     username VARCHAR(255) UNIQUE NOT NULL,
                                     password_hash VARCHAR(255) NOT NULL,
                                     coins INTEGER NOT NULL DEFAULT 1000,
                                     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создаем таблицу товаров (мерч)
CREATE TABLE IF NOT EXISTS merch_items (
                                           id SERIAL PRIMARY KEY,
                                           name VARCHAR(255) UNIQUE NOT NULL,
                                           price INTEGER NOT NULL CHECK (price > 0),
                                           created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создаем таблицу инвентаря пользователей
CREATE TABLE IF NOT EXISTS user_inventory (
                                              id SERIAL PRIMARY KEY,
                                              user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
                                              item_id INTEGER REFERENCES merch_items(id) ON DELETE CASCADE,
                                              quantity INTEGER NOT NULL DEFAULT 1,
                                              purchased_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                              CONSTRAINT unique_user_item UNIQUE(user_id, item_id)
);

-- Создаем таблицу транзакций
CREATE TABLE IF NOT EXISTS transactions (
                                            id SERIAL PRIMARY KEY,
                                            from_user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
                                            to_user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
                                            amount INTEGER NOT NULL CHECK (amount > 0),
                                            type VARCHAR(50) NOT NULL CHECK (type IN ('transfer', 'purchase')),
                                            item_id INTEGER REFERENCES merch_items(id) ON DELETE SET NULL,
                                            created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Создаем индексы для оптимизации запросов
CREATE INDEX IF NOT EXISTS idx_transactions_from_user ON transactions(from_user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_to_user ON transactions(to_user_id);
CREATE INDEX IF NOT EXISTS idx_user_inventory_user ON user_inventory(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at);

COMMIT;