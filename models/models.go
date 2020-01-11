package models

const UrlSchema = `
	CREATE TABLE IF NOT EXISTS url (
          id INTEGER PRIMARY KEY AUTOINCREMENT,
          url VARCHAR(255),
          shorturl VARCHAR(255),
          created_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
`
