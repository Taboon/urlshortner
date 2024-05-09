-- +goose Up
CREATE TABLE users
(
    id SERIAL PRIMARY KEY
);

ALTER TABLE url
    ADD COLUMN is_deleted BOOLEAN,
    ADD COLUMN user_id    INTEGER,
    ADD CONSTRAINT url_user_id_fkey FOREIGN KEY (user_id) REFERENCES users (id);

-- +goose Down
DROP TABLE users;
ALTER TABLE url
    DROP COLUMN user_id,
    DROP COLUMN is_deleted;
