package main

import (
	"fmt"

	"github.com/alokrkmv/fetch_data"
	"github.com/alokrkmv/read_from_file"
)

func main() {
	ids := read_from_file.ReadFromTextFile()
	final_res,failed_ids:=fetch_data.GetData(ids)
	fmt.Println(final_res)
	fmt.Println(failed_ids)
}
