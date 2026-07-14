import { createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { CalendarsService, EventsService } from '../gen/today/v1/today_pb';

const transport = createConnectTransport({
	baseUrl: import.meta.env.VITE_API_URL ?? window.location.origin
});

const calendarsClient = createClient(CalendarsService, transport);
const eventsClient = createClient(EventsService, transport);
let calendarsPromise;

const listCalendars = () => {
	if (!calendarsPromise) {
		calendarsPromise = calendarsClient.listCalendars({}).catch((error) => {
			calendarsPromise = undefined;
			throw error;
		});
	}

	return calendarsPromise;
};

const formatLocalDate = (date) => {
	const year = date.getFullYear();
	const month = String(date.getMonth() + 1).padStart(2, '0');
	const day = String(date.getDate()).padStart(2, '0');

	return `${year}-${month}-${day}`;
};

export async function listDayEvents(date, options) {
	const { calendars } = await listCalendars();
	const calendar = calendars.map(({ name }) => name);
	const calendarColorByName = new Map(
		calendars.map(({ name, color }) => [name, color]).filter(([, color]) => color)
	);

	if (calendar.length === 0) {
		return { date: formatLocalDate(date), allDayEvents: [], events: [] };
	}

	const response = await eventsClient.listEvents(
		{
			calendar,
			date: formatLocalDate(date)
		},
		options
	);

	return {
		...response,
		allDayEvents: response.allDayEvents.map((event) => ({
			...event,
			calendarColor: calendarColorByName.get(event.calendar) || event.calendarColor
		})),
		events: response.events.map((event) => ({
			...event,
			calendarColor: calendarColorByName.get(event.calendar) || event.calendarColor
		}))
	};
}
