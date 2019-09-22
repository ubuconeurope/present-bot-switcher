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
	Days    string `xml:"days"`
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
			fmt.Printf("= Room: %v\n", room.Name)
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

func main() {

	// Fix: maybe not so hardcoded
	scheduleEventURL := "https://manage.ubucon.org/eu2019/schedule/export/schedule.xml"

	// Get schedule from the official URL
	resp, err := http.Get(scheduleEventURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	// Parse XML
	schedule := Schedule{}
	err = xml.Unmarshal([]byte(body), &schedule)
	if err != nil {
		fmt.Printf("error: %v", err)
		return
	}

	// Print parsed XML info
	PrintScheduleInfo(schedule)
}
