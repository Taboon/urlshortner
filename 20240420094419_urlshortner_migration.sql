-- +goose Up
-- +goose StatementBegin
CREATE TABLE urls (
                                    id VARCHAR(8) PRIMARY KEY,
                                    url VARCHAR(2048) UNIQUE
);
-- +goose StatementEnd