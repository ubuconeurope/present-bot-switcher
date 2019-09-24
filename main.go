package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"encoding/xml"
)

// Schedule is a sigleton containing all schedule info (see Days)
type Schedule struct {
	XMLName    xml.Name   `xml:"schedule"`
	Version    string     `xml:"version"`
	Conference Conference `xml:"conference"`
	Days       []Day      `xml:"day"`
}

// Conference contains conference info (meta)
type Conference struct {
	Acronym string `xml:"acronym"`
	Title   string `xml:"title"`
	Start   string `xml:"start"`
	End     string `xml:"end"`
	Days    int    `xml:"days"`
}

// Day contains each Day's schedule (per room)
type Day struct {
	Date  string `xml:"date,attr"`
	Start string `xml:"start,attr"`
	End   string `xml:"end,attr"`
	Rooms []Room `xml:"room"`
}

// Room contains each Room's schedule (each event)
type Room struct {
	ID     int
	Name   string  `xml:"name,attr"`
	Events []Event `xml:"event"`
}

// Event contains each talk data (the most important data)
type Event struct {
	ID          int      `xml:"id,attr"`
	GUID        string   `xml:"guid,attr"`
	Title       string   `xml:"title"`
	Start       string   `xml:"start"`    // Hour:Minute
	Duration    string   `xml:"duration"` // Hour:Minute
	URL         string   `xml:"url"`
	Slug        string   `xml:"slug"`
	Type        string   `xml:"type"`
	Abstract    string   `xml:"abstract"`
	Description string   `xml:"description"`
	Persons     []Person `xml:"persons>person"`
}

// Person is the person entity with ID
type Person struct {
	ID   int    `xml:"id,attr"`
	Name string `xml:",innerxml"`
}

// PrintScheduleInfo prints unmarshaled XML schedule
func PrintScheduleInfo(schedule Schedule) {
	fmt.Printf("XMLName: %#v\n", schedule.XMLName)
	fmt.Printf("Event: %v\n", schedule.Conference.Title)
	fmt.Printf("From %v to %v (%v days)\n\n", schedule.Conference.Start, schedule.Conference.End, schedule.Conference.Days)

	for i, day := range schedule.Days {
		fmt.Printf("Day %v: %v\n", i+1, day.Start)
		for _, room := range day.Rooms {
			fmt.Printf("= Room %v: %v\n", room.ID, room.Name)
			for _, event := range room.Events {

				// join multiple people per event
				var personsStr []string
				for _, s := range event.Persons {
					personsStr = append(personsStr, fmt.Sprintf("%v (%v)", s.Name, s.ID))
				}

				fmt.Printf("--- %v: %v - by %v\n", event.Start, event.Title, strings.Join(personsStr, ", "))
			}
		}
		fmt.Println("")
	}
}

// This function parses all rooms names and associate an id to them.
// ID starts at 1 and associates with first seen room name.
// Then it associates that ID with the room obj inside schedule.Days.
func fixScheduleRoomsID(schedule *Schedule) {
	roomsSlice := make([]string, 0, 10)

	// iterate over rooms on all days
	for d, day := range schedule.Days {
	RoomsLoop:
		for r, room := range day.Rooms {

			// keep updating the roomsSlice if a new name comes.
			// also updates the room.ID (starting in 1)
			for i, rSlice := range roomsSlice {
				if rSlice == room.Name {
					schedule.Days[d].Rooms[r].ID = i + 1
					continue RoomsLoop // just continue to the next room
				}
			}

			// only reaches this line if the room.Name is not stored yet
			roomsSlice = append(roomsSlice, room.Name)
			schedule.Days[d].Rooms[r].ID = len(roomsSlice)
		}
	}
}

func main() {

	// Fix: maybe not so hardcoded
	scheduleEventURL := "https://manage.ubucon.org/eu2019/schedule/export/schedule.xml"

	// Get schedule from the official URL
	resp, err := http.Get(scheduleEventURL)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	// Parse XML
	schedule := Schedule{}
	err = xml.Unmarshal([]byte(body), &schedule)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}
	fixScheduleRoomsID(&schedule)

	// Print parsed XML info
	PrintScheduleInfo(schedule)
}
