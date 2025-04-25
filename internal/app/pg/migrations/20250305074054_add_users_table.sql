-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS urls (
	short_url TEXT NOT NULL PRIMARY KEY,
 	original_url TEXT NOT NULL UNIQUE,
	user_id INT,
	is_deleted BOOLEAN
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE urls;
-- +goose StatementEnd