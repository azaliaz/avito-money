BEGIN;

CREATE TABLE users 
(
                       id SERIAL PRIMARY KEY,
                       username TEXT UNIQUE NOT NULL,
                       password_hash TEXT NOT NULL,
                       balance INT DEFAULT 1000 CHECK (balance >= 0),
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE inventory (
                           id SERIAL PRIMARY KEY,
                           user_id INT REFERENCES users(id) ON DELETE CASCADE,
                           item TEXT NOT NULL,
                           quantity INT DEFAULT 0 CHECK (quantity >= 0),
                           UNIQUE (user_id, item)
);
CREATE TABLE transactions (
                              id SERIAL PRIMARY KEY,
                              from_user_id INT REFERENCES users(id) ON DELETE
                                  SET
                                  NULL,
                              to_user_id INT REFERENCES users(id) ON DELETE
                                  SET
                                  NULL,
                              amount INT NOT NULL CHECK (amount > 0),
                              created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE items (
    name TEXT,
    price INT NOT NULL
);

INSERT INTO items (name, price) VALUES
                                    ('t-shirt', 80),
                                    ('cup', 20),
                                    ('book', 50),
                                    ('pen', 10),
                                    ('powerbank', 200),
                                    ('hoody', 300),
                                    ('umbrella', 200),
                                    ('socks', 10),
                                    ('wallet', 50),
                                    ('pink-hoody', 500);

COMMIT;