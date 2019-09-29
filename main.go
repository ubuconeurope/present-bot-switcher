package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"encoding/json"
	"encoding/xml"
)

// GetEnv returns the Environment variable by key, or return a fallback value if the key is not set
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var scheduleEventURL = GetEnv("SCHEDULE_URL", "https://manage.ubucon.org/eu2019/schedule/export/schedule.xml")
var altLocalScheduleFile = GetEnv("SCHEDULE_FILE", "schedule.xml")
var externalUpdateURL = GetEnv("EXTERNAL_UPDATE_URL", "http://user@passw:localhost:3000/room/")
var wg sync.WaitGroup
var waitCounter time.Duration = 1 * time.Second

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
	Date        string   `xml:"date"`
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
		log.Println("Could not marshal roomInfo")
		panic(err)
	}
	return roomInfoJSON
}

func callEventUpdater(waitDuration time.Duration, URL string, roomInfoJSON []byte) {
	defer wg.Done()

	time.Sleep(waitDuration)
	log.Printf("CALL POST %v - %v\n", URL, string(roomInfoJSON))
	resp, err := http.Post(URL, "application/json", bytes.NewBuffer(roomInfoJSON))

	if err != nil {
		log.Println(err)
	} else if resp.StatusCode != http.StatusOK {
		log.Println(resp)
	}
}

// ParseCustomDuration parses HH:MM format. Returns 0 duration on error.
func ParseCustomDuration(durationStr string) (time.Duration, error) {
	var duration time.Duration

	hoursMinutes := strings.Split(durationStr, ":")

	if len(hoursMinutes) != 2 {
		return duration, fmt.Errorf("error: invalid format for durationStr. Expected 'hh:mm' got: %v", durationStr)
	}

	hours, err := strconv.Atoi(hoursMinutes[0])
	if err != nil {
		return duration, fmt.Errorf("error parsing hour value (%v): %v", durationStr, err)
	}
	mins, err2 := strconv.Atoi(hoursMinutes[1])
	if err2 != nil {
		return duration, fmt.Errorf("error parsing minute value (%v): %v", durationStr, err)
	}

	duration = time.Duration(hours)*time.Hour + time.Duration(mins)*time.Minute
	return duration, nil
}

func dispachEventUpdate(room Room, event Event, roomInfoJSON []byte) {
	eventTime, err := time.Parse("2006-01-02T15:04:05-07:00", event.Date)
	if err != nil {
		log.Println("ERROR parsing date time. ", err)
	}

	nowTime := time.Now()
	durationUntilEvent := eventTime.Sub(nowTime)

	roomURL := externalUpdateURL + strconv.Itoa(room.ID)
	log.Printf("(updating in %v) %v - %v...\n", durationUntilEvent, roomURL, string(roomInfoJSON)[:60])

	wg.Add(1)
	go callEventUpdater(durationUntilEvent, roomURL, roomInfoJSON)
}

// function with side effects
func remapScheduleToEventsPerRoom(roomsMap *map[int]Room, eventsPerRoom *map[int][]Event, schedule Schedule) {
	for i, day := range schedule.Days {
		log.Printf("Processing Day %v: %v\n", i+1, day.Start)

		for _, room := range day.Rooms {
			log.Printf("= Processing Room: %v\n", room.Name)

			for _, event := range room.Events {
				(*roomsMap)[room.ID] = room
				(*eventsPerRoom)[room.ID] = append((*eventsPerRoom)[room.ID], event)
			}
		}

		log.Println("")
	}
}

// ScheduleEventUpdaters will create a goroutine for each event,
//   and request an update at the event time
func ScheduleEventUpdaters(schedule Schedule) {

	// map(Room.ID)Room
	roomsMap := make(map[int]Room)
	eventsPerRoom := make(map[int][]Event)

	remapScheduleToEventsPerRoom(&roomsMap, &eventsPerRoom, schedule)

	log.Println("#################")
	for roomID, eventsOnRoom := range eventsPerRoom {
		log.Printf("... Processing events for room %v: %v\n", roomID, roomsMap[roomID].Name)
		for _, event := range eventsOnRoom {
			log.Printf("... ... Processing event %v: %v: %v\n", event.GUID, event.Date, event.Title)
			roomInfoJSON := createRoomInfoJSONBody(roomsMap[roomID], event)

			// this will create the goroutine:
			dispachEventUpdate(roomsMap[roomID], event, roomInfoJSON)
		}
	}
	log.Println("#################")

}

// PrintScheduleInfo prints unmarshaled XML schedule
func PrintScheduleInfo(schedule Schedule) {
	log.Printf("XMLName: %#v\n", schedule.XMLName)
	log.Printf("Event: %v\n", schedule.Conference.Title)
	log.Printf("From %v to %v (%v days)\n\n", schedule.Conference.Start, schedule.Conference.End, schedule.Conference.Days)

	for i, day := range schedule.Days {
		log.Printf("Day %v: %v\n", i+1, day.Start)
		for _, room := range day.Rooms {
			log.Printf("= Room %v: %v\n", room.ID, room.Name)
			for _, event := range room.Events {

				// join multiple people per event
				var personsStr []string
				for _, s := range event.Persons {
					personsStr = append(personsStr, fmt.Sprintf("%v (%v)", s.Name, s.ID))
				}

				log.Printf("--- %v: %v - by %v\n", event.Start, event.Title, strings.Join(personsStr, ", "))
			}
		}
		log.Println("")
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
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var body []byte
	// Get schedule from the official URL, or failback to local file
	resp, err := http.Get(scheduleEventURL)
	if err != nil {
		log.Println("WARNING: Could not read remote URL. Fallbacking to local file")
		body, err = ioutil.ReadFile(altLocalScheduleFile)
		if err != nil {
			log.Println("Error reading file. Does it exist?")
			panic(err)
		}
	} else {
		body, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
	}

	// Parse XML
	schedule := Schedule{}
	err = xml.Unmarshal([]byte(body), &schedule)
	if err != nil {
		log.Printf("error: %v", err)
		return
	}
	fixScheduleRoomsID(&schedule)

	// Print parsed XML info
	log.Println("############ Printing Schedule Info ############")
	PrintScheduleInfo(schedule)

	log.Println("############ Scheduling Event Updaters ############")
	ScheduleEventUpdaters(schedule)

	log.Println("############ Updates were scheduled. Just wait for them to finish... ############")

	wg.Wait()
	log.Println("Finished! No more events to update")
}
