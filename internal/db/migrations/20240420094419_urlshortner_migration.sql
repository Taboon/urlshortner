-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS urls (
                                    id TEXT PRIMARY KEY,
                                    url TEXT,
                                    count INT
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
