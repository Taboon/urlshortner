-- +goose Up
CREATE TABLE users (
                       id SERIAL PRIMARY KEY
);

CREATE TABLE urls (
                      id VARCHAR(8) PRIMARY KEY,
                      url VARCHAR(2048),
                    deleted bool,
                      userid INT,
                      FOREIGN KEY (userID) REFERENCES users(id)
);

-- +goose Down
DROP TABLE urls;
DROP TABLE users;