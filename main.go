package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"encoding/json"
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

// RoomInfo is the same structure of github.com/ubuconeurope/present-switch:RoomInfo
type RoomInfo struct {
	ID             int    `json:"room_id"` // room number
	RoomName       string `json:"room"`
	CurrentTitle   string `json:"title"`
	CurrentSpeaker string `json:"speaker"`
	CurrentTime    string `json:"time"`
	NextTitle      string `json:"n_title"`
	NextSpeaker    string `json:"n_speaker"`
	NextTime       string `json:"n_time"`
}

// createRoomInfoJSONBody creates a goroutine and request an update at the event time
func createRoomInfoJSONBody(room Room, event Event) []byte {
	var roomInfo RoomInfo

	// join multiple people per event
	var speakers []string
	for _, p := range event.Persons {
		speakers = append(speakers, fmt.Sprintf("%v", p.Name))
	}

	roomInfo.ID = room.ID
	roomInfo.RoomName = room.Name
	roomInfo.CurrentTitle = event.Title
	roomInfo.CurrentSpeaker = strings.Join(speakers, ", ")
	roomInfo.CurrentTime = event.Start

	// TODO: fill with next event
	roomInfo.NextTitle = ""
	roomInfo.NextSpeaker = ""
	roomInfo.NextTime = ""

	roomInfoJSON, err := json.Marshal(roomInfo)
	if err != nil {
		fmt.Println("Could not marshal roomInfo")
		panic(err)
	}
	return roomInfoJSON
}

func dispachEventUpdate(room Room, event Event, roomInfoJSON []byte) {
	// TODO: call goroutine
	fmt.Printf("    STUB: room %v - (%v) %v...\n", room.ID, event.Start, string(roomInfoJSON)[:60])

}

// function with side effects
func remapScheduleToEventsPerRoom(roomsMap *map[int]Room, eventsPerRoom *map[int][]Event, schedule Schedule) {
	for i, day := range schedule.Days {
		fmt.Printf("Processing Day %v: %v\n", i+1, day.Start)

		for _, room := range day.Rooms {
			fmt.Printf("= Processing Room: %v\n", room.Name)

			for _, event := range room.Events {
				(*roomsMap)[room.ID] = room
				(*eventsPerRoom)[room.ID] = append((*eventsPerRoom)[room.ID], event)
			}
		}

		fmt.Println("")
	}
}

// ScheduleEventUpdaters will create a goroutine for each event,
//   and request an update at the event time
func ScheduleEventUpdaters(schedule Schedule) {

	// map(Room.ID)Room
	roomsMap := make(map[int]Room)
	eventsPerRoom := make(map[int][]Event)

	remapScheduleToEventsPerRoom(&roomsMap, &eventsPerRoom, schedule)

	fmt.Println("#################")
	for roomID, eventsOnRoom := range eventsPerRoom {
		fmt.Printf("... Processing events for room %v: %v\n", roomID, roomsMap[roomID].Name)
		for _, event := range eventsOnRoom {
			fmt.Printf("... ... Processing event %v: %v: %v\n", event.GUID, event.Start, event.Title)
			roomInfoJSON := createRoomInfoJSONBody(roomsMap[roomID], event)
			// this will create the goroutine:
			dispachEventUpdate(roomsMap[roomID], event, roomInfoJSON)
		}
	}
	fmt.Println("#################")

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
	fmt.Println("############ Printing Schedule Info ############")
	PrintScheduleInfo(schedule)

	fmt.Println("############ Scheduling Event Updaters ############")
	ScheduleEventUpdaters(schedule)

	fmt.Println("############ Updates were scheduled. Just wait for them to finish... (WIP - NOT WORKING YET) ############")
}
