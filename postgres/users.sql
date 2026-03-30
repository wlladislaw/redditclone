DROP TABLE IF EXISTS "users";
CREATE TABLE "users" (
    "id" SERIAL PRIMARY KEY,
    "login" varchar(50) UNIQUE NOT NULL,
    "password" varchar(255) NOT NULL,
    "created_at" timestamp DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO users (login, password) VALUES
('user1', 'qwerty123');