<script>
	import { onMount, tick } from 'svelte';
	import AgendaEvent from './AgendaEvent.svelte';
	import { hourFormatter, timeFormatter } from './timeFormat.js';

	let {
		events = [],
		date,
		currentTime,
		loading = false,
		error = '',
		onRetry = undefined
	} = $props();

	const headingId = $props.id();
	const dayStartMinutes = 0;
	const dayEndMinutes = 24 * 60;
	const agendaHoursToShow = 24;
	const inactivityDelayMs = 5 * 60 * 1000;
	const activityEvents = ['scroll', 'wheel', 'pointerdown', 'touchstart', 'keydown'];
	// Keep timed-event loading text-only. Event-shaped skeletons read as fake calendar data, then vanish when real events arrive.
	const loadingLabel = 'Loading timed events';

	const formatHour = (hour) => hourFormatter.format(new Date(2000, 0, 1, hour % 24));

	const getEventEndMinutes = (event) => event.startMinutes + Math.max(1, event.durationMinutes);
	const getCurrentHourStartMinutes = () => currentTime.getHours() * 60;
	const formatCurrentTime = () => timeFormatter.format(currentTime);
	const clampEventToDay = (event) => ({
		...event,
		startMinutes: Math.max(dayStartMinutes, event.startMinutes),
		durationMinutes: Math.min(getEventEndMinutes(event), dayEndMinutes) - Math.max(dayStartMinutes, event.startMinutes)
	});

	const positionEvents = (events) => {
		const sortedEvents = events
			.map((event, index) => ({ event, index, endMinutes: getEventEndMinutes(event) }))
			.sort((a, b) => {
				const startDifference = a.event.startMinutes - b.event.startMinutes;

				if (startDifference !== 0) {
					return startDifference;
				}

				return a.endMinutes - b.endMinutes;
			});

		const positionedEvents = [];
		let group = [];
		let groupEndMinutes = -Infinity;

		const flushGroup = () => {
			if (group.length === 0) {
				return;
			}

			const laneEndMinutes = [];
			const groupEvents = group.map((item) => {
				const lane = laneEndMinutes.findIndex((endMinutes) => endMinutes <= item.event.startMinutes);
				const agendaLane = lane === -1 ? laneEndMinutes.length : lane;

				laneEndMinutes[agendaLane] = item.endMinutes;

				return { ...item, agendaLane };
			});

			const agendaLaneCount = laneEndMinutes.length;

			for (const item of groupEvents) {
				const overlappingEvents = groupEvents.filter(
					(other) =>
						other !== item &&
						item.event.startMinutes < other.endMinutes &&
						other.event.startMinutes < item.endMinutes
				);

				positionedEvents.push({
					...item.event,
					agendaLane: item.agendaLane,
					agendaLaneCount,
					conflictCount: overlappingEvents.length,
					conflictLabel: overlappingEvents.length
						? `Overlaps with ${overlappingEvents.map((other) => other.event.title).join(', ')}`
						: '',
					originalIndex: item.index
				});
			}

			group = [];
		};

		for (const item of sortedEvents) {
			if (group.length > 0 && item.event.startMinutes >= groupEndMinutes) {
				flushGroup();
			}

			group.push(item);
			groupEndMinutes = Math.max(groupEndMinutes, item.endMinutes);
		}

		flushGroup();

		return positionedEvents
			.sort((a, b) => a.originalIndex - b.originalIndex)
			.map(({ originalIndex, ...event }) => event);
	};

	let timelineViewport = $state(null);
	let halfHourRowHeight = $state(64);

	let halfHourRows = $derived(agendaHoursToShow * 2);
	let agendaHours = $derived(Array.from({ length: agendaHoursToShow }, (_, index) => formatHour(index)));
	let visibleEvents = $derived(
		events
			.filter((event) => event.startMinutes < dayEndMinutes && getEventEndMinutes(event) > dayStartMinutes)
			.map(clampEventToDay)
	);
	let positionedEvents = $derived(positionEvents(visibleEvents));
	let isToday = $derived(
		date.getFullYear() === currentTime.getFullYear() &&
			date.getMonth() === currentTime.getMonth() &&
			date.getDate() === currentTime.getDate()
	);
	let currentMinutes = $derived(currentTime.getHours() * 60 + currentTime.getMinutes());
	let currentTimeLabel = $derived(formatCurrentTime());
	let currentTimeOffset = $derived((currentMinutes / 30) * halfHourRowHeight);
	const formatEventCount = (count) => {
		if (count === 0) {
			return 'No timed events scheduled';
		}

		return `${count} timed ${count === 1 ? 'event' : 'events'} scheduled`;
	};

	let eventCountLabel = $derived(formatEventCount(positionedEvents.length));
	let currentHour = $derived(currentTime.getHours());

	$effect(() => {
		if (isToday) {
			void currentHour;
			tick().then(scrollToCurrentHourTop);
		}
	});

	const scrollToCurrentHourTop = () => {
		if (!isToday || !timelineViewport) {
			return;
		}

		timelineViewport.scrollTop = (getCurrentHourStartMinutes() / 30) * halfHourRowHeight;
	};

	const handleJumpToNow = () => {
		scrollToCurrentHourTop();
	};

	onMount(() => {
		let cancelled = false;
		let resizeObserver;
		let inactivityTimeout;

		const clearInactivityTimeout = () => {
			if (inactivityTimeout) {
				clearTimeout(inactivityTimeout);
			}
		};

		const resetInactivityTimer = () => {
			clearInactivityTimeout();
			inactivityTimeout = setTimeout(scrollToCurrentHourTop, inactivityDelayMs);
		};

		requestAnimationFrame(async () => {
			await tick();

			if (cancelled || !timelineViewport) {
				return;
			}

			const updateRowHeight = () => {
				if (timelineViewport.clientHeight > 0) {
					halfHourRowHeight = timelineViewport.clientHeight / 8;
				}
			};

			resizeObserver = new ResizeObserver(updateRowHeight);
			resizeObserver.observe(timelineViewport);
			updateRowHeight();
			await tick();
			scrollToCurrentHourTop();

			for (const event of activityEvents) {
				timelineViewport.addEventListener(event, resetInactivityTimer, { passive: true });
			}

			resetInactivityTimer();
		});

		return () => {
			cancelled = true;
			clearInactivityTimeout();

			if (timelineViewport) {
				for (const event of activityEvents) {
					timelineViewport.removeEventListener(event, resetInactivityTimer);
				}
			}

			resizeObserver?.disconnect();
		};
	});
</script>

<section
	class="agenda-panel"
	aria-labelledby={headingId}
	style={`--half-hour-rows: ${halfHourRows}; --half-hour-row-height: ${halfHourRowHeight}px;`}
>
	<div class="section-heading agenda-heading-row">
		<div>
			<h2 class="sr-only" id={headingId}>Timed events</h2>
			<p class="eyebrow">{loading ? loadingLabel : error ? 'Timed events unavailable' : eventCountLabel}</p>
		</div>
		{#if isToday}
			<button type="button" class="jump-to-now" onclick={handleJumpToNow}>
				Now
			</button>
		{/if}
	</div>

	<!-- svelte-ignore a11y_no_noninteractive_tabindex (the scroll region must be keyboard-focusable across browsers) -->
	<div
		bind:this={timelineViewport}
		class="agenda-scroll-viewport"
		role="region"
		aria-label="Timed events timeline"
		tabindex="0"
	>
		<div class="timeline">
			<div class="time-axis" aria-hidden="true">
				{#each agendaHours as hour, index (index)}
					<span>{hour}</span>
				{/each}
			</div>

			<div class="event-grid">
				{#if isToday}
					<div
						class="now-marker"
						aria-hidden="true"
						style={`--now-offset: ${currentTimeOffset}px;`}
					>
						<span>Now {currentTimeLabel}</span>
					</div>
					<p class="sr-only">Current time is {currentTimeLabel}</p>
				{/if}

				{#if loading}
					<p class="empty-state" aria-live="polite">{loadingLabel}</p>
				{:else if error}
					<div class="empty-state empty-state-action">
						<p>Timed events could not load.</p>
						{#if onRetry}
							<button type="button" onclick={onRetry}>Try again</button>
						{/if}
					</div>
				{:else if positionedEvents.length === 0}
					<p class="empty-state">No timed events scheduled</p>
				{:else}
					{#each positionedEvents as event (event.id)}
						<AgendaEvent {event} {date} />
					{/each}
				{/if}
			</div>
		</div>
	</div>
</section>

<style>
	.agenda-panel {
		display: grid;
		grid-template-rows: auto minmax(0, 1fr);
		gap: 0.2rem;
		height: 100%;
		min-width: 0;
		min-height: 0;
		padding: var(--space-panel-block) var(--space-panel-inline);
		overflow: hidden;
		--half-hour-row-height: 4rem;
	}

	.section-heading {
		display: grid;
		gap: 0.32rem;
		margin-bottom: 1rem;
	}

	.agenda-heading-row {
		grid-template-columns: minmax(0, 1fr) auto;
		align-items: start;
	}

	.jump-to-now {
		appearance: none;
		align-self: center;
		min-width: var(--size-touch);
		min-height: var(--size-touch);
		padding: 0.62rem 0.82rem;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-pill);
		background: transparent;
		font: inherit;
		font-size: var(--text-label);
		line-height: var(--leading-label);
		font-weight: 800;
		letter-spacing: var(--tracking-wide);
		text-transform: uppercase;
		color: var(--color-ink-soft);
		cursor: pointer;
		transition:
			border-color var(--duration-state) var(--ease-out-quart),
			color var(--duration-state) var(--ease-out-quart),
			transform var(--duration-state) var(--ease-out-quart);
	}

	.jump-to-now:hover,
	.jump-to-now:focus-visible {
		border-color: var(--color-moss-border-hover);
		color: var(--color-moss-deep);
	}

	.jump-to-now:focus-visible {
		outline: var(--focus-ring);
		outline-offset: 2px;
	}

	.jump-to-now:active {
		transform: translateY(1px);
	}

	.eyebrow,
	p {
		margin: 0;
	}

	.eyebrow {
		font-size: var(--text-label);
		line-height: var(--leading-label);
		font-weight: 750;
		letter-spacing: var(--tracking-label);
		text-transform: uppercase;
		color: var(--color-moss-deep);
	}

	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip: rect(0, 0, 0, 0);
		white-space: nowrap;
		border: 0;
	}

	.agenda-scroll-viewport {
		min-height: 0;
		height: 100%;
		overflow-y: auto;
		overscroll-behavior: contain;
		scrollbar-gutter: stable;
	}

	.agenda-scroll-viewport:focus-visible {
		outline: var(--focus-ring);
		outline-offset: -2px;
	}

	.timeline {
		display: grid;
		grid-template-columns: 4.5rem minmax(0, 1fr);
		height: calc(var(--half-hour-row-height) * var(--half-hour-rows));
		padding-top: 0.35rem;
	}

	.time-axis,
	.event-grid {
		display: grid;
		grid-template-rows: repeat(var(--half-hour-rows), var(--half-hour-row-height));
	}

	.time-axis span {
		grid-row: span 2;
		padding-top: 0.15rem;
		font-size: var(--text-label);
		line-height: var(--leading-label);
		font-weight: 700;
		font-variant-numeric: tabular-nums;
		color: var(--color-ink-soft);
	}

	.event-grid {
		position: relative;
		border-left: 1px solid var(--color-border);
		background: repeating-linear-gradient(
			to bottom,
			var(--color-grid) 0 1px,
			transparent 1px calc(var(--half-hour-row-height) * 2)
		);
	}

	.now-marker {
		position: absolute;
		z-index: 3;
		top: clamp(0rem, var(--now-offset), calc(100% - 1.4rem));
		left: -0.42rem;
		right: 0;
		display: flex;
		align-items: center;
		gap: 0.55rem;
		pointer-events: none;
		transform: translateY(-50%);
	}

	.now-marker::before {
		content: '';
		width: 0.82rem;
		aspect-ratio: 1;
		border: 2px solid var(--color-surface-raised);
		border-radius: var(--radius-pill);
		background: var(--color-moss-deep);
		box-shadow: var(--shadow-control-hover);
	}

	.now-marker::after {
		content: '';
		flex: 1;
		height: 1px;
		background: linear-gradient(90deg, var(--color-moss-deep), var(--color-moss-deep-transparent));
	}

	.now-marker span {
		order: 3;
		padding: 0.2rem 0.42rem;
		border: 1px solid var(--color-moss-border);
		border-radius: var(--radius-pill);
		background: var(--color-surface-raised);
		font-size: var(--text-label);
		line-height: var(--leading-label);
		font-weight: 800;
		letter-spacing: var(--tracking-wide);
		text-transform: uppercase;
		font-variant-numeric: tabular-nums;
		color: var(--color-moss-deep);
	}

	.empty-state {
		display: grid;
		gap: 0.25rem;
		padding: 0.85rem 0.9rem;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-item);
		background: var(--color-surface-raised);
	}

	.empty-state-action {
		grid-template-columns: minmax(0, 1fr) auto;
		align-items: center;
		gap: 0.9rem;
	}

	.empty-state-action button {
		appearance: none;
		min-width: var(--size-touch);
		min-height: var(--size-touch);
		padding: 0.58rem 0.78rem;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-pill);
		background: transparent;
		font: inherit;
		font-size: var(--text-label);
		line-height: var(--leading-label);
		font-weight: 800;
		letter-spacing: var(--tracking-wide);
		text-transform: uppercase;
		color: var(--color-ink-soft);
		cursor: pointer;
		transition:
			border-color var(--duration-state) var(--ease-out-quart),
			color var(--duration-state) var(--ease-out-quart),
			transform var(--duration-state) var(--ease-out-quart);
	}

	.empty-state-action button:hover,
	.empty-state-action button:focus-visible {
		border-color: var(--color-moss-border-hover);
		color: var(--color-moss-deep);
	}

	.empty-state-action button:focus-visible {
		outline: var(--focus-ring);
		outline-offset: 2px;
	}

	.empty-state-action button:active {
		transform: translateY(1px);
	}

	@media (max-width: 820px) {
		.timeline {
			grid-template-columns: 3.8rem minmax(0, 1fr);
			min-height: 38rem;
		}
	}

	@media (max-width: 520px) {
		.agenda-panel {
			padding-inline: var(--space-mobile-inline);
		}

		.timeline {
			grid-template-columns: 1fr;
			min-height: 0;
		}

		.time-axis {
			display: none;
		}

		.agenda-scroll-viewport {
			overflow: visible;
			scrollbar-gutter: auto;
		}

		.timeline {
			height: auto;
		}

		.event-grid {
			display: grid;
			grid-template-rows: none;
			gap: 0.75rem;
			border-left: 0;
			background: transparent;
		}

		.empty-state-action {
			grid-template-columns: 1fr;
			justify-items: start;
		}

		.now-marker {
			display: none;
		}
	}
</style>
