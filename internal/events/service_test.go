package events

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/trly/today/internal/caldav"
)

func TestCalendarsListsAllSortedByName(t *testing.T) {
	client := &fakeClient{calendars: []caldav.Calendar{
		{Name: "Work", Path: "/work/", Color: "#c15a2a"},
		{Name: "Home", Path: "/home/"},
	}}
	service := Service{Client: client}

	got, err := service.Calendars(context.Background())
	if err != nil {
		t.Fatalf("Calendars() error = %v", err)
	}
	want := []Calendar{
		{Name: "Home", Path: "/home/"},
		{Name: "Work", Path: "/work/", Color: "#c15a2a"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Calendars() = %#v, want %#v", got, want)
	}
}

func TestTodayFetchesRequestedCalendarsForCurrentLocalDay(t *testing.T) {
	loc := time.FixedZone("test", -5*60*60)
	client := &fakeClient{calendars: []caldav.Calendar{
		{Name: "Work", Path: "/work/", Color: "#c15a2a"},
		{Name: "Home", Path: "/home/"},
	}}
	client.events = map[string][]caldav.Event{
		"Work": {
			{Title: "later", Start: time.Date(2026, 7, 14, 15, 0, 0, 0, loc)},
			{Title: "earlier", Start: time.Date(2026, 7, 14, 9, 0, 0, 0, loc)},
		},
	}
	service := Service{
		Client:   client,
		Location: loc,
		Now: func() time.Time {
			return time.Date(2026, 7, 14, 12, 0, 0, 0, loc)
		},
	}

	got, err := service.Today(context.Background(), []string{" Work ", "Work"})
	if err != nil {
		t.Fatalf("Today() error = %v", err)
	}
	if got.Date != "2026-07-14" {
		t.Fatalf("Today().Date = %q, want 2026-07-14", got.Date)
	}
	if len(got.Events) != 2 {
		t.Fatalf("len(Today().Events) = %d, want 2", len(got.Events))
	}
	if got.Events[0].Title != "earlier" || got.Events[1].Title != "later" {
		t.Fatalf("events not sorted by start: %#v", got.Events)
	}
	if got.Events[0].Calendar != "Work" || got.Events[0].CalendarColor != "#c15a2a" {
		t.Fatalf("event calendar metadata = %#v", got.Events[0])
	}
	if !client.lastStart.Equal(time.Date(2026, 7, 14, 0, 0, 0, 0, loc)) {
		t.Fatalf("start = %v, want local day start", client.lastStart)
	}
	if !client.lastEnd.Equal(time.Date(2026, 7, 15, 0, 0, 0, 0, loc)) {
		t.Fatalf("end = %v, want next local day start", client.lastEnd)
	}
}

func TestDayFetchesRequestedDateInServiceLocation(t *testing.T) {
	loc := time.FixedZone("test", -5*60*60)
	client := &fakeClient{calendars: []caldav.Calendar{{Name: "Work", Path: "/work/"}}}
	service := Service{Client: client, Location: loc}

	got, err := service.Day(context.Background(), []string{"Work"}, "2026-07-15")
	if err != nil {
		t.Fatalf("Day() error = %v", err)
	}
	if got.Date != "2026-07-15" {
		t.Fatalf("Day().Date = %q, want 2026-07-15", got.Date)
	}
	if !client.lastStart.Equal(time.Date(2026, 7, 15, 0, 0, 0, 0, loc)) {
		t.Fatalf("start = %v, want requested local day start", client.lastStart)
	}
	if !client.lastEnd.Equal(time.Date(2026, 7, 16, 0, 0, 0, 0, loc)) {
		t.Fatalf("end = %v, want next local day start", client.lastEnd)
	}
}

func TestTodayFiltersTimedEventsToCurrentLocalDay(t *testing.T) {
	loc := time.FixedZone("test", -5*60*60)
	client := &fakeClient{calendars: []caldav.Calendar{{Name: "Work", Path: "/work/"}}}
	client.events = map[string][]caldav.Event{
		"Work": {
			{
				Title: "yesterday timed",
				Start: time.Date(2026, 7, 13, 23, 0, 0, 0, loc),
				End:   time.Date(2026, 7, 14, 0, 0, 0, 0, loc),
			},
			{
				Title:    "today timed",
				Start:    time.Date(2026, 7, 14, 9, 0, 0, 0, loc),
				End:      time.Date(2026, 7, 14, 10, 0, 0, 0, loc),
				Priority: 5,
			},
			{
				Title: "tomorrow timed",
				Start: time.Date(2026, 7, 15, 0, 0, 0, 0, loc),
				End:   time.Date(2026, 7, 15, 1, 0, 0, 0, loc),
			},
			{
				Title:       "today parsed as utc",
				Start:       time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC),
				End:         time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC),
				AllDay:      true,
				AllDayStart: "2026-07-14",
				AllDayEnd:   "2026-07-15",
			},
			{
				Title:       "yesterday all day",
				Start:       time.Date(2026, 7, 13, 0, 0, 0, 0, loc),
				End:         time.Date(2026, 7, 14, 0, 0, 0, 0, loc),
				AllDay:      true,
				AllDayStart: "2026-07-13",
				AllDayEnd:   "2026-07-14",
			},
			{
				Title:       "spans today",
				Start:       time.Date(2026, 7, 13, 0, 0, 0, 0, loc),
				End:         time.Date(2026, 7, 15, 0, 0, 0, 0, loc),
				AllDay:      true,
				AllDayStart: "2026-07-13",
				AllDayEnd:   "2026-07-15",
				Priority:    2,
			},
			{
				Title:       "today all day",
				Start:       time.Date(2026, 7, 14, 0, 0, 0, 0, loc),
				End:         time.Date(2026, 7, 15, 0, 0, 0, 0, loc),
				AllDay:      true,
				AllDayStart: "2026-07-14",
				AllDayEnd:   "2026-07-15",
			},
			{
				Title:       "tomorrow all day",
				Start:       time.Date(2026, 7, 15, 0, 0, 0, 0, loc),
				End:         time.Date(2026, 7, 16, 0, 0, 0, 0, loc),
				AllDay:      true,
				AllDayStart: "2026-07-15",
				AllDayEnd:   "2026-07-16",
			},
		},
	}
	service := Service{
		Client:   client,
		Location: loc,
		Now: func() time.Time {
			return time.Date(2026, 7, 14, 12, 0, 0, 0, loc)
		},
	}

	got, err := service.Today(context.Background(), []string{"Work"})
	if err != nil {
		t.Fatalf("Today() error = %v", err)
	}
	var allDayTitles []string
	for _, event := range got.AllDayEvents {
		allDayTitles = append(allDayTitles, event.Title)
	}
	wantAllDay := []string{"spans today", "today all day", "today parsed as utc"}
	if !reflect.DeepEqual(allDayTitles, wantAllDay) {
		t.Fatalf("all-day event titles = %#v, want %#v", allDayTitles, wantAllDay)
	}
	var timedTitles []string
	for _, event := range got.Events {
		timedTitles = append(timedTitles, event.Title)
	}
	wantTimed := []string{"today timed"}
	if !reflect.DeepEqual(timedTitles, wantTimed) {
		t.Fatalf("timed event titles = %#v, want %#v", timedTitles, wantTimed)
	}
	if got.Events[0].Priority != 5 {
		t.Fatalf("priority = %d, want 5", got.Events[0].Priority)
	}
}

func TestTodayReportsMissingCalendars(t *testing.T) {
	service := Service{Client: &fakeClient{calendars: []caldav.Calendar{{Name: "Work"}}}}

	_, err := service.Today(context.Background(), []string{"Work", "Missing"})
	var missing MissingCalendarsError
	if !errors.As(err, &missing) {
		t.Fatalf("Today() error = %v, want MissingCalendarsError", err)
	}
	if !reflect.DeepEqual(missing.Names, []string{"Missing"}) {
		t.Fatalf("missing.Names = %#v, want Missing", missing.Names)
	}
}

func TestTodayRequiresCalendarNames(t *testing.T) {
	service := Service{Client: &fakeClient{calendars: []caldav.Calendar{{Name: "Work"}}}}

	_, err := service.Today(context.Background(), nil)
	var missing MissingCalendarsError
	if !errors.As(err, &missing) {
		t.Fatalf("Today() error = %v, want MissingCalendarsError", err)
	}
}

type fakeClient struct {
	calendars []caldav.Calendar
	events    map[string][]caldav.Event
	lastStart time.Time
	lastEnd   time.Time
}

func (f *fakeClient) Calendars(context.Context, []string) ([]caldav.Calendar, error) {
	return f.calendars, nil
}

func (f *fakeClient) Events(_ context.Context, cal caldav.Calendar, start, end time.Time, _ *time.Location) ([]caldav.Event, error) {
	f.lastStart = start
	f.lastEnd = end
	return f.events[cal.Name], nil
}
