package read_from_file

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// This function reads ids from a text file
func ReadFromTextFile() []string {

	// Fetch the location of the current directory
	mydir, _ := os.Getwd()
	// Generate the complete file path
	filePath := fmt.Sprintf("%s/%s", mydir, "read_from_file/ids.txt")

	content, err := ioutil.ReadFile(filePath)

	fmt.Println(mydir)

	if err != nil {
		log.Fatal(err)
	}

	ids := strings.Split(string(content), ",")
	return ids
}
