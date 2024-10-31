CREATE TABLE secrets(
    name VARCHAR(200) UNIQUE NOT NULL,
    type VARCHAR(32) NOT NULL,
    meta VARCHAR(255),
    data BYTEA NOT NULL,
    version BIGINT NOT NULL,
    PRIMARY KEY (name)
);

