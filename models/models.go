package models

// UserSchema - User schema for postgres
const UserSchema = `CREATE TABLE IF NOT EXISTS appuser (id SERIAL PRIMARY KEY, username varchar(255) UNIQUE NOT NULL, password varchar(255) NULL, token varchar(255) NULL, created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL)`

// URLSchema - URL schema for postgres
const URLSchema = `CREATE TABLE IF NOT EXISTS appurl (id SERIAL, user_id INTEGER, url varchar(255) NOT NULL, short_url varchar(255) UNIQUE NULL, created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL, updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL, PRIMARY KEY(id, user_id), FOREIGN KEY (user_id) REFERENCES appuser (id) ON DELETE CASCADE)`
