-- +goose Up
CREATE TABLE IF NOT EXISTS budgets (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    category TEXT NOT NULL,
    limit_amount NUMERIC(14,2) NOT NULL CHECK (limit_amount > 0),
    period TEXT NOT NULL DEFAULT 'fixed',
    UNIQUE(user_id, category)
);

CREATE TABLE IF NOT EXISTS expenses (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    amount NUMERIC(14,2) NOT NULL CHECK (amount <> 0),
    category TEXT NOT NULL,
    description TEXT,
    date DATE NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_expenses_user_cat_date ON expenses(user_id, category, date);

-- +goose Down
DROP TABLE IF EXISTS expenses;
DROP TABLE IF EXISTS budgets;
