package main

import (
	"encoding/xml"
	"testing"
)

func TestFixScheduleRoomsID(t *testing.T) {

	exampleXML := `
<?xml version='1.0' encoding='utf-8' ?>
<!-- Made with love by pretalx v0.9.0. -->
<schedule>
    <version>0.13</version>
    <conference>
        <acronym>eu2019</acronym>
        <title>Conference Title</title>
        <start>2019-10-10</start>
        <end>2019-10-13</end>
        <days>4</days>
        <timeslot_duration>00:05</timeslot_duration>
    </conference>
    <day index='1' date='2019-10-10' start='2019-10-10T04:00:00' end='2019-10-11T03:59:00'>
        <room name='Room1'></room>
        <room name='Room2'></room>
        <room name='Room3'></room>
        <room name='Room3' comment='This is repeated on purpose, for testing'></room>
        <room name='Room4'></room>
        <room name='Room5'></room>
		</day>
    <day index='2' date='2019-10-11' start='2019-10-11T04:00:00' end='2019-10-12T03:59:00'>
        <room name='Room4'></room>
        <room name='Room6'></room>
		</day>
</schedule>
`

	// Parse XML
	schedule := Schedule{}
	err := xml.Unmarshal([]byte(exampleXML), &schedule)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	fixScheduleRoomsID(&schedule)

	if schedule.Days[0].Rooms[0].ID != 1 {
		t.Error("The first room.ID should be 1. Instead got:", schedule.Days[0].Rooms[0].ID)
	}
	if schedule.Days[0].Rooms[3].ID != 3 {
		t.Error("The fourth room is repeated and should have ID=3. Instead got:", schedule.Days[0].Rooms[3].ID)
	}
	if schedule.Days[1].Rooms[0].ID != 4 {
		t.Error("The first room.ID on the second day should be 4. Instead got:", schedule.Days[1].Rooms[0].ID)
	}
	if schedule.Days[0].Rooms[5].ID != 5 {
		t.Error("The sixth room.ID (day 1) should be 5. Instead got:", schedule.Days[0].Rooms[5].ID)
	}
	if schedule.Days[1].Rooms[1].ID != 6 {
		t.Error("The second room.ID (day 2) should be 6. Instead got:", schedule.Days[1].Rooms[1].ID)
	}
}

func TestCreateRoomInfoJSONBody(t *testing.T) {
	room := Room{ID: 1, Name: "RoomName"}
	event := Event{
		Title: "EventTitle",
		Start: "10:00",
		Persons: []Person{
			Person{
				Name: "PersonName1",
			},
			Person{
				Name: "PersonName2",
			},
		},
	}

	roomInfoJSON := createRoomInfoJSONBody(room, event)
	expectedJSON := `{"room_id":1,"room":"RoomName","title":"EventTitle","speaker":"PersonName1, PersonName2","time":"10:00","n_title":"","n_speaker":"","n_time":""}`

	if string(roomInfoJSON) != expectedJSON {
		t.Errorf("Result was not expected\n\nGot:\n%v\nExpected:\n%v\n..........\n", string(roomInfoJSON), expectedJSON)
	}
}

func TestRemapScheduleToEventsPerRoom(t *testing.T) {

	testingSchedule := Schedule{
		Days: []Day{
			Day{
				Rooms: []Room{
					Room{
						ID:   1,
						Name: "Room1",
						Events: []Event{
							Event{
								GUID:  "abc1",
								Title: "Event1",
							},
							Event{
								GUID:  "abc2",
								Title: "Event2",
							},
							Event{
								GUID:  "abc3",
								Title: "Event3",
							},
						},
					},
					Room{
						ID:   2,
						Name: "Room2",
						Events: []Event{
							Event{
								GUID:  "abc4",
								Title: "Event4",
							},
							Event{
								GUID:  "abc5",
								Title: "Event5",
							},
							Event{
								GUID:  "abc6",
								Title: "Event6",
							},
						},
					},
				},
			},
			Day{
				Rooms: []Room{
					Room{
						ID:   1,
						Name: "Room1",
						Events: []Event{
							Event{
								GUID:  "cde1",
								Title: "OtherEvent1",
							},
						},
					},
					Room{
						ID:   2,
						Name: "Room2",
						Events: []Event{
							Event{
								GUID:  "cde4",
								Title: "OtherEvent4",
							},
							Event{
								GUID:  "cde5",
								Title: "OtherEvent5",
							},
							Event{
								GUID:  "cde6",
								Title: "OtherEvent6",
							},
						},
					},
				},
			},
		},
	}

	roomsMap := make(map[int]Room)
	eventsPerRoom := make(map[int][]Event)

	remapScheduleToEventsPerRoom(&roomsMap, &eventsPerRoom, testingSchedule)

	if len(roomsMap) != 2 || len(eventsPerRoom) != 2 {
		t.Error("Unexpected length for 2 rooms. Got", len(roomsMap), ",", len(eventsPerRoom))
	}

	// ########## Test roomMap
	expectedRoomMap := map[int]Room{
		1: Room{ID: 1, Name: "Room1"},
		2: Room{ID: 2, Name: "Room2"},
	}
	for i, r := range roomsMap {
		if r.ID != expectedRoomMap[i].ID || r.Name != expectedRoomMap[i].Name {
			t.Errorf("Unexpected roomMap on id=%v.\nGot: %v-%v\nExpected: %v-%v", i, r.ID, r.Name, expectedRoomMap[i].ID, expectedRoomMap[i].Name)
		}
	}

	// ########## Test eventsPerRoom
	expectedEventsPerRoom := map[int][]struct {
		GUID  string
		title string
	}{
		1: {
			{GUID: "abc1", title: "Event1"},
			{GUID: "abc2", title: "Event2"},
			{GUID: "abc3", title: "Event3"},
			{GUID: "cde1", title: "OtherEvent1"},
		},
		2: {
			{GUID: "abc4", title: "Event4"},
			{GUID: "abc5", title: "Event5"},
			{GUID: "abc6", title: "Event6"},
			{GUID: "cde4", title: "OtherEvent4"},
			{GUID: "cde5", title: "OtherEvent5"},
			{GUID: "cde6", title: "OtherEvent6"},
		},
	}

	for i, evts := range eventsPerRoom {
		for j, ev := range evts {
			if ev.GUID != expectedEventsPerRoom[i][j].GUID || ev.Title != expectedEventsPerRoom[i][j].title {
				t.Errorf("Unexpected eventsPerRoom on id=%v.\nGot: %v-%v\nExpected: %v-%v", i, ev.GUID, ev.Title, expectedEventsPerRoom[i][j].GUID, expectedEventsPerRoom[i][j].title)
			}
		}
	}

}
