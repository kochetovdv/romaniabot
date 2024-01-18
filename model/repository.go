package model

const (
	CreateOrderFilesDB string = `CREATE TABLE IF NOT EXISTS OrderFiles
	(
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		Date TEXT NOT NULL,
		URL TEXT UNIQUE NOT NULL,
		Filename TEXT UNIQUE NOT NULL,
		Name TEXT NOT NULL,
		IsURLBroken BOOLEAN DEFAULT FALSE,
		IsDownloaded BOOLEAN DEFAULT FALSE,
		CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
		UpdatedAt DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	CreateOrdersDB string = `CREATE TABLE IF NOT EXISTS Orders
	(
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		FileId TEXT NOT NULL,
		Year TEXT NOT NULL,
		Number TEXT NOT NULL,
		FullNameFormatted TEXT NOT NULL UNIQUE,
		CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
		UpdatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (FileId) REFERENCES OrderFile (id) ON DELETE CASCADE
	)`
	Insert_Order_File string = `INSERT INTO OrderFiles (Date, URL, Filename, Name) VALUES (?, ?, ?, ?)`
	Get_new_Filenames string = `SELECT Filename FROM OrderFiles WHERE IsDownloaded = false;`
	Get_Valid_URLs    string = `SELECT URL FROM OrderFiles WHERE IsURLBroken = false AND IsDownloaded = false;`

	Get_Files_to_download string = `SELECT URL, Filename FROM OrderFiles WHERE IsURLBroken = false AND IsDownloaded = false;` // добавить не скачанные и не битые
	Set_broken_URLs       string = `UPDATE OrderFiles
	SET IsURLBroken = true
	WHERE URL = ?;`
	Set_is_Downloaded string = `UPDATE OrderFiles
	SET IsDownloaded = true
	WHERE Filename = ?;`
)
