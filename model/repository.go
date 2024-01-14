package model

import (
	_ "database/sql"

//	_ "github.com/mattn/go-sqlite3"
)



// stmt, err := db.Prepare("INSERT INTO OrderFile (Date, URL, Filename, Name, IsURLBroken) VALUES (?, ?, ?, ?, ?)")
// if err != nil {
//     log.Fatal(err)
// }
// defer stmt.Close()

// // Выполнение запроса с конкретными параметрами
// _, err = stmt.Exec("2024-01-14", "http://example.com", "example.txt", "Example", false)
// if err != nil {
//     log.Fatal(err)
// }