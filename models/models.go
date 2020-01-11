package models

const UrlSchema = `
	CREATE TABLE IF NOT EXISTS url (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          url VARCHAR(255),
          short_url VARCHAR(255) UNIQUE,
          created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
          updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
`

const UserSchema = `
	CREATE TABLE IF NOT EXISTS user (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          username VARCHAR(255) UNIQUE,
          password VARCHAR(255),
          token VARCHAR(255),
          created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
          updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
`
