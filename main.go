package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

const FILE_PATH string = "./data/data.json"

func main() {
	file, err := ioutil.ReadFile(FILE_PATH)
	if err != nil {
		fmt.Println("Error reading file. Does it exist?")
		os.Exit(1)
	}

	var schedule ScheduleObj
	err = json.Unmarshal(file, &schedule)
	if err != nil {
		fmt.Printf("Error parsing JSON file: \n %s", err)
		os.Exit(1)
	}

	fmt.Println(schedule)
	// 	day := schedule.conf.days[i]
	// 	fmt.Printf("Got room %d. Printing start and end dates. \n", i)
	// 	fmt.Println(day.day_start)
	// 	fmt.Println(day.day_end)

}
