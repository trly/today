<script>
	import { timeFormatter } from './timeFormat.js';

	let { event, date } = $props();

	let startRow = $derived(Math.floor(event.startMinutes / 30) + 1);
	let span = $derived(Math.max(1, Math.ceil(event.durationMinutes / 30)));
	let lane = $derived(event.agendaLane ?? 0);
	let laneCount = $derived(event.agendaLaneCount ?? 1);
	let laneWidth = $derived(laneCount <= 2 ? 1 / laneCount : 1 / (1 + (laneCount - 1) * 0.4));
	let laneOffset = $derived(laneCount <= 1 ? 0 : ((1 - laneWidth) / (laneCount - 1)) * lane);
	let calendarColor = $derived(event.calendarColor || event.calendar_color || 'var(--color-moss)');
	let calendarName = $derived(event.calendar || event.calendarName || event.calendar_name);
	let hasConflict = $derived((event.conflictCount ?? 0) > 0);

	const formatMinutes = (totalMinutes) => {
		const normalizedMinutes = totalMinutes % (24 * 60);
		const hours = Math.floor(normalizedMinutes / 60);
		const minutes = normalizedMinutes % 60;

		return timeFormatter.format(new Date(2000, 0, 1, hours, minutes));
	};
	const formatDateTime = (totalMinutes) => {
		const dayOffset = Math.floor(totalMinutes / (24 * 60));
		const minutesInDay = ((totalMinutes % (24 * 60)) + 24 * 60) % (24 * 60);
		const eventDate = new Date(date.getFullYear(), date.getMonth(), date.getDate() + dayOffset);
		const year = eventDate.getFullYear();
		const month = String(eventDate.getMonth() + 1).padStart(2, '0');
		const day = String(eventDate.getDate()).padStart(2, '0');
		const hours = String(Math.floor(minutesInDay / 60)).padStart(2, '0');
		const minutes = String(minutesInDay % 60).padStart(2, '0');

		return `${year}-${month}-${day}T${hours}:${minutes}`;
	};

	let endMinutes = $derived(event.startMinutes + Math.max(1, event.durationMinutes));
	let endTime = $derived(
		event.endTime || event.end_time || formatMinutes(endMinutes)
	);
	let startDateTime = $derived(formatDateTime(event.startMinutes));
	let endDateTime = $derived(formatDateTime(endMinutes));
	let accessibleLabel = $derived(
		`${event.title}, ${event.time} to ${endTime}${calendarName ? `, ${calendarName}` : ''}${
			event.conflictLabel ? `, ${event.conflictLabel}` : hasConflict ? ', conflict' : ''
		}`
	);
</script>

<article
	class="time-block"
	aria-label={accessibleLabel}
	style:--start={startRow}
	style:--span={span}
	style:--lane-width={laneWidth}
	style:--lane-offset={laneOffset}
	style:--calendar-color={calendarColor}
>
	<div class="event-content">
		<span class="event-accent" aria-hidden="true"></span>
		<div class="event-meta">
			<span class="event-time">
				<time datetime={startDateTime}>{event.time}</time> -
				<time datetime={endDateTime}>{endTime}</time>
			</span>
			{#if calendarName}
				<span class="calendar-name">{calendarName}</span>
			{/if}
		</div>
		<h3>{event.title}</h3>
	</div>
</article>

<style>
	.time-block {
		grid-row: var(--start) / span var(--span);
		box-sizing: border-box;
		width: calc((100% - 0.8rem) * var(--lane-width));
		margin-left: calc(0.8rem + (100% - 0.8rem) * var(--lane-offset));
		overflow: hidden;
		border: 1px solid color-mix(in oklch, var(--calendar-color, var(--color-moss)) 72%, var(--color-ink) 8%);
		border-radius: var(--radius-event);
		background: linear-gradient(
			135deg,
			var(--color-surface-raised),
			color-mix(in oklch, var(--calendar-color, var(--color-moss)) 14%, transparent)
		);
		box-shadow: var(--shadow-event);
		transition:
			border-color var(--duration-state) var(--ease-out-quart),
			box-shadow var(--duration-state) var(--ease-out-quart);
		container: agenda-event / size;
	}

	.event-content {
		display: grid;
		align-content: start;
		gap: 0.2rem;
		box-sizing: border-box;
		height: 100%;
		min-width: 0;
		padding: 0.75rem 0.85rem;
	}

	.event-accent {
		width: 2.4rem;
		height: 0.22rem;
		flex: none;
		border-radius: var(--radius-pill);
		background: var(--calendar-color, var(--color-moss));
	}

	.event-meta {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		align-items: center;
		gap: 0.45rem;
		min-width: 0;
		overflow: hidden;
	}

	.event-time,
	.event-meta span {
		font-size: var(--text-label);
		line-height: var(--leading-label);
		font-weight: 800;
		letter-spacing: var(--tracking-wide);
		text-transform: uppercase;
		font-variant-numeric: tabular-nums;
		color: var(--color-moss-deep);
	}

	.calendar-name {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		color: var(--color-ink-muted);
	}

	h3 {
		margin: 0;
		overflow-wrap: anywhere;
		font-size: var(--text-title);
		line-height: 1.12;
		letter-spacing: var(--tracking-tight);
	}

	@container agenda-event (max-width: 14rem) {
		.calendar-name {
			display: none;
		}
	}

	@container agenda-event (max-width: 9rem) {
		.event-meta {
			grid-template-columns: minmax(0, 1fr);
		}

		.event-meta time {
			overflow: hidden;
			text-overflow: ellipsis;
			white-space: nowrap;
		}
	}

	@container agenda-event (max-height: 5rem) {
		.event-content {
			gap: 0.12rem;
			padding: 0.42rem 0.58rem;
		}

		.event-accent,
		.calendar-name {
			display: none;
		}

		.event-meta {
			grid-template-columns: minmax(0, 1fr);
		}
	}

	@container agenda-event (max-height: 3.25rem) {
		.event-content {
			align-content: center;
			padding-block: 0.25rem;
		}

		.event-meta {
			display: none;
		}
	}

	@media (max-width: 520px) {
		.time-block {
			grid-row: auto;
			width: auto;
			margin-left: 0;
			overflow: visible;
			container-type: normal;
		}

		.event-content {
			height: auto;
		}

		.event-meta {
			display: flex;
			flex-wrap: wrap;
		}
	}
</style>
