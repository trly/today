// Package caldav provides a thin client wrapper around
// github.com/emersion/go-webdav/caldav for discovering calendars by name and
// fetching events in a time range as structured Events.
package caldav

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/emersion/go-ical"
	"github.com/emersion/go-webdav"
	"github.com/emersion/go-webdav/caldav"
)

// Client wraps a caldav.Client with credentials baked in. The underlying
// HTTP client and endpoint URL are retained so Events can issue a custom
// REPORT (see queryCalendar) that works around servers which mishandle the
// nested calendar-data comp request that go-webdav emits.
type Client struct {
	caldav     *caldav.Client
	httpClient webdav.HTTPClient
	endpoint   string
}

// Calendar is the calendar metadata needed by the app.
type Calendar struct {
	Path                  string
	Name                  string
	Description           string
	Color                 string
	MaxResourceSize       int64
	SupportedComponentSet []string
}

// Event is a minimal projection of a VEVENT component.
type Event struct {
	UID         string
	Title       string
	Description string
	Location    string
	Start       time.Time
	End         time.Time
	AllDay      bool
	AllDayStart string
	AllDayEnd   string
	Priority    int
}

// NewClient returns a CalDAV client authenticated via HTTP basic auth.
// An empty username produces an unauthenticated client.
func NewClient(endpoint, username, password string) (*Client, error) {
	var httpClient webdav.HTTPClient = http.DefaultClient
	if username != "" {
		httpClient = webdav.HTTPClientWithBasicAuth(httpClient, username, password)
	}
	c, err := caldav.NewClient(httpClient, endpoint)
	if err != nil {
		return nil, fmt.Errorf("caldav: new client: %w", err)
	}
	return &Client{caldav: c, httpClient: httpClient, endpoint: endpoint}, nil
}

// Calendars returns the user's calendars whose Name matches any of the provided
// names. If names is empty, all calendars are returned.
func (c *Client) Calendars(ctx context.Context, names []string) ([]Calendar, error) {
	principal, err := c.caldav.FindCurrentUserPrincipal(ctx)
	if err != nil {
		return nil, fmt.Errorf("caldav: find current user principal: %w", err)
	}
	homeSet, err := c.caldav.FindCalendarHomeSet(ctx, principal)
	if err != nil {
		return nil, fmt.Errorf("caldav: find calendar home set: %w", err)
	}
	all, err := c.caldav.FindCalendars(ctx, homeSet)
	if err != nil {
		return nil, fmt.Errorf("caldav: find calendars: %w", err)
	}
	colors := c.calendarColors(ctx, homeSet)
	calendars := make([]Calendar, 0, len(all))
	for _, cal := range all {
		calendars = append(calendars, Calendar{
			Path:                  cal.Path,
			Name:                  cal.Name,
			Description:           cal.Description,
			Color:                 colors[normalizeCalDAVPath(cal.Path)],
			MaxResourceSize:       cal.MaxResourceSize,
			SupportedComponentSet: cal.SupportedComponentSet,
		})
	}
	if len(names) == 0 {
		return calendars, nil
	}
	want := make(map[string]struct{}, len(names))
	for _, n := range names {
		want[strings.TrimSpace(n)] = struct{}{}
	}
	var matched []Calendar
	for _, cal := range calendars {
		if _, ok := want[cal.Name]; ok {
			matched = append(matched, cal)
		}
	}
	if len(matched) == 0 {
		return nil, fmt.Errorf("caldav: no calendars matched names %v", names)
	}
	return matched, nil
}

func (c *Client) calendarColors(ctx context.Context, calendarHomeSet string) map[string]string {
	reqURL, err := resolveURL(c.endpoint, calendarHomeSet)
	if err != nil {
		return nil
	}
	body := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<propfind xmlns="DAV:">
 <prop>
  <calendar-color xmlns="http://apple.com/ns/ical/"/>
 </prop>
</propfind>`)
	req, err := http.NewRequestWithContext(ctx, "PROPFIND", reqURL, bytes.NewReader(body))
	if err != nil {
		return nil
	}
	req.Header.Set("Content-Type", "application/xml; charset=utf-8")
	req.Header.Set("Depth", "1")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMultiStatus {
		return nil
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}
	var ms colorMultistatus
	if err := xml.Unmarshal(data, &ms); err != nil {
		return nil
	}
	colors := make(map[string]string, len(ms.Responses))
	for _, r := range ms.Responses {
		for _, ps := range r.Propstats {
			if !strings.Contains(ps.Status, " 200 ") && !strings.HasSuffix(strings.TrimSpace(ps.Status), "200 OK") {
				continue
			}
			color := strings.TrimSpace(ps.Prop.CalendarColor)
			if color == "" {
				continue
			}
			colors[normalizeCalDAVPath(r.Href)] = color
		}
	}
	return colors
}

type colorMultistatus struct {
	XMLName   xml.Name          `xml:"DAV: multistatus"`
	Responses []colorMSResponse `xml:"response"`
}

type colorMSResponse struct {
	Href      string            `xml:"href"`
	Propstats []colorMSPropstat `xml:"propstat"`
}

type colorMSPropstat struct {
	Prop   colorMSProp `xml:"prop"`
	Status string      `xml:"status"`
}

type colorMSProp struct {
	CalendarColor string `xml:"http://apple.com/ns/ical/ calendar-color"`
}

func normalizeCalDAVPath(value string) string {
	if value == "" {
		return ""
	}
	if u, err := url.Parse(value); err == nil && u.Path != "" {
		value = u.Path
	}
	cleaned := path.Clean("/" + strings.TrimSpace(value))
	if strings.HasSuffix(value, "/") && !strings.HasSuffix(cleaned, "/") {
		cleaned += "/"
	}
	return cleaned
}

// Events returns the VEVENTs in cal whose start time falls within [start,
// end). A zero end means open-ended. Times are parsed with loc as the
// fallback location for floating times; the TZID parameter is honored when
// present. Malformed events are skipped rather than failing the whole call,
// so a single bad event doesn't discard the rest of the calendar.
func (c *Client) Events(ctx context.Context, cal Calendar, start, end time.Time, loc *time.Location) ([]Event, error) {
	if loc == nil {
		loc = time.Local
	}
	objs, err := c.queryCalendar(ctx, cal, start, end)
	if err != nil {
		return nil, err
	}
	out := make([]Event, 0, len(objs))
	for _, cal := range objs {
		for _, ve := range cal.Events() {
			e, ok := projectEvent(ve, loc)
			if !ok {
				continue
			}
			out = append(out, e)
		}
	}
	return out, nil
}

// queryCalendar issues a CalDAV calendar-query REPORT requesting the bare
// calendar-data property. Some servers (notably iCloud) return empty VEVENT
// shells when calendar-data is requested with a nested comp/allprop
// structure, so we request the full iCalendar payload and parse it
// ourselves. The time-range filter is expressed in UTC per RFC 4791.
func (c *Client) queryCalendar(ctx context.Context, cal Calendar, start, end time.Time) ([]*ical.Calendar, error) {
	calURL, err := resolveURL(c.endpoint, cal.Path)
	if err != nil {
		return nil, fmt.Errorf("caldav: resolve calendar URL: %w", err)
	}
	body := buildCalendarQuery(start, end)

	req, err := http.NewRequestWithContext(ctx, "REPORT", calURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("caldav: build report request: %w", err)
	}
	req.Header.Set("Content-Type", "application/xml; charset=utf-8")
	req.Header.Set("Depth", "1")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("caldav: query calendar %q: %w", cal.Name, err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("caldav: read calendar %q response: %w", cal.Name, err)
	}
	if resp.StatusCode != http.StatusMultiStatus {
		return nil, fmt.Errorf("caldav: query calendar %q: unexpected status %d: %s", cal.Name, resp.StatusCode, truncate(string(data), 200))
	}

	var ms multistatus
	if err := xml.Unmarshal(data, &ms); err != nil {
		return nil, fmt.Errorf("caldav: decode calendar %q multistatus: %w", cal.Name, err)
	}

	cals := make([]*ical.Calendar, 0, len(ms.Responses))
	for _, r := range ms.Responses {
		for _, ps := range r.Propstats {
			if !strings.Contains(ps.Status, " 200 ") && !strings.HasSuffix(strings.TrimSpace(ps.Status), "200 OK") {
				continue
			}
			if len(ps.Prop.CalendarData) == 0 {
				continue
			}
			decoded, err := ical.NewDecoder(bytes.NewReader(ps.Prop.CalendarData)).Decode()
			if err != nil {
				continue
			}
			cals = append(cals, decoded)
		}
	}
	return cals, nil
}

// buildCalendarQuery returns the REPORT request body for a VEVENT time-range
// query. The timestamps are formatted as UTC "date with UTC time" per RFC
// 5545; they contain only digits and the letters T/Z, so string assembly is
// safe here.
func buildCalendarQuery(start, end time.Time) []byte {
	const layout = "20060102T150405Z"
	startAttr := ""
	if !start.IsZero() {
		startAttr = ` start="` + start.UTC().Format(layout) + `"`
	}
	endAttr := ""
	if !end.IsZero() {
		endAttr = ` end="` + end.UTC().Format(layout) + `"`
	}
	return []byte(`<?xml version="1.0" encoding="UTF-8"?>
<calendar-query xmlns="urn:ietf:params:xml:ns:caldav">
 <prop xmlns="DAV:">
  <calendar-data xmlns="urn:ietf:params:xml:ns:caldav">
   <expand` + startAttr + endAttr + `/>
  </calendar-data>
 </prop>
 <filter xmlns="urn:ietf:params:xml:ns:caldav">
  <comp-filter name="VCALENDAR">
   <comp-filter name="VEVENT">
    <time-range` + startAttr + endAttr + `/>
   </comp-filter>
  </comp-filter>
 </filter>
</calendar-query>`)
}

// resolveURL joins a CalDAV endpoint with a calendar path, which may be
// either an absolute path or a full URL.
func resolveURL(endpoint, path string) (string, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path, nil
	}
	base, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	ref, err := url.Parse(path)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(ref).String(), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

// multistatus models the DAV multistatus response body.
type multistatus struct {
	XMLName   xml.Name     `xml:"DAV: multistatus"`
	Responses []msResponse `xml:"response"`
}

type msResponse struct {
	Href      string       `xml:"href"`
	Propstats []msPropstat `xml:"propstat"`
}

type msPropstat struct {
	Prop   msProp `xml:"prop"`
	Status string `xml:"status"`
}

type msProp struct {
	CalendarData []byte `xml:"urn:ietf:params:xml:ns:caldav calendar-data"`
}

// projectEvent converts an ical.Event into our Event projection. Returns
// ok=false to indicate the event should be skipped (missing or unparseable
// start).
func projectEvent(ve ical.Event, loc *time.Location) (Event, bool) {
	e := Event{}
	if p := ve.Props.Get(ical.PropUID); p != nil {
		e.UID = p.Value
	}
	if t, err := ve.Props.Text(ical.PropSummary); err == nil {
		e.Title = t
	}
	if t, err := ve.Props.Text(ical.PropDescription); err == nil {
		e.Description = t
	}
	if t, err := ve.Props.Text(ical.PropLocation); err == nil {
		e.Location = t
	}
	if p := ve.Props.Get(ical.PropPriority); p != nil {
		if priority, err := p.Int(); err == nil {
			e.Priority = priority
		}
	}
	if p := ve.Props.Get(ical.PropDateTimeStart); p != nil {
		e.AllDay = p.ValueType() == ical.ValueDate
		if e.AllDay {
			e.AllDayStart = formatICalDate(p.Value)
		}
	}
	start, err := ve.DateTimeStart(loc)
	if err != nil {
		return Event{}, false
	}
	if start.IsZero() {
		return Event{}, false
	}
	e.Start = start
	if end, err := ve.DateTimeEnd(loc); err == nil {
		e.End = end
	}
	if e.AllDay {
		if p := ve.Props.Get(ical.PropDateTimeEnd); p != nil {
			e.AllDayEnd = formatICalDate(p.Value)
		}
		if e.AllDayEnd == "" && e.AllDayStart != "" {
			e.AllDayEnd = addDays(e.AllDayStart, 1)
		}
	}
	return e, true
}

func formatICalDate(value string) string {
	date, err := time.Parse("20060102", value)
	if err != nil {
		return ""
	}
	return date.Format("2006-01-02")
}

func addDays(value string, days int) string {
	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		return ""
	}
	return date.AddDate(0, 0, days).Format("2006-01-02")
}
