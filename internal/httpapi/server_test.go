package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"

	todayv1 "github.com/trly/today/gen/today/v1"
	"github.com/trly/today/gen/today/v1/todayv1connect"
	"github.com/trly/today/internal/events"
)

func TestWebMissingPathReturnsNotFound(t *testing.T) {
	rec := httptest.NewRecorder()
	New(&fakeService{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/missing", nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestAPIDocs(t *testing.T) {
	rec := httptest.NewRecorder()
	New(&fakeService{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api-docs", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "text/markdown; charset=utf-8" {
		t.Fatalf("content type = %q, want text/markdown; charset=utf-8", got)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "HealthService") ||
		!strings.Contains(body, "ListEvents") ||
		!strings.Contains(body, "/today.v1.EventsService/ListEvents") ||
		!strings.Contains(body, "CodeInvalidArgument") {
		t.Fatalf("body missing API docs: %q", body)
	}
}

func TestHealth(t *testing.T) {
	server := newServer(t, &fakeService{})
	defer server.Close()
	client := todayv1connect.NewHealthServiceClient(server.Client(), server.URL)

	got, err := client.Health(context.Background(), &todayv1.HealthRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if got.GetStatus() != "ok" {
		t.Fatalf("status = %q, want ok", got.GetStatus())
	}
}

func TestCalendars(t *testing.T) {
	service := &fakeService{calendars: []events.Calendar{{Name: "Work", Path: "/work/", Color: "#c15a2a"}}}
	server := newServer(t, service)
	defer server.Close()
	client := todayv1connect.NewCalendarsServiceClient(server.Client(), server.URL)

	got, err := client.ListCalendars(context.Background(), &todayv1.ListCalendarsRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.GetCalendars()) != 1 || got.GetCalendars()[0].GetName() != "Work" || got.GetCalendars()[0].GetColor() != "#c15a2a" {
		t.Fatalf("calendars = %#v", got.GetCalendars())
	}
}

func TestEventsPassesRepeatedCalendarParams(t *testing.T) {
	service := &fakeService{daily: events.DailyEvents{
		Date: "2026-07-14",
		AllDayEvents: []events.Event{{
			UID:           "all-day-1",
			Title:         "company retreat",
			Description:   "Bring badge",
			StartDate:     "2026-07-14",
			EndDate:       "2026-07-15",
			Priority:      3,
			Calendar:      "Work",
			CalendarColor: "#c15a2a",
		}},
		Events: []events.Event{{
			UID:           "standup-1",
			Title:         "standup",
			Description:   "Daily sync",
			Start:         time.Date(2026, 7, 14, 9, 0, 0, 0, time.UTC),
			End:           timePtr(time.Date(2026, 7, 14, 9, 30, 0, 0, time.UTC)),
			Priority:      5,
			Calendar:      "Work",
			CalendarColor: "#c15a2a",
		}},
	}}
	server := newServer(t, service)
	defer server.Close()
	client := todayv1connect.NewEventsServiceClient(server.Client(), server.URL)

	got, err := client.ListEvents(context.Background(), &todayv1.ListEventsRequest{Calendar: []string{"Work", "Home"}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(service.names, []string{"Work", "Home"}) {
		t.Fatalf("names = %#v, want repeated calendar values", service.names)
	}
	if len(got.GetAllDayEvents()) != 1 || got.GetAllDayEvents()[0].GetId() != "all-day-1" || got.GetAllDayEvents()[0].GetMeta() != "Work" {
		t.Fatalf("all-day events = %#v", got.GetAllDayEvents())
	}
	if got.GetAllDayEvents()[0].GetDescription() != "Bring badge" || got.GetAllDayEvents()[0].GetCalendarColor() != "#c15a2a" {
		t.Fatalf("all-day event details = %#v", got.GetAllDayEvents()[0])
	}
	if len(got.GetEvents()) != 1 || got.GetEvents()[0].GetId() != "standup-1" || got.GetEvents()[0].GetTitle() != "standup" || got.GetEvents()[0].GetCalendarColor() != "#c15a2a" {
		t.Fatalf("events = %#v", got.GetEvents())
	}
	if got.GetEvents()[0].GetTime() != "9:00 AM" || got.GetEvents()[0].GetStartMinutes() != 9*60 || got.GetEvents()[0].GetDurationMinutes() != 30 || got.GetEvents()[0].GetNote() != "Daily sync" {
		t.Fatalf("timed event = %#v", got.GetEvents()[0])
	}
}

func TestEventsPassesDateParam(t *testing.T) {
	service := &fakeService{daily: events.DailyEvents{Date: "2026-07-15"}}
	server := newServer(t, service)
	defer server.Close()
	client := todayv1connect.NewEventsServiceClient(server.Client(), server.URL)

	_, err := client.ListEvents(context.Background(), &todayv1.ListEventsRequest{Calendar: []string{"Work"}, Date: "2026-07-15"})
	if err != nil {
		t.Fatal(err)
	}
	if service.date != "2026-07-15" {
		t.Fatalf("date = %q, want 2026-07-15", service.date)
	}
}

func TestEventsInvalidDateParamReturnsInvalidArgument(t *testing.T) {
	service := &fakeService{}
	server := newServer(t, service)
	defer server.Close()
	client := todayv1connect.NewEventsServiceClient(server.Client(), server.URL)

	_, err := client.ListEvents(context.Background(), &todayv1.ListEventsRequest{Calendar: []string{"Work"}, Date: "tomorrow"})
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("code = %v, want %v", connect.CodeOf(err), connect.CodeInvalidArgument)
	}
}

func TestEventsSemanticallyInvalidDateParamReturnsInvalidArgument(t *testing.T) {
	service := &fakeService{err: events.InvalidDateError{Date: "2026-99-99"}}
	server := newServer(t, service)
	defer server.Close()
	client := todayv1connect.NewEventsServiceClient(server.Client(), server.URL)

	_, err := client.ListEvents(context.Background(), &todayv1.ListEventsRequest{Calendar: []string{"Work"}, Date: "2026-99-99"})
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("code = %v, want %v", connect.CodeOf(err), connect.CodeInvalidArgument)
	}
}

func TestEventsMissingCalendarsReturnsInvalidArgument(t *testing.T) {
	service := &fakeService{err: events.MissingCalendarsError{Names: []string{"Missing"}}}
	server := newServer(t, service)
	defer server.Close()
	client := todayv1connect.NewEventsServiceClient(server.Client(), server.URL)

	_, err := client.ListEvents(context.Background(), &todayv1.ListEventsRequest{Calendar: []string{"Missing"}})
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Fatalf("code = %v, want %v", connect.CodeOf(err), connect.CodeInvalidArgument)
	}
}

func TestEventsServiceFailureReturnsUnavailable(t *testing.T) {
	service := &fakeService{err: errors.New("boom")}
	server := newServer(t, service)
	defer server.Close()
	client := todayv1connect.NewEventsServiceClient(server.Client(), server.URL)

	_, err := client.ListEvents(context.Background(), &todayv1.ListEventsRequest{Calendar: []string{"Work"}})
	if connect.CodeOf(err) != connect.CodeUnavailable {
		t.Fatalf("code = %v, want %v", connect.CodeOf(err), connect.CodeUnavailable)
	}
}

func newServer(t *testing.T, service *fakeService) *httptest.Server {
	t.Helper()
	return httptest.NewServer(New(service))
}

func timePtr(t time.Time) *time.Time {
	return &t
}

type fakeService struct {
	calendars []events.Calendar
	daily     events.DailyEvents
	err       error
	names     []string
	date      string
}

func (f *fakeService) Calendars(context.Context) ([]events.Calendar, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.calendars, nil
}

func (f *fakeService) Today(_ context.Context, names []string) (events.DailyEvents, error) {
	f.names = names
	if f.err != nil {
		return events.DailyEvents{}, f.err
	}
	return f.daily, nil
}

func (f *fakeService) Day(_ context.Context, names []string, date string) (events.DailyEvents, error) {
	f.names = names
	f.date = date
	if f.err != nil {
		return events.DailyEvents{}, f.err
	}
	return f.daily, nil
}
