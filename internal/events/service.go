// Package events contains application logic for listing calendars and fetching
// the current day's events.
package events

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/trly/today/internal/caldav"
)

// Client is the CalDAV behavior required by Service.
type Client interface {
	Calendars(ctx context.Context, names []string) ([]caldav.Calendar, error)
	Events(ctx context.Context, cal caldav.Calendar, start, end time.Time, loc *time.Location) ([]caldav.Event, error)
}

// Service fetches calendar metadata and events.
type Service struct {
	Client   Client
	Location *time.Location
	Now      func() time.Time
}

// Calendar is an API-safe projection of a CalDAV calendar.
type Calendar struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Path        string `json:"path"`
	Color       string `json:"color,omitempty"`
}

// Event is an API-safe projection of a CalDAV event.
type Event struct {
	UID           string     `json:"uid,omitempty"`
	Title         string     `json:"title"`
	Description   string     `json:"description,omitempty"`
	Start         time.Time  `json:"start"`
	End           *time.Time `json:"end,omitempty"`
	AllDay        bool       `json:"allDay"`
	StartDate     string     `json:"startDate,omitempty"`
	EndDate       string     `json:"endDate,omitempty"`
	Priority      int        `json:"priority,omitempty"`
	Calendar      string     `json:"calendar"`
	CalendarColor string     `json:"calendarColor,omitempty"`
}

// DailyEvents is the result for one local day.
type DailyEvents struct {
	Date         string  `json:"date"`
	AllDayEvents []Event `json:"allDayEvents"`
	Events       []Event `json:"events"`
}

// MissingCalendarsError reports requested calendars that were not available.
type MissingCalendarsError struct {
	Names []string
}

// InvalidDateError reports a date string that cannot be parsed as YYYY-MM-DD.
type InvalidDateError struct {
	Date string
}

func (e MissingCalendarsError) Error() string {
	return "missing calendars: " + strings.Join(e.Names, ", ")
}

func (e InvalidDateError) Error() string {
	return "invalid date: " + e.Date
}

// Calendars lists all calendars available to the configured credentials.
func (s Service) Calendars(ctx context.Context) ([]Calendar, error) {
	if s.Client == nil {
		return nil, fmt.Errorf("events: nil client")
	}
	calendars, err := s.Client.Calendars(ctx, nil)
	if err != nil {
		return nil, err
	}
	out := make([]Calendar, 0, len(calendars))
	for _, cal := range calendars {
		out = append(out, calendarProjection(cal))
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// Today returns events from the current local day for the requested calendars.
func (s Service) Today(ctx context.Context, names []string) (DailyEvents, error) {
	loc := s.Location
	if loc == nil {
		loc = time.Local
	}
	now := time.Now
	if s.Now != nil {
		now = s.Now
	}
	return s.eventsForDay(ctx, names, now().In(loc))
}

// Day returns events from the given YYYY-MM-DD local day for the requested calendars.
func (s Service) Day(ctx context.Context, names []string, date string) (DailyEvents, error) {
	loc := s.Location
	if loc == nil {
		loc = time.Local
	}
	day, err := time.ParseInLocation("2006-01-02", date, loc)
	if err != nil {
		return DailyEvents{}, InvalidDateError{Date: date}
	}
	return s.eventsForDay(ctx, names, day)
}

func (s Service) eventsForDay(ctx context.Context, names []string, day time.Time) (DailyEvents, error) {
	if s.Client == nil {
		return DailyEvents{}, fmt.Errorf("events: nil client")
	}
	names = cleanNames(names)
	if len(names) == 0 {
		return DailyEvents{}, MissingCalendarsError{}
	}
	loc := s.Location
	if loc == nil {
		loc = time.Local
	}
	start, end := dayRange(day.In(loc), loc)

	available, err := s.Client.Calendars(ctx, nil)
	if err != nil {
		return DailyEvents{}, err
	}
	selected, missing := selectCalendars(available, names)
	if len(missing) > 0 {
		return DailyEvents{}, MissingCalendarsError{Names: missing}
	}

	out := DailyEvents{Date: start.Format("2006-01-02")}
	for _, cal := range selected {
		calEvents, err := s.Client.Events(ctx, cal, start, end, loc)
		if err != nil {
			return DailyEvents{}, fmt.Errorf("events: fetch %q: %w", cal.Name, err)
		}
		for _, event := range eventsInRange(calEvents, start, end, loc) {
			projected := eventProjection(event, cal)
			if projected.AllDay {
				out.AllDayEvents = append(out.AllDayEvents, projected)
				continue
			}
			out.Events = append(out.Events, projected)
		}
	}
	sort.SliceStable(out.AllDayEvents, lessAllDay(out.AllDayEvents))
	sort.SliceStable(out.Events, lessTimed(out.Events))
	return out, nil
}

func dayRange(now time.Time, loc *time.Location) (time.Time, time.Time) {
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	return start, start.AddDate(0, 0, 1)
}

func eventsInRange(events []caldav.Event, start, end time.Time, loc *time.Location) []caldav.Event {
	out := make([]caldav.Event, 0, len(events))
	for _, event := range events {
		if eventInDay(event, start, end, loc) {
			out = append(out, event)
		}
	}
	return out
}

func eventInDay(event caldav.Event, start, end time.Time, loc *time.Location) bool {
	if event.AllDay {
		return allDayIncludesDate(event, start)
	}
	eventStart := event.Start.In(loc)
	return !eventStart.Before(start) && eventStart.Before(end)
}

func allDayIncludesDate(event caldav.Event, day time.Time) bool {
	startDate := event.AllDayStart
	if startDate == "" {
		startDate = event.Start.Format("2006-01-02")
	}
	endDate := event.AllDayEnd
	if endDate == "" {
		endDate = event.End.Format("2006-01-02")
	}
	if endDate == "" || endDate == "0001-01-01" {
		endDate = addDays(startDate, 1)
	}
	date := day.Format("2006-01-02")
	return date >= startDate && date < endDate
}

func cleanNames(names []string) []string {
	seen := make(map[string]struct{}, len(names))
	out := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

func selectCalendars(available []caldav.Calendar, names []string) ([]caldav.Calendar, []string) {
	byName := make(map[string]caldav.Calendar, len(available))
	for _, cal := range available {
		byName[cal.Name] = cal
	}
	selected := make([]caldav.Calendar, 0, len(names))
	var missing []string
	for _, name := range names {
		cal, ok := byName[name]
		if !ok {
			missing = append(missing, name)
			continue
		}
		selected = append(selected, cal)
	}
	return selected, missing
}

func calendarProjection(cal caldav.Calendar) Calendar {
	name := cal.Name
	if name == "" {
		name = cal.Path
	}
	return Calendar{Name: name, Description: cal.Description, Path: cal.Path, Color: cal.Color}
}

func eventProjection(e caldav.Event, cal caldav.Calendar) Event {
	var end *time.Time
	if !e.End.IsZero() {
		end = &e.End
	}
	return Event{
		UID:           e.UID,
		Title:         e.Title,
		Description:   e.Description,
		Start:         e.Start,
		End:           end,
		AllDay:        e.AllDay,
		StartDate:     e.AllDayStart,
		EndDate:       e.AllDayEnd,
		Priority:      e.Priority,
		Calendar:      cal.Name,
		CalendarColor: cal.Color,
	}
}

func lessAllDay(events []Event) func(i, j int) bool {
	return func(i, j int) bool {
		if events[i].Priority != events[j].Priority {
			return priorityLess(events[i].Priority, events[j].Priority)
		}
		return events[i].Title < events[j].Title
	}
}

func lessTimed(events []Event) func(i, j int) bool {
	return func(i, j int) bool {
		if !events[i].Start.Equal(events[j].Start) {
			return events[i].Start.Before(events[j].Start)
		}
		if events[i].Priority != events[j].Priority {
			return priorityLess(events[i].Priority, events[j].Priority)
		}
		return events[i].Title < events[j].Title
	}
}

func priorityLess(a, b int) bool {
	if a == 0 {
		return false
	}
	if b == 0 {
		return true
	}
	return a < b
}

func addDays(value string, days int) string {
	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		return ""
	}
	return date.AddDate(0, 0, days).Format("2006-01-02")
}
