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
