-- +goose Up
-- +goose StatementBegin
CREATE TABLE url
(
    id  VARCHAR(8) PRIMARY KEY,
    url VARCHAR(2048)
);
-- +goose StatementEnd