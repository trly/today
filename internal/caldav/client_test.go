package caldav

import (
	"strings"
	"testing"
	"time"

	"github.com/emersion/go-ical"
)

func TestProjectEventPreservesAllDayDates(t *testing.T) {
	ve := ical.NewEvent()
	ve.Props.SetText(ical.PropSummary, "all day")
	ve.Props.SetDate(ical.PropDateTimeStart, time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC))
	ve.Props.SetDate(ical.PropDateTimeEnd, time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC))

	got, ok := projectEvent(*ve, time.FixedZone("test", -5*60*60))
	if !ok {
		t.Fatal("projectEvent() ok = false, want true")
	}
	if !got.AllDay {
		t.Fatal("projectEvent().AllDay = false, want true")
	}
	if got.AllDayStart != "2026-07-14" {
		t.Fatalf("projectEvent().AllDayStart = %q, want 2026-07-14", got.AllDayStart)
	}
	if got.AllDayEnd != "2026-07-15" {
		t.Fatalf("projectEvent().AllDayEnd = %q, want 2026-07-15", got.AllDayEnd)
	}
}

func TestProjectEventPreservesTitleAndPriority(t *testing.T) {
	ve := ical.NewEvent()
	ve.Props.SetText(ical.PropUID, "event-1")
	ve.Props.SetText(ical.PropSummary, "standup")
	ve.Props.SetDateTime(ical.PropDateTimeStart, time.Date(2026, 7, 14, 9, 0, 0, 0, time.UTC))
	priority := ical.NewProp(ical.PropPriority)
	priority.SetValueType(ical.ValueInt)
	priority.Value = "3"
	ve.Props.Set(priority)

	got, ok := projectEvent(*ve, time.UTC)
	if !ok {
		t.Fatal("projectEvent() ok = false, want true")
	}
	if got.Title != "standup" {
		t.Fatalf("projectEvent().Title = %q, want standup", got.Title)
	}
	if got.Priority != 3 {
		t.Fatalf("projectEvent().Priority = %d, want 3", got.Priority)
	}
}

func TestBuildCalendarQueryRequestsExpandedCalendarData(t *testing.T) {
	start := time.Date(2026, 7, 14, 0, 0, 0, 0, time.FixedZone("test", -5*60*60))
	end := start.AddDate(0, 0, 1)

	got := string(buildCalendarQuery(start, end))
	for _, want := range []string{
		`<calendar-data xmlns="urn:ietf:params:xml:ns:caldav">`,
		`<expand start="20260714T050000Z" end="20260715T050000Z"/>`,
		`<time-range start="20260714T050000Z" end="20260715T050000Z"/>`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("buildCalendarQuery() missing %s in:\n%s", want, got)
		}
	}
}
