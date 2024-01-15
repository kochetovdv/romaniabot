package model

const (
	CreateDB string = `CREATE TABLE IF NOT EXISTS OrderFile (
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		Date TEXT,
		URL TEXT UNIQUE,
		Filename TEXT UNIQUE,
		Name TEXT,
		IsURLBroken BOOLEAN,
		IsDownloaded BOOLEAN DEFAULT FALSE,
		CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
		UpdatedAt DATETIME DEFAULT CURRENT_TIMESTAMP);`
	OrderFileToDB string = `INSERT INTO OrderFile (Date, URL, Filename, Name) VALUES (?, ?, ?, ?)`
	FilesToDownload string = `SELECT URL, Filename FROM OrderFile;`
)