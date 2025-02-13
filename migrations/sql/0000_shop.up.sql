BEGIN;

CREATE TABLE users (
                       id SERIAL PRIMARY KEY,
                       username TEXT UNIQUE NOT NULL,
                       password_hash TEXT NOT NULL,
                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE coins (
                       user_id INT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
                       balance INT DEFAULT 1000 CHECK (balance >= 0)
);
CREATE TABLE inventory (
                           id SERIAL PRIMARY KEY,
                           user_id INT REFERENCES users(id) ON DELETE CASCADE,
                           item VARCHAR(255) NOT NULL,
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
    name VARCHAR(255),
    price INT NOT NULL
);



COMMIT;