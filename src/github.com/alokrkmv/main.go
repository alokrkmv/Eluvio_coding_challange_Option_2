package main

import (
	"fmt"

	"github.com/alokrkmv/fetch_data"
	"github.com/alokrkmv/read_from_file"
	"github.com/alokrkmv/writer"
)

func main() {
	ids := read_from_file.ReadFromTextFile()
	final_res, failed_ids := fetch_data.GetConcurrentData(ids)
	err := writer.WriteOutput(final_res)
	if err != nil {
		fmt.Println("Something went wrong in writing response to the output file")
	}
	// fmt.Println(final_res)
	fmt.Println(failed_ids)
}
