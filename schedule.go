package main

import (
	"errors"
	"fmt"
	"time"
)

type ScheduleObj struct {
	Schedule struct {
		Version    string `json:"version"`
		Conference struct {
			Days struct {
				Rooms map[string][]struct {
					Id       int       `json:"id"`
					Date     time.Time `json:"date"`
					Duration string    `json:"duration"`
					Title    string    `json:"title"`
					Subtitle string    `json:"subtitle"`
					Persons  []struct {
						Id        int    `json:"id"`
						Name      string `json:"name"`
						Biography string `json:"biography"`
					}
				}
			}
		}
	}
}

// GetRoomDataOfDay will return a map containing a specific format for all talks.
func (so *ScheduleObj) GetRoomDataOfDay(roomName string, day int) ([]map[string]string, error) {
	if day < 1 {
		return nil, errors.New("Cannot get invalid day. Day parameter must be more than 1.")
	}

	var result []map[string]string
	for _, el := range so.Schedule.Conference.Days.Rooms[roomName] {
		var author string
		if len(el.Persons) > 1 {
			author = so.getAuthorsNames(roomName, day)
		} else {
			author = el.Persons[0].Name
		}
		entry := map[string]string{
			"startingDate": el.Date.String(),
			"Title":        el.Title,
			"Author":       author,
		}
		result = append(result, entry)
	}
	return result, nil
}

// getAuthorsNames will return a string containing all authors names in a single row, useful for when there are more than just one author.
// It should return a string like X, Y and Z (in case of three authors.
func (so *ScheduleObj) getAuthorsNames(roomName string, day int) string {
	var AuthorNumber = len(so.Schedule.Conference.Days.Rooms[roomName][day].Persons)
	var result []byte
	for i := 0; i < len(so.Schedule.Conference.Days.Rooms[roomName][day].Persons); i++ {
		el := so.Schedule.Conference.Days.Rooms[roomName][day].Persons[i]
		if i == AuthorNumber-1 {
			result = append(result, fmt.Sprintf("%s and ", el.Name)...)
		} else {
			result = append(result, el.Name...)
		}
	}
	return string(result)
}
