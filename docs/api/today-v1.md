# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [today/v1/today.proto](#today_v1_today-proto)
    - [AllDayEvent](#today-v1-AllDayEvent)
    - [Calendar](#today-v1-Calendar)
    - [HealthRequest](#today-v1-HealthRequest)
    - [HealthResponse](#today-v1-HealthResponse)
    - [ListCalendarsRequest](#today-v1-ListCalendarsRequest)
    - [ListCalendarsResponse](#today-v1-ListCalendarsResponse)
    - [ListEventsRequest](#today-v1-ListEventsRequest)
    - [ListEventsResponse](#today-v1-ListEventsResponse)
    - [TimedEvent](#today-v1-TimedEvent)
  
    - [CalendarsService](#today-v1-CalendarsService)
    - [EventsService](#today-v1-EventsService)
    - [HealthService](#today-v1-HealthService)
  
- [Scalar Value Types](#scalar-value-types)



<a name="today_v1_today-proto"></a>
<p align="right"><a href="#top">Top</a></p>

## today/v1/today.proto



<a name="today-v1-AllDayEvent"></a>

### AllDayEvent
AllDayEvent describes one all-day event ready for display by the web UI.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  | Stable event identifier for keyed rendering. |
| title | [string](#string) |  | Short display title. |
| meta | [string](#string) |  | Secondary label shown under the title. The service uses the calendar name. |
| description | [string](#string) |  | Full event description or notes, when available. |
| calendar | [string](#string) |  | Calendar display name this event came from. |
| calendar_color | [string](#string) |  | Calendar color this event came from, usually a CSS hex color. |
| priority | [int32](#int32) |  | iCalendar PRIORITY. 0 means unset; 1 is highest and 9 is lowest. |
| start_date | [string](#string) |  | Local start date in YYYY-MM-DD form. |
| end_date | [string](#string) |  | Local end date in YYYY-MM-DD form. For all-day events this is exclusive. |






<a name="today-v1-Calendar"></a>

### Calendar
Calendar describes one queryable calendar collection.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  | Human-readable calendar display name. |
| description | [string](#string) |  | Provider-supplied calendar description. |
| path | [string](#string) |  | CalDAV collection path for diagnostics and stable identification. |
| color | [string](#string) |  | Calendar color as supplied by the provider, usually a CSS hex color. |






<a name="today-v1-HealthRequest"></a>

### HealthRequest
HealthRequest requests the server health status.






<a name="today-v1-HealthResponse"></a>

### HealthResponse
HealthResponse describes server availability.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| status | [string](#string) |  | Machine-readable health state. |






<a name="today-v1-ListCalendarsRequest"></a>

### ListCalendarsRequest
ListCalendarsRequest requests all available calendars.






<a name="today-v1-ListCalendarsResponse"></a>

### ListCalendarsResponse
ListCalendarsResponse contains available calendars.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| calendars | [Calendar](#today-v1-Calendar) | repeated | Calendars available to query. |






<a name="today-v1-ListEventsRequest"></a>

### ListEventsRequest
ListEventsRequest selects calendars and date for an event query.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| calendar | [string](#string) | repeated | Calendar display names to include. Names are trimmed and de-duplicated by the service. At least one non-empty calendar is required. |
| date | [string](#string) |  | Local date to query in YYYY-MM-DD form. Empty means today in the server&#39;s configured location. Values that do not parse as a real local date return CodeInvalidArgument. |






<a name="today-v1-ListEventsResponse"></a>

### ListEventsResponse
ListEventsResponse contains all-day and timed events for one date.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| date | [string](#string) |  | Local date represented by this response in YYYY-MM-DD form. |
| all_day_events | [AllDayEvent](#today-v1-AllDayEvent) | repeated | All-day events active on date. Multi-day all-day events are included for each covered date. Sorted by priority, then title. |
| events | [TimedEvent](#today-v1-TimedEvent) | repeated | Timed events whose local start date equals date. Sorted by start time, then priority, then title. |






<a name="today-v1-TimedEvent"></a>

### TimedEvent
TimedEvent describes one timed event ready for display by the web UI.


| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| id | [string](#string) |  | Stable event identifier for keyed rendering. |
| time | [string](#string) |  | Local display time, formatted for the agenda event label. |
| title | [string](#string) |  | Short display title. |
| note | [string](#string) |  | Notes shown under the title, when available. |
| start_minutes | [int32](#int32) |  | Event start offset from local midnight, in minutes. |
| duration_minutes | [int32](#int32) |  | Event duration in minutes. Values are always at least 1. |
| calendar | [string](#string) |  | Calendar display name this event came from. |
| calendar_color | [string](#string) |  | Calendar color this event came from, usually a CSS hex color. |
| priority | [int32](#int32) |  | iCalendar PRIORITY. 0 means unset; 1 is highest and 9 is lowest. |





 

 

 


<a name="today-v1-CalendarsService"></a>

### CalendarsService
CalendarsService exposes configured calendar collections.

Connect RPC path: /today.v1.CalendarsService/ListCalendars.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ListCalendars | [ListCalendarsRequest](#today-v1-ListCalendarsRequest) | [ListCalendarsResponse](#today-v1-ListCalendarsResponse) | Lists calendars available to the service. Calendars are sorted by display name. Backend calendar discovery failures return CodeUnavailable. |


<a name="today-v1-EventsService"></a>

### EventsService
EventsService exposes calendar events.

Connect RPC path: /today.v1.EventsService/ListEvents.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| ListEvents | [ListEventsRequest](#today-v1-ListEventsRequest) | [ListEventsResponse](#today-v1-ListEventsResponse) | Lists events for selected calendars on a local date. At least one calendar display name is required. Missing calendars or invalid dates return CodeInvalidArgument. Backend calendar fetch failures return CodeUnavailable. |


<a name="today-v1-HealthService"></a>

### HealthService
HealthService reports server availability.

Connect RPC path: /today.v1.HealthService/Health.

| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| Health | [HealthRequest](#today-v1-HealthRequest) | [HealthResponse](#today-v1-HealthResponse) | Returns the current server health status. The server returns status &#34;ok&#34; when the API process is reachable. |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

