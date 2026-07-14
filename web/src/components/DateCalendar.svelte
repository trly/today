<script>
	import { CalendarDate } from '@internationalized/date';
	import { Calendar } from 'bits-ui';

	let {
		date,
		currentTime,
		panelId,
		onDateSelect = () => {},
		onClose = () => {}
	} = $props();

	const panelHeadingId = $props.id();
	const fullDateFormatter = new Intl.DateTimeFormat(undefined, {
		weekday: 'long',
		month: 'long',
		day: 'numeric',
		year: 'numeric'
	});

	const toCalendarDate = (value) =>
		new CalendarDate(value.getFullYear(), value.getMonth() + 1, value.getDate());
	const toDate = (value) => new Date(value.year, value.month - 1, value.day);
	const isSameDay = (left, right) =>
		left.getFullYear() === right.getFullYear() &&
		left.getMonth() === right.getMonth() &&
		left.getDate() === right.getDate();

	let calendarValue = $derived(toCalendarDate(date));
	let selectedLabel = $derived(fullDateFormatter.format(date));
	let isToday = $derived(isSameDay(date, currentTime));

	const handleValueChange = (value) => {
		if (!value) {
			return;
		}

		onDateSelect(toDate(value));
	};

	const handleToday = () => {
		onDateSelect(new Date(currentTime.getFullYear(), currentTime.getMonth(), currentTime.getDate()));
	};

	const handleKeydown = (event) => {
		if (event.key !== 'Escape') {
			return;
		}

		event.preventDefault();
		event.stopPropagation();
		onClose();
	};

</script>

<svelte:window onkeydown={handleKeydown} />

<section
	id={panelId}
	class="date-calendar-panel"
	aria-labelledby={panelHeadingId}
>
	<div class="calendar-intro">
		<p class="eyebrow" id={panelHeadingId}>Choose day</p>

		<div class="calendar-actions" role="group" aria-label="Calendar actions">
			<button type="button" class="calendar-action" disabled={isToday} onclick={handleToday}>
				Today
			</button>
			<button type="button" class="calendar-action" onclick={onClose}>Back</button>
		</div>
	</div>

	<Calendar.Root
		class="calendar-root"
		type="single"
		weekdayFormat="short"
		fixedWeeks={true}
		numberOfMonths={1}
		pagedNavigation={true}
		calendarLabel="Choose date"
		value={calendarValue}
		onValueChange={handleValueChange}
	>
		{#snippet children({ months, weekdays })}
			<Calendar.Header class="calendar-header">
				<Calendar.PrevButton class="calendar-nav" aria-label="Previous month">
					<span aria-hidden="true">&lsaquo;</span>
				</Calendar.PrevButton>
				<Calendar.Heading class="calendar-heading" />
				<Calendar.NextButton class="calendar-nav" aria-label="Next month">
					<span aria-hidden="true">&rsaquo;</span>
				</Calendar.NextButton>
			</Calendar.Header>

			<div class="calendar-months">
				{#each months as month (month.value.toString())}
					<div class="calendar-month">
						<Calendar.Grid class="calendar-grid">
							<Calendar.GridHead>
								<Calendar.GridRow class="calendar-week calendar-week-head">
									{#each weekdays as day (day)}
										<Calendar.HeadCell class="calendar-head-cell">
											<abbr title={day}>{day.slice(0, 2)}</abbr>
										</Calendar.HeadCell>
									{/each}
								</Calendar.GridRow>
							</Calendar.GridHead>
							<Calendar.GridBody>
								{#each month.weeks as weekDates, weekIndex (weekIndex)}
									<Calendar.GridRow class="calendar-week">
										{#each weekDates as dayDate (dayDate.toString())}
											<Calendar.Cell date={dayDate} month={month.value} class="calendar-cell">
												<Calendar.Day class="calendar-day">
													<span class="today-dot" aria-hidden="true"></span>
													<span class="selected-marker" aria-hidden="true"></span>
													<span class="day-number">{dayDate.day}</span>
												</Calendar.Day>
											</Calendar.Cell>
										{/each}
									</Calendar.GridRow>
								{/each}
							</Calendar.GridBody>
						</Calendar.Grid>
					</div>
				{/each}
			</div>
		{/snippet}
	</Calendar.Root>

	<p class="sr-only" aria-live="polite">Selected date: {selectedLabel}. Selecting a date returns to the day view.</p>
</section>

<style>
	.date-calendar-panel {
		display: grid;
		grid-template-rows: auto minmax(0, 1fr);
		gap: 0.75rem;
		height: 100%;
		min-width: 0;
		min-height: 0;
		padding: clamp(0.8rem, 3cqi, 1.15rem);
		background: linear-gradient(180deg, var(--color-header-surface-start), var(--color-header-surface-end));
		overflow: hidden;
	}

	.calendar-intro {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto;
		align-items: center;
		gap: 0.65rem;
	}

	.eyebrow,
	.sr-only {
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

	.calendar-actions {
		display: flex;
		flex-wrap: wrap;
		justify-content: end;
		gap: 0.45rem;
	}

	.calendar-action,
	:global(.calendar-nav) {
		appearance: none;
		border: 1px solid var(--color-border);
		background: var(--color-surface-raised);
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
			background-color var(--duration-state) var(--ease-out-quart),
			color var(--duration-state) var(--ease-out-quart),
			transform var(--duration-state) var(--ease-out-quart);
	}

	.calendar-action {
		min-width: 3.8rem;
		min-height: var(--size-touch);
		padding: 0.48rem 0.7rem;
		border-radius: var(--radius-pill);
	}

	.calendar-action:not(:disabled):hover,
	.calendar-action:focus-visible,
	:global(.calendar-nav:hover),
	:global(.calendar-nav:focus-visible) {
		border-color: var(--color-moss-border-hover);
		color: var(--color-moss-deep);
	}

	.calendar-action:focus-visible,
	:global(.calendar-nav:focus-visible),
	:global(.calendar-day:focus-visible) {
		outline: var(--focus-ring);
		outline-offset: var(--focus-offset);
	}

	.calendar-action:not(:disabled):active,
	:global(.calendar-nav:active),
	:global(.calendar-day:active) {
		transform: translateY(1px);
	}

	.calendar-action:disabled {
		cursor: default;
		opacity: 0.52;
	}

	:global(.calendar-root) {
		display: grid;
		grid-template-rows: auto minmax(0, 1fr);
		gap: 0.55rem;
		min-height: 0;
		padding: 0.65rem;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-panel);
		background: color-mix(in oklch, var(--color-panel) 56%, var(--color-surface-raised));
	}

	:global(.calendar-header) {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr) auto;
		align-items: center;
		gap: 0.55rem;
	}

	:global(.calendar-heading) {
		justify-self: center;
		font-size: 1rem;
		line-height: var(--leading-tight);
		font-weight: var(--weight-heading);
		letter-spacing: var(--tracking-tight);
	}

	:global(.calendar-nav) {
		display: inline-grid;
		place-items: center;
		width: var(--size-touch);
		min-height: var(--size-touch);
		border-radius: var(--radius-pill);
	}

	:global(.calendar-nav span) {
		font-size: 2rem;
		line-height: 0.7;
		font-weight: 500;
	}

	.calendar-months {
		display: grid;
		grid-template-columns: minmax(0, 1fr);
		min-height: 0;
	}

	.calendar-month {
		display: grid;
		min-width: 0;
		min-height: 0;
	}

	:global(.calendar-grid) {
		display: grid;
		gap: 0.2rem;
		width: 100%;
		min-height: 0;
		border-collapse: collapse;
		user-select: none;
	}

	:global(.calendar-week) {
		display: grid;
		grid-template-columns: repeat(7, minmax(0, 1fr));
		gap: 0.2rem;
	}

	:global(.calendar-head-cell) {
		display: grid;
		place-items: center;
		min-height: 1.35rem;
		font-size: var(--text-label);
		line-height: var(--leading-label);
		font-weight: var(--weight-heading);
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--color-ink-soft);
	}

	abbr {
		text-decoration: none;
	}

	:global(.calendar-cell) {
		min-width: 0;
	}

	:global(.calendar-day) {
		position: relative;
		display: grid;
		place-items: center;
		width: 100%;
		height: 100%;
		min-height: var(--size-touch);
		border: 1px solid var(--color-border-soft);
		border-radius: 0.65rem;
		background: color-mix(in oklch, var(--color-surface-raised) 82%, transparent);
		font-size: 0.95rem;
		line-height: 1;
		font-weight: var(--weight-heading);
		font-variant-numeric: tabular-nums;
		color: var(--color-ink);
		cursor: pointer;
		transition:
			border-color var(--duration-state) var(--ease-out-quart),
			background-color var(--duration-state) var(--ease-out-quart),
			color var(--duration-state) var(--ease-out-quart),
			transform var(--duration-state) var(--ease-out-quart);
	}

	:global(.calendar-day:hover) {
		border-color: var(--color-moss-focus);
		background: var(--color-moss-soft);
	}

	:global(.calendar-day[data-selected]) {
		border-color: var(--color-moss-deep);
		border-width: 2px;
		background: var(--color-moss-deep);
		color: var(--color-surface-raised);
	}

	.selected-marker {
		position: absolute;
		bottom: 0.28rem;
		width: 0.36rem;
		aspect-ratio: 1;
		border-radius: var(--radius-pill);
		background: var(--color-surface-raised);
		display: none;
	}

	:global(.calendar-day[data-selected] .selected-marker) {
		display: block;
	}

	:global(.calendar-day[data-today]:not([data-selected])) {
		border-color: var(--color-moss-border-today);
		color: var(--color-moss-deep);
	}

	:global(.calendar-day[data-outside-month]) {
		background: transparent;
		color: var(--color-ink-soft-disabled);
	}

	:global(.calendar-day[data-disabled]) {
		pointer-events: none;
		opacity: 0.42;
	}

	.today-dot {
		position: absolute;
		top: 0.42rem;
		width: 0.34rem;
		aspect-ratio: 1;
		border-radius: var(--radius-pill);
		background: currentColor;
		display: none;
	}

	:global(.calendar-day[data-today] .today-dot) {
		display: block;
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

	@media (max-width: 820px) {
		.date-calendar-panel {
			height: auto;
			min-height: 20rem;
		}
	}

	@media (max-width: 520px) {
		.date-calendar-panel {
			padding-inline: var(--space-mobile-inline);
		}

		:global(.calendar-root) {
			padding: 0.72rem;
		}

		:global(.calendar-week),
		:global(.calendar-grid) {
			gap: 0.24rem;
		}

		:global(.calendar-day) {
			min-height: var(--size-touch);
			border-radius: 0.68rem;
		}
	}
</style>
