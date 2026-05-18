package main

import (
	"fmt"
	"mydb/internal/wal"
)

func main() {
	data := "Hello, World"
	w, err := wal.New("test.bin")

	if err != nil {
		fmt.Println(err)
	}

	w.Append(&wal.Begin{XID: 1})
	w.Append(&wal.Update{XID: 1, PageID: 1, Offset: 1, After: []byte(data)})
}
