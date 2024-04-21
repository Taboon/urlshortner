-- +goose Up
-- +goose StatementBegin
CREATE TABLE urls (
                                    id TEXT PRIMARY KEY,
                                    url TEXT
);
-- +goose StatementEnd