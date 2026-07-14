// Package httpapi contains the HTTP handlers for the today API.
package httpapi

import (
	"context"
	"errors"
	"net/http"
	"time"

	"connectrpc.com/connect"

	apidocs "github.com/trly/today/docs/api"
	todayv1 "github.com/trly/today/gen/today/v1"
	"github.com/trly/today/gen/today/v1/todayv1connect"
	"github.com/trly/today/internal/events"
)

// Service is the application behavior required by the HTTP API.
type Service interface {
	Calendars(ctx context.Context) ([]events.Calendar, error)
	Day(ctx context.Context, names []string, date string) (events.DailyEvents, error)
	Today(ctx context.Context, names []string) (events.DailyEvents, error)
}

// Handler serves the HTTP API.
type Handler struct {
	Service Service
}

// New returns an HTTP handler with all routes registered.
func New(service Service) http.Handler {
	h := Handler{Service: service}
	mux := http.NewServeMux()
	healthPath, healthHandler := todayv1connect.NewHealthServiceHandler(h)
	mux.Handle(healthPath, healthHandler)
	calendarsPath, calendarsHandler := todayv1connect.NewCalendarsServiceHandler(h)
	mux.Handle(calendarsPath, calendarsHandler)
	eventsPath, eventsHandler := todayv1connect.NewEventsServiceHandler(h)
	mux.Handle(eventsPath, eventsHandler)
	mux.Handle("/api-docs", apidocs.Handler())
	return mux
}

func (h Handler) Health(context.Context, *todayv1.HealthRequest) (*todayv1.HealthResponse, error) {
	return &todayv1.HealthResponse{Status: "ok"}, nil
}

func (h Handler) ListCalendars(ctx context.Context, _ *todayv1.ListCalendarsRequest) (*todayv1.ListCalendarsResponse, error) {
	calendars, err := h.Service.Calendars(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("failed to list calendars"))
	}
	return &todayv1.ListCalendarsResponse{Calendars: calendarsResponse(calendars)}, nil
}

func (h Handler) ListEvents(ctx context.Context, req *todayv1.ListEventsRequest) (*todayv1.ListEventsResponse, error) {
	names := req.GetCalendar()
	var result events.DailyEvents
	var err error
	if date := req.GetDate(); date != "" {
		if !validDate(date) {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("date must use YYYY-MM-DD"))
		}
		result, err = h.Service.Day(ctx, names, date)
	} else {
		result, err = h.Service.Today(ctx, names)
	}
	if err != nil {
		var missing events.MissingCalendarsError
		if errors.As(err, &missing) {
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		}
		var invalidDate events.InvalidDateError
		if errors.As(err, &invalidDate) {
			return nil, connect.NewError(connect.CodeInvalidArgument, errors.New("date must use YYYY-MM-DD"))
		}
		return nil, connect.NewError(connect.CodeUnavailable, errors.New("failed to fetch events"))
	}
	return eventsResponse(result), nil
}

func validDate(date string) bool {
	if len(date) != len("2006-01-02") {
		return false
	}
	for i, r := range date {
		if i == 4 || i == 7 {
			if r != '-' {
				return false
			}
			continue
		}
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func calendarsResponse(calendars []events.Calendar) []*todayv1.Calendar {
	out := make([]*todayv1.Calendar, 0, len(calendars))
	for _, cal := range calendars {
		out = append(out, &todayv1.Calendar{
			Name:        cal.Name,
			Description: cal.Description,
			Path:        cal.Path,
			Color:       cal.Color,
		})
	}
	return out
}

func eventsResponse(daily events.DailyEvents) *todayv1.ListEventsResponse {
	return &todayv1.ListEventsResponse{
		Date:         daily.Date,
		AllDayEvents: allDayEventResponses(daily.AllDayEvents),
		Events:       timedEventResponses(daily.Events),
	}
}

func allDayEventResponses(in []events.Event) []*todayv1.AllDayEvent {
	out := make([]*todayv1.AllDayEvent, 0, len(in))
	for _, event := range in {
		out = append(out, &todayv1.AllDayEvent{
			Id:            event.UID,
			Title:         event.Title,
			Meta:          event.Calendar,
			Description:   event.Description,
			Calendar:      event.Calendar,
			CalendarColor: event.CalendarColor,
			Priority:      int32(event.Priority),
			StartDate:     event.StartDate,
			EndDate:       event.EndDate,
		})
	}
	return out
}

func timedEventResponses(in []events.Event) []*todayv1.TimedEvent {
	out := make([]*todayv1.TimedEvent, 0, len(in))
	for _, event := range in {
		out = append(out, &todayv1.TimedEvent{
			Id:              event.UID,
			Time:            event.Start.Format("3:04 PM"),
			Title:           event.Title,
			Note:            event.Description,
			StartMinutes:    int32(event.Start.Hour()*60 + event.Start.Minute()),
			DurationMinutes: int32(eventDurationMinutes(event)),
			Calendar:        event.Calendar,
			CalendarColor:   event.CalendarColor,
			Priority:        int32(event.Priority),
		})
	}
	return out
}

func eventDurationMinutes(event events.Event) int {
	if event.End == nil {
		return 30
	}
	duration := int(event.End.Sub(event.Start) / time.Minute)
	if duration < 1 {
		return 1
	}
	return duration
}
