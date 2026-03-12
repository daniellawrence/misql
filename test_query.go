//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:33060)/company")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, name, email FROM users")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(cols)

	for rows.Next() {
		vals := make([]string, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			log.Fatal(err)
		}
		for i, col := range cols {
			fmt.Printf("%s=%s ", col, vals[i])
		}
		fmt.Println()
	}
}
