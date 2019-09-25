package main

import (
	"encoding/xml"
	"strconv"
	"testing"
)

func TestXMLParser(t *testing.T) {

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
        <room name='Great Auditorium'>
            <event guid='ed6568d5-48ed-5b18-90fc-a3bd52c7e543' id='55'>
                <date>2019-10-10T10:00:00+01:00</date>
                <start>10:00</start>
                <duration>00:45</duration>
                <room>Great Auditorium</room>
                <slug>eu2019-55-opening-session</slug>
                <url>https://test.dev/talk/77YFDG</url>
                <recording>
                    <license></license>
                    <optout>false</optout>
                </recording>
                <title>Opening session</title>
                <subtitle></subtitle>
                <track></track>
                <type>Long Talk</type>
                <language>en</language>
                <abstract>All cons need an opening session!</abstract>
                <description>Me and other org staff members will welcome all!</description>
                <persons>
                    <person id='4'>Tiago A</person>
                    <person id='5'>Tiago B</person>
                    <person id='6'>Tiago C</person>
                </persons>
                <links></links>
            </event>
            <event guid='d54ac76b-b13c-53a8-a418-f67502e351e7' id='63'>
                <date>2019-10-10T10:45:00+01:00</date>
                <start>10:45</start>
                <duration>00:30</duration>
                <room>Great Auditorium</room>
                <slug>eu2019-63-happy-15th-birthday-</slug>
                <url>https://test.dev/talk/FNKHDB</url>
                <recording>
                    <license></license>
                    <optout>false</optout>
                </recording>
                <title>Happy 15th birthday!</title>
                <subtitle></subtitle>
                <track></track>
                <type>Talk</type>
                <language>en</language>
                <abstract>It&apos;s an orignal idea , but he couldn&apos;t make it, so I&apos;ll be his voice!</abstract>
                <description>I&apos;ll make a retrospective of the first 15 years of our beloved Ubuntu!</description>
                <persons>
                    <person id='4'>Tiago A</person>
                </persons>
                <links></links>
            </event>
        </room>
        <room name='Another Room'>
            <event guid='337b1e60-d596-560e-86c6-12b706269be0' id='34'>
                <date>2019-10-10T11:30:00+01:00</date>
                <start>11:30</start>
                <duration>00:45</duration>
                <room>Another Room</room>
                <slug>eu2019-34-privacy-and-decentralisation-with-multicast</slug>
                <url>https://test.dev/talk/BVCA3C</url>
                <recording>
                    <license></license>
                    <optout>false</optout>
                </recording>
                <title>Privacy and Decentralisation with Multicast</title>
                <subtitle></subtitle>
                <track></track>
                <type>Long Talk</type>
                <language>en</language>
                <abstract>This talk explains why multicast is the missing piece in the decentralisation puzzle, how multicast can help the Internet continue to scale, better protect our privacy, solve IOT problems and make polar bears happier at the same time.</abstract>
                <description>Written in 2001, RFC 3170 states:.</description>
                <persons>
                    <person id='37'>BB</person>
                </persons>
                <links></links>
            </event>
            
        </room>
        
    </day>
    <day index='2' date='2019-10-11' start='2019-10-11T04:00:00' end='2019-10-12T03:59:00'>
        <room name='Great Auditorium'>
            <event guid='8fe2f64a-536a-5196-9fec-a773552263e4' id='19'>
                <date>2019-10-11T10:00:00+01:00</date>
                <start>10:00</start>
                <duration>00:30</duration>
                <room>Great Auditorium</room>
                <slug>eu2019-19-introduction-to-the-new-oracle-dba-mysql-oracle-</slug>
                <url>https://test.dev/talk/YU3RGP</url>
                <recording>
                    <license></license>
                    <optout>false</optout>
                </recording>
                <title>Introduction to the new Oracle DBA: MySQL &amp; Oracle.</title>
                <subtitle></subtitle>
                <track></track>
                <type>Talk</type>
                <language>en</language>
                <abstract>If you&apos;re new to MySQL but know other RDBMS&apos; come along to find out how to get rid of some fears and clarify all those doubts.</abstract>
                <description>Coming across new technology can be daunting and a challenge sometimes. In this session we aim at underlining that the technical aspect of the MySQL RDBMS are quite similar to pre-existing knowledge of other solutions. Find out what concepts are common to both, and what&apos;s different in the worlds most popular open source database.</description>
                <persons>
                    <person id='25'>KK</person>
                </persons>
                <links></links>
            </event>
            
        </room>
        
    </day>
</schedule>

`
	data := []byte(exampleXML)

	// You may test the direct url, but hardcoded values have changed :/
	// scheduleEventURL := "https://manage.ubucon.org/eu2019/schedule/export/schedule.xml"
	// // Get schedule from the official URL
	// resp, _ := http.Get(scheduleEventURL)
	// data, _ = ioutil.ReadAll(resp.Body)
	// resp.Body.Close()

	// Parse XML
	schedule := Schedule{}
	err := xml.Unmarshal([]byte(data), &schedule)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// Check some Conference meta
	if "Conference Title" != schedule.Conference.Title {
		t.Errorf("Unexpected Title: %s", schedule.Conference.Title)
	}
	if "2019-10-10" != schedule.Conference.Start {
		t.Errorf("Unexpected Start: %s", schedule.Conference.Start)
	}
	if 4 != schedule.Conference.Days {
		t.Errorf("Unexpected Days: %s", strconv.Itoa(schedule.Conference.Days))
	}

	// Check Lengths
	if 2 != len(schedule.Days) {
		t.Errorf("Unexpected length of days: %v", len(schedule.Days))
	}
	if 2 != len(schedule.Days[0].Rooms) {
		t.Errorf("Unexpected length of Rooms on day 0: %v", len(schedule.Days[0].Rooms))
	}

	// Check some information on day 0, room 0, event 0
	if "Great Auditorium" != schedule.Days[0].Rooms[0].Name {
		t.Errorf("Unexpected Room name: %s", schedule.Days[0].Rooms[0].Name)
	}
	if "Opening session" != schedule.Days[0].Rooms[0].Events[0].Title {
		t.Errorf("Unexpected event Title: %s", schedule.Days[0].Rooms[0].Events[0].Title)
	}
	if "https://test.dev/talk/77YFDG" != schedule.Days[0].Rooms[0].Events[0].URL {
		t.Errorf("Unexpected event URL: %s", schedule.Days[0].Rooms[0].Events[0].URL)
	}
	if "2019-10-10T10:00:00+01:00" != schedule.Days[0].Rooms[0].Events[0].Date {
		t.Errorf("Unexpected date time: %s", schedule.Days[0].Rooms[0].Events[0].Date)
	}
	if "10:00" != schedule.Days[0].Rooms[0].Events[0].Start {
		t.Errorf("Unexpected start time: %s", schedule.Days[0].Rooms[0].Events[0].Start)
	}
	if 3 != len(schedule.Days[0].Rooms[0].Events[0].Persons) {
		t.Errorf("Unexpected length of People: %v", len(schedule.Days[0].Rooms[0].Events[0].Persons))
	}
	if "Tiago A" != schedule.Days[0].Rooms[0].Events[0].Persons[0].Name {
		t.Errorf("Unexpected Person name: %s", schedule.Days[0].Rooms[0].Name)
	}

}
