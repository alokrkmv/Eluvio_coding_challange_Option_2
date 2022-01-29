package main

import (
	"fmt"
	"time"

	"github.com/alokrkmv/fetch_data"
	"github.com/alokrkmv/read_from_file"
	"github.com/alokrkmv/writer"
)

func main() {
	ids := read_from_file.ReadFromTextFile()
	start_time := time.Now()
	final_res, meta_map := fetch_data.GetConcurrentData(ids)
	end_time := time.Now()
	total_response_time := end_time.Sub(start_time)
	meta_map["total_response_time"] = total_response_time
	go writer.MetaDataWriter(meta_map)
	err := writer.WriteOutput(final_res)

	// Writing meta data is not a primary task so we will spawn a go routine to
	// write meta data hence it won't hamper the performance of the actual program

	if err != nil {
		fmt.Println("Something went wrong in writing response to the output file")
	}

}
