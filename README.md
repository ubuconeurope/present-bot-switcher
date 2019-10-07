# Presentation Bot Switcher (updater)

This Bot reads a xml Schedule file (like https://manage.ubucon.org/eu2019/schedule/export/schedule.xml) 
and calls the `EXTERNAL_UPDATE_URL` in order to update the Room Information for the next event. 

If there is no internet connection, it will fallback to reading the `schedule.xml` local file. (you may force this with Env variables.

This Bot is intended to integrate with [UbuconEU Present Switch](https://github.com/ubuconeurope/present-switch).


## How to run 

If running with defaults:
`go run main.go`

Use non-defaults with Environment variables:

```
EXTERNAL_UPDATE_URL="http://user:passw@localhost:3000/rooms/ SCHEDULE_URL="" SCHEDULE_FILE="localschedule.xml"  go run main.go
```

## Environment Variables (and defaults)

```
SCHEDULE_URL="https://manage.ubucon.org/eu2019/schedule/export/schedule.xml"
SCHEDULE_FILE="schedule.xml"
EXTERNAL_UPDATE_URL="http://user:passw@localhost:3000/rooms/"
TEST_MODE="false"
```

## Test mode

In order to test everything, you may enable the `TEST_MODE=true`.
This will update the events every second (instead of waiting for the event's time. 
With this, you can see clients updating their room information.