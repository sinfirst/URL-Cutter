-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS urls (
	short_url TEXT NOT NULL PRIMARY KEY,
 	original_url TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd