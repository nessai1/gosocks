CREATE TABLE clients (
                         id INTEGER PRIMARY KEY AUTOINCREMENT,
                         login VARCHAR(255) UNIQUE,
                         password VARCHAR(255)
);