<script>
	let {
		date,
		currentTime,
		calendarOpen = false,
		calendarPanelId,
		calendarTriggerId,
		onCalendarToggle = () => {},
		onReturnToToday = () => {}
	} = $props();

	const weekdayFormatter = new Intl.DateTimeFormat(undefined, { weekday: 'long' });
	const monthFormatter = new Intl.DateTimeFormat(undefined, { month: 'long' });
	const dayFormatter = new Intl.DateTimeFormat(undefined, { day: '2-digit' });
	const yearFormatter = new Intl.DateTimeFormat(undefined, { year: 'numeric' });

	const dateTime = $derived(
		[
			date.getFullYear(),
			String(date.getMonth() + 1).padStart(2, '0'),
			String(date.getDate()).padStart(2, '0')
		].join('-')
	);

	const isToday = $derived(
		date.getFullYear() === currentTime.getFullYear() &&
			date.getMonth() === currentTime.getMonth() &&
			date.getDate() === currentTime.getDate()
	);
	const fullDateLabel = $derived(
		date.toLocaleDateString(undefined, {
			weekday: 'long',
			month: 'long',
			day: 'numeric',
			year: 'numeric'
		})
	);
	const dayContext = $derived(isToday ? 'Today' : 'Selected day');
	const calendarActionLabel = $derived(calendarOpen ? 'Close calendar' : 'Choose date');
</script>

<div class="date-stack">
	<div class="date-row">
		<h1 class="date-heading" aria-label={fullDateLabel}>
			<button
				id={calendarTriggerId}
				type="button"
				class="date-mark"
				aria-label={`${calendarActionLabel}. ${dayContext}: ${fullDateLabel}`}
				aria-expanded={calendarOpen}
				aria-controls={calendarPanelId}
				onclick={onCalendarToggle}
			>
				<span class="date-mark-binding" aria-hidden="true">
					<span></span>
					<span></span>
				</span>
				{#key dateTime}
					<time
						class="date-mark-content"
						datetime={dateTime}
						aria-label={date.toLocaleDateString()}
						aria-current={isToday ? 'date' : undefined}
					>
						<span class="date-mark-eyebrow" aria-hidden={isToday}>
							{isToday ? '' : 'Selected day'}
						</span>
						<span class="date-mark-day">{dayFormatter.format(date)}</span>
						<span class="date-mark-weekday">{weekdayFormatter.format(date)}</span>
						<span class="date-mark-month">{monthFormatter.format(date)}</span>
						<span class="date-mark-year">{yearFormatter.format(date)}</span>
					</time>
				{/key}
				<span class="date-mark-action" aria-hidden="true">
					<span>{calendarActionLabel}</span>
					<svg class="date-mark-cue" viewBox="0 0 12 8">
						<path d="M1.5 1.75 6 6.25l4.5-4.5" />
					</svg>
				</span>
			</button>
		</h1>
	</div>
	<div class="return-today-slot">
		{#if !isToday}
			<button type="button" class="return-today" onclick={onReturnToToday}>Return to today</button>
		{/if}
	</div>
</div>

<style>
	.date-stack {
		display: grid;
		justify-items: center;
		width: 100%;
		min-width: 0;
	}

	.date-row {
		display: flex;
		align-items: start;
		justify-content: center;
		width: 100%;
	}

	.date-heading {
		display: grid;
		justify-items: center;
		width: 100%;
		margin: 0;
		font: inherit;
	}

	.date-mark {
		appearance: none;
		position: relative;
		display: grid;
		grid-template-rows: minmax(0, 1fr) auto;
		margin: 0;
		border: 1px solid var(--color-date-mark-border);
		border-radius: var(--radius-event);
		background: var(--color-date-mark-background);
		box-shadow: var(--shadow-control);
		font: inherit;
		justify-items: center;
		width: min(100%, 13.25rem);
		aspect-ratio: 1;
		padding: clamp(0.68rem, 2.2vw, 0.82rem);
		font-variant-numeric: tabular-nums;
		color: var(--color-header-ink-muted, var(--color-ink-muted));
		cursor: pointer;
		transition:
			border-color var(--duration-state) var(--ease-out-quart),
			background-color var(--duration-state) var(--ease-out-quart),
			box-shadow var(--duration-state) var(--ease-out-quart),
			color var(--duration-state) var(--ease-out-quart);
	}

	.date-mark time {
		display: grid;
		align-content: center;
		justify-items: center;
		height: 100%;
		width: 100%;
		padding: clamp(0.86rem, 2.2vw, 1rem) 0.62rem clamp(0.62rem, 1.8vw, 0.76rem);
	}

	.date-mark-content {
		animation: date-mark-enter 220ms var(--ease-out-quart) both;
	}

	.date-mark-binding {
		position: absolute;
		inset: 0.28rem 1.3rem auto;
		display: flex;
		z-index: 2;
		justify-content: space-between;
		pointer-events: none;
	}

	.date-mark-binding span {
		width: 0.38rem;
		height: 0.38rem;
		border: 1px solid var(--color-moss-border);
		border-radius: var(--radius-pill);
		background: var(--color-panel);
	}

	.date-mark-action {
		position: relative;
		z-index: 1;
		display: inline-flex;
		align-items: center;
		gap: 0.38rem;
		padding-top: 0.32rem;
		font-size: var(--text-label);
		font-weight: 700;
		line-height: var(--leading-label);
		letter-spacing: 0.055em;
		color: var(--color-moss-deep);
	}

	.date-mark-cue {
		width: 0.65rem;
		height: auto;
		flex: 0 0 auto;
		fill: none;
		stroke: currentColor;
		stroke-width: 1.6;
		stroke-linecap: round;
		stroke-linejoin: round;
		opacity: 0.72;
		transition:
			opacity var(--duration-state) var(--ease-out-quart),
			transform var(--duration-state) var(--ease-out-quart);
	}

	.date-mark[aria-expanded='true'] {
		border-color: var(--color-moss-border-hover);
		background-color: var(--color-moss-control-wash);
		box-shadow: var(--shadow-control-hover);
		color: var(--color-moss-deep);
	}

	.date-mark[aria-expanded='true'] .date-mark-cue {
		opacity: 0.82;
	}

	.date-mark[aria-expanded='true'] .date-mark-cue {
		transform: rotate(180deg);
	}

	.date-mark[aria-expanded='true'] {
		border-color: var(--color-moss-border-active);
	}

	.date-mark:active {
		box-shadow: var(--shadow-control-active);
	}

	.date-mark:focus-visible {
		outline: var(--focus-ring);
		outline-offset: var(--focus-offset);
	}

	.date-mark-month {
		margin-bottom: 0.12rem;
		font-size: 0.8125rem;
		line-height: var(--leading-label);
		font-weight: 750;
		letter-spacing: 0.085em;
		text-transform: uppercase;
		color: var(--color-moss-deep);
	}

	.date-mark-eyebrow {
		min-height: 0.8625rem;
		font-size: var(--text-label);
		font-weight: 750;
		line-height: var(--leading-label);
		letter-spacing: var(--tracking-label);
		text-transform: uppercase;
		color: var(--color-header-ink-muted, var(--color-ink-muted));
	}

	.return-today-slot {
		display: grid;
		place-items: start center;
		min-height: var(--size-touch);
		margin-top: 0.55rem;
	}

	.date-mark-day {
		margin-bottom: 0.2rem;
		font-size: clamp(var(--text-display), 10cqi, 3.55rem);
		line-height: var(--leading-tight);
		font-weight: 800;
		letter-spacing: var(--tracking-display);
		color: var(--color-ink-strong);
	}

	.date-mark-weekday {
		margin-bottom: 0.52rem;
		font-family: var(--font-display);
		font-size: clamp(var(--text-title), 4.5cqi, 1.45rem);
		font-weight: 700;
		line-height: var(--leading-tight);
		letter-spacing: var(--tracking-tight);
		color: var(--color-ink);
	}

	.date-mark-year {
		font-size: var(--text-label);
		line-height: var(--leading-label);
		font-weight: 650;
		letter-spacing: 0.12em;
		color: var(--color-header-ink-muted, var(--color-ink-muted));
	}

	.return-today {
		appearance: none;
		border: 0;
		border-radius: var(--radius-pill);
		background: transparent;
		min-height: var(--size-touch);
		padding: 0.38rem 0.62rem;
		font: inherit;
		font-size: var(--text-label);
		font-weight: 750;
		line-height: var(--leading-label);
		letter-spacing: 0.055em;
		color: var(--color-moss-deep);
		cursor: pointer;
		transition:
			background-color var(--duration-state) var(--ease-out-quart),
			color var(--duration-state) var(--ease-out-quart);
	}

	.return-today:active {
		background-color: var(--color-moss-control-wash);
	}

	.return-today:focus-visible {
		outline: var(--focus-ring);
		outline-offset: var(--focus-offset);
	}

	@keyframes date-mark-enter {
		from {
			opacity: 0;
			transform: translateY(0.35rem);
		}

		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	@media (hover: hover) {
		.date-mark:hover {
			border-color: var(--color-moss-border-hover);
			background-color: var(--color-moss-control-wash);
			box-shadow: var(--shadow-control-hover);
			color: var(--color-moss-deep);
		}

		.date-mark:hover .date-mark-cue {
			opacity: 0.82;
		}

		.return-today:hover {
			background-color: var(--color-moss-control-wash);
			color: var(--color-ink-strong);
		}
	}

	@media (max-width: 520px) {
		.date-row {
			align-items: center;
		}

		.date-mark {
			width: 9.5rem;
			padding: 0.58rem;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.date-mark,
		.date-mark-cue {
			transition-duration: 1ms;
		}

		.date-mark-content {
			animation-duration: 1ms;
		}
	}
</style>
