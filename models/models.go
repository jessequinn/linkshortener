package models

const UrlSchema = `
	CREATE TABLE IF NOT EXISTS url (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          url VARCHAR(255),
          short_url VARCHAR(255) NULL UNIQUE,
          created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
`
