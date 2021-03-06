package writer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

func WriteOutput(res map[string]string) error {
	// Fetch the location of the current directory
	mydir, _ := os.Getwd()
	// Generate the complete file path to write the data
	filePath := fmt.Sprintf("%s/%s", mydir, "writer/output.json")
	final_res := make(map[string]interface{})
	var res_array []interface{}
	for key, value := range res {
		temp_res := map[string]string{
			"Id":       key,
			"Response": value,
		}
		res_array = append(res_array, temp_res)
	}
	final_res["result"] = res_array
	// Marshaling to json
	file, _ := json.MarshalIndent(final_res, "", " ")
	err := ioutil.WriteFile(filePath, file, 0644)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil

}
