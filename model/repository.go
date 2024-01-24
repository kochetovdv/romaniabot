package model

const (
	CreateOrderFilesDB string = `CREATE TABLE IF NOT EXISTS OrderFiles
	(
		Filename TEXT UNIQUE NOT NULL,
		Date TEXT NOT NULL,
		URL TEXT UNIQUE NOT NULL,
		Name TEXT NOT NULL,
		IsURLBroken BOOLEAN DEFAULT FALSE,
		IsDownloaded BOOLEAN DEFAULT FALSE,
		IsParsed BOOLEAN DEFAULT FALSE,
		CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
		UpdatedAt DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	CreateOrdersDB string = `CREATE TABLE IF NOT EXISTS Orders
	(
		Filename TEXT NOT NULL,
		Number INT NOT NULL,
		Year INT NOT NULL,
		FullNameFormatted TEXT NOT NULL UNIQUE,
		CreatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
		UpdatedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (Filename) REFERENCES OrderFile (Filename) ON DELETE CASCADE
	)`
	Insert_Order_File string = `INSERT INTO OrderFiles (Date, URL, Filename, Name) VALUES (?, ?, ?, ?)`
	Insert_Order      string = `INSERT INTO Orders (Filename, Number, Year, FullNameFormatted) VALUES (?, ?, ?, ?)`
	Get_new_Filenames string = `SELECT Filename FROM OrderFiles WHERE IsDownloaded = false;`
	Get_Valid_URLs    string = `SELECT URL FROM OrderFiles WHERE IsURLBroken = false AND IsDownloaded = false;`

	Get_Files_to_download string = `SELECT URL, Filename FROM OrderFiles WHERE IsURLBroken = false AND IsDownloaded = false;`
	Get_Files_not_parsed  string = `SELECT Filename FROM OrderFiles WHERE IsParsed = false;`
	//	Get_Files_downloaded_to_parse string = `SELECT Filename FROM OrderFiles WHERE IsParsed = false AND IsDownloaded = true;`
	Set_broken_URLs string = `UPDATE OrderFiles
	SET IsURLBroken = true
	WHERE URL = ?;`
	Set_is_Downloaded string = `UPDATE OrderFiles
	SET IsDownloaded = true
	WHERE Filename = ?;`
	Set_is_Parsed string = `UPDATE OrderFiles
	SET IsParsed = true
	WHERE Filename = ?;`
)
