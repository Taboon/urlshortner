-- +goose Up
-- +goose StatementBegin
CREATE TABLE urls (
                                    id TEXT PRIMARY KEY,
                                    url TEXT UNIQUE
);
-- +goose StatementEnd