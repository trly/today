<script>
	import HeaderClock from './HeaderClock.svelte';
	import HeaderDate from './HeaderDate.svelte';
	import DateCalendar from './DateCalendar.svelte';

	let {
		date,
		currentTime,
		calendarOpen = false,
		calendarPanelId,
		calendarTriggerId,
		onCalendarToggle = () => {},
		onDateSelect = () => {},
		onCalendarClose = () => {},
		onReturnToToday = () => {}
	} = $props();
</script>

<header class="date-header" aria-label="Selected day and current local time">
	<div class="header-turntable" class:is-flipped={calendarOpen}>
		<div class="header-face header-front" aria-hidden={calendarOpen} inert={calendarOpen}>
			<div class="header-composition">
				<div class="date-pane">
					<HeaderDate
						{date}
						{currentTime}
						{calendarOpen}
						{calendarPanelId}
						{calendarTriggerId}
						{onCalendarToggle}
						{onReturnToToday}
					/>
				</div>
				<div class="clock-pane">
					<HeaderClock {currentTime} />
				</div>
			</div>
		</div>

		<div class="header-face header-back" aria-hidden={!calendarOpen} inert={!calendarOpen}>
			<DateCalendar
				{date}
				{currentTime}
				panelId={calendarPanelId}
				onDateSelect={onDateSelect}
				onClose={onCalendarClose}
			/>
		</div>
	</div>
</header>

<style>
	.date-header {
		--color-muted: var(--color-header-ink-muted);

		container-type: inline-size;
		border-bottom: 1px solid var(--color-header-rule);
		background:
			linear-gradient(180deg, var(--color-header-surface-start), var(--color-header-surface-end)),
			radial-gradient(circle at 0% 0%, var(--color-header-wash), transparent 42%);
	}

	.header-turntable {
		position: relative;
		display: grid;
		min-height: clamp(20rem, 72cqi, 23rem);
	}

	.header-face {
		grid-area: 1 / 1;
		min-width: 0;
		opacity: 1;
		transition: opacity 200ms ease;
	}

	.header-face[aria-hidden='true'] {
		opacity: 0;
		pointer-events: none;
	}

	.header-front {
		display: grid;
		align-items: center;
		padding: clamp(1.25rem, 2.5vw, 1.75rem) clamp(1rem, 2vw, 1.3rem)
			clamp(1.3rem, 2.2vw, 1.65rem);
	}

	.header-composition {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		align-items: center;
		width: 100%;
		max-width: 32.5rem;
		margin-inline: auto;
	}

	.date-pane,
	.clock-pane {
		min-width: 0;
	}

	.date-pane {
		display: grid;
		justify-items: center;
		padding-right: clamp(1rem, 4cqi, 1.4rem);
	}

	.clock-pane {
		display: grid;
		align-self: stretch;
		align-items: center;
		padding-left: clamp(1rem, 4cqi, 1.4rem);
		border-left: 1px solid var(--color-header-rule);
	}

	@container (max-width: 21.5rem) {
		.header-composition {
			grid-template-columns: 1fr;
			gap: 1.2rem;
		}

		.date-pane {
			padding-right: 0;
		}

		.clock-pane {
			justify-items: center;
			align-self: auto;
			padding-top: 1.2rem;
			padding-left: 0;
			border-left: 0;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.header-face {
			transition-duration: 1ms;
		}
	}
</style>
