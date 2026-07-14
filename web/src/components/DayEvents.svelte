<script>
	import { Dialog } from 'bits-ui';
	import { tick } from 'svelte';
	import Agenda from './Agenda.svelte';
	import AllDayEvents from './AllDayEvents.svelte';
	import Header from './Header.svelte';
	import { listDayEvents } from './todayClient';

	let { date, currentTime, pollRequest = 0, onDateChange = () => {} } = $props();

	let allDayEvents = $state([]);
	let timedEvents = $state([]);
	let loading = $state(true);
	let error = $state('');
	let refreshError = $state('');
	let lastUpdated = $state(null);
	let calendarOpen = $state(false);
	let retryRequest = $state(0);
	let loadedDateKey;
	const calendarId = $props.id();
	const calendarPanelId = `${calendarId}-panel`;
	const calendarTriggerId = `${calendarId}-trigger`;
	const lastUpdatedFormatter = new Intl.DateTimeFormat(undefined, {
		dateStyle: 'medium',
		timeStyle: 'short'
	});
	const dateKey = (value) =>
		`${value.getFullYear()}-${value.getMonth() + 1}-${value.getDate()}`;

	const restoreCalendarTriggerFocus = async () => {
		await tick();
		document.getElementById(calendarTriggerId)?.focus({ preventScroll: true });
	};
	const focusCalendarDate = async () => {
		await tick();
		const panel = document.getElementById(calendarPanelId);
		const focusTarget = panel?.querySelector(
			'[data-bits-day][data-selected]:not([data-disabled]):not([data-outside-month]), [data-bits-day][data-focused]:not([data-disabled]), [data-bits-day][data-today]:not([data-disabled]):not([data-outside-month]), [data-bits-day]:not([data-disabled]):not([data-outside-month])'
		);

		if (focusTarget instanceof HTMLElement) {
			focusTarget.focus({ preventScroll: true });
		}
	};

	const handleCalendarClose = () => {
		if (!calendarOpen) {
			return;
		}

		calendarOpen = false;
		void restoreCalendarTriggerFocus();
	};

	const handleCalendarToggle = () => {
		if (calendarOpen) {
			handleCalendarClose();
			return;
		}

		calendarOpen = true;
		void focusCalendarDate();
	};

	const handleDateSelect = (nextDate) => {
		onDateChange(nextDate);
		handleCalendarClose();
	};
	const handleReturnToToday = () => {
		onDateChange(new Date());
	};

	const handleRetry = () => {
		retryRequest += 1;
	};

	const loadEvents = async (requestDate, requestDateKey, signal, background) => {
		if (!background) {
			loading = true;
			error = '';
			refreshError = '';
		}

		try {
			const response = await listDayEvents(requestDate, { signal });
			allDayEvents = response.allDayEvents;
			timedEvents = response.events;
			error = '';
			refreshError = '';
			lastUpdated = new Date();
			loadedDateKey = requestDateKey;
			loading = false;
		} catch (caught) {
			if (signal.aborted) {
				return;
			}

			const message = caught instanceof Error ? caught.message : 'Failed to load events';
			if (background) {
				refreshError = message;
				return;
			}

			allDayEvents = [];
			timedEvents = [];
			error = message;
		} finally {
			if (!signal.aborted && !background) {
				loading = false;
			}
		}
	};

	$effect(() => {
		retryRequest;
		const requestDate = date;
		pollRequest;
		const requestDateKey = dateKey(requestDate);
		const background = loadedDateKey === requestDateKey;

		const controller = new AbortController();
		void loadEvents(requestDate, requestDateKey, controller.signal, background);

		return () => controller.abort();
	});
</script>

<div class="dashboard-grid">
	{#if error || refreshError}
		<Dialog.Root>
			<Dialog.Trigger
				class={refreshError ? 'refresh-status' : 'error-status'}
				aria-label={refreshError ? undefined : 'Open schedule error details'}
			>
				<svg viewBox="0 0 24 24" aria-hidden="true">
					<path d="M12 8v5m0 3.5v.01M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
				</svg>
				{#if refreshError}
					<span aria-live="polite">
						<strong>Unable to refresh</strong>
						<small>Last updated {lastUpdatedFormatter.format(lastUpdated)}</small>
					</span>
				{/if}
			</Dialog.Trigger>
			<Dialog.Portal>
				<Dialog.Overlay class="error-dialog-overlay" />
				<Dialog.Content class="error-dialog-content">
					<Dialog.Title class="error-dialog-title">
						{refreshError ? 'Unable to refresh schedule' : 'Schedule unavailable'}
					</Dialog.Title>
					<Dialog.Description class="error-dialog-description">
						{#if refreshError}
							Showing events from the last successful update at
							{lastUpdatedFormatter.format(lastUpdated)}. The service reported:
						{:else}
							This day’s events could not be loaded. The service reported:
						{/if}
					</Dialog.Description>
					<p class="error-message">{refreshError || error}</p>
					<div class="error-dialog-actions">
						<Dialog.Close class="dialog-button dialog-button-secondary">Close</Dialog.Close>
						<button type="button" class="dialog-button" onclick={handleRetry}>Try again</button>
					</div>
					<Dialog.Close class="error-dialog-close" aria-label="Close error details">
						<svg viewBox="0 0 24 24" aria-hidden="true">
							<path d="m6 6 12 12M18 6 6 18" />
						</svg>
					</Dialog.Close>
				</Dialog.Content>
			</Dialog.Portal>
		</Dialog.Root>
	{/if}

	<section class="orientation-rail" aria-label="Day orientation">
		<Header
			{date}
			{currentTime}
			{calendarOpen}
			{calendarPanelId}
			{calendarTriggerId}
			onCalendarToggle={handleCalendarToggle}
			onDateSelect={handleDateSelect}
			onCalendarClose={handleCalendarClose}
			onReturnToToday={handleReturnToToday}
		/>

		<AllDayEvents events={allDayEvents} {loading} {error} onRetry={handleRetry} />
	</section>

	<div class="agenda-main">
		<Agenda events={timedEvents} {date} {currentTime} {loading} {error} onRetry={handleRetry} />
	</div>
</div>

<style>
	.dashboard-grid {
		position: relative;
		display: grid;
		grid-template-columns: clamp(22rem, 34vw, 34rem) minmax(0, 1fr);
		height: 100%;
		min-height: 0;
		background: linear-gradient(180deg, var(--color-surface), var(--color-paper));
	}

	:global(.error-status) {
		position: absolute;
		top: 1rem;
		left: 1rem;
		z-index: 10;
		display: grid;
		place-items: center;
		width: var(--size-touch);
		height: var(--size-touch);
		padding: 0;
		border: 1px solid var(--color-error);
		border-radius: var(--radius-pill);
		background: var(--color-error-bg);
		color: var(--color-error-text);
		box-shadow: 0 0 0 0 oklch(from var(--color-error) l c h / 0.35);
		cursor: pointer;
		animation: error-pulse 2s var(--ease-out-quart) infinite;
	}

	:global(.refresh-status) {
		position: absolute;
		top: 1rem;
		left: 50%;
		z-index: 10;
		display: flex;
		align-items: center;
		gap: 0.65rem;
		max-width: calc(100% - 2rem);
		min-height: var(--size-touch);
		padding: 0.5rem 0.85rem;
		border: 1px solid var(--color-warning);
		border-radius: var(--radius-pill);
		background: var(--color-warning-bg);
		color: var(--color-warning-text);
		font: inherit;
		text-align: left;
		box-shadow: 0 0.4rem 1.2rem oklch(from var(--color-warning-text) l c h / 0.12);
		transform: translateX(-50%);
		cursor: pointer;
	}

	:global(.refresh-status span) {
		display: grid;
		min-width: 0;
	}

	:global(.refresh-status strong) {
		font-size: var(--text-label);
		letter-spacing: 0.04em;
		text-transform: uppercase;
	}

	:global(.refresh-status small) {
		overflow: hidden;
		font-size: var(--text-caption);
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	:global(.error-status svg),
	:global(.refresh-status svg),
	:global(.error-dialog-close svg) {
		flex: 0 0 auto;
		width: 1.35rem;
		height: 1.35rem;
		fill: none;
		stroke: currentColor;
		stroke-linecap: round;
		stroke-linejoin: round;
		stroke-width: 1.8;
	}

	:global(.error-status:hover) {
		background: var(--color-error-hover);
	}

	:global(.refresh-status:hover) {
		background: var(--color-warning-hover);
	}

	:global(.error-status:focus-visible),
	:global(.refresh-status:focus-visible),
	:global(.error-dialog-close:focus-visible),
	:global(.dialog-button:focus-visible) {
		outline: var(--focus-ring);
		outline-offset: 3px;
	}

	:global(.error-dialog-overlay) {
		position: fixed;
		inset: 0;
		z-index: 40;
		background: oklch(20% 0.01 145deg / 0.48);
	}

	:global(.error-dialog-content) {
		position: fixed;
		top: 50%;
		left: 50%;
		z-index: 41;
		width: min(30rem, calc(100vw - 2rem));
		max-height: calc(100vh - 2rem);
		padding: 1.5rem;
		overflow: auto;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-event);
		background: var(--color-surface-raised);
		color: var(--color-ink);
		box-shadow: 0 1.5rem 4rem oklch(20% 0.01 145deg / 0.22);
		transform: translate(-50%, -50%);
	}

	:global(.error-dialog-title) {
		padding-right: 2.5rem;
		font-size: var(--text-title);
		font-weight: 750;
		letter-spacing: var(--tracking-title);
	}

	:global(.error-dialog-description) {
		margin-top: 0.65rem;
		color: var(--color-ink-muted);
		line-height: var(--leading-body);
	}

	.error-message {
		margin: 1rem 0 0;
		padding: 0.85rem;
		overflow-wrap: anywhere;
		border: 1px solid var(--color-error-border);
		border-radius: var(--radius-pill);
		background: var(--color-error-surface);
		color: var(--color-error-surface-text);
		font-size: var(--text-caption);
		line-height: var(--leading-body);
	}

	.error-dialog-actions {
		display: flex;
		justify-content: flex-end;
		gap: 0.7rem;
		margin-top: 1.25rem;
	}

	:global(.dialog-button),
	:global(.error-dialog-close) {
		appearance: none;
		border: 1px solid var(--color-moss-deep);
		background: var(--color-moss-deep);
		color: var(--color-surface-raised);
		font: inherit;
		cursor: pointer;
	}

	:global(.dialog-button) {
		min-height: var(--size-touch);
		padding: 0.55rem 0.9rem;
		border-radius: var(--radius-pill);
		font-size: var(--text-label);
		font-weight: 800;
		letter-spacing: var(--tracking-wide);
		text-transform: uppercase;
	}

	:global(.dialog-button-secondary) {
		border-color: var(--color-border);
		background: transparent;
		color: var(--color-ink-soft);
	}

	:global(.error-dialog-close) {
		position: absolute;
		top: 0.8rem;
		right: 0.8rem;
		display: grid;
		place-items: center;
		width: var(--size-touch);
		height: var(--size-touch);
		padding: 0;
		border-color: transparent;
		border-radius: var(--radius-pill);
		background: transparent;
		color: var(--color-ink-soft);
	}

	@keyframes error-pulse {
		50% {
			box-shadow: 0 0 0 0.5rem oklch(from var(--color-error) l c h / 0);
		}
	}

	.orientation-rail {
		display: grid;
		grid-template-rows: auto minmax(0, 1fr);
		min-width: 0;
		min-height: 0;
		border-right: 1px solid var(--color-border);
		background: var(--color-panel);
		overflow: hidden;
	}

	.agenda-main {
		min-width: 0;
		min-height: 0;
	}

	@media (max-width: 820px) {
		.dashboard-grid {
			height: auto;
			grid-template-columns: 1fr;
		}

		.orientation-rail {
			border-right: 0;
			border-bottom: 1px solid var(--color-border);
			overflow: visible;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		:global(.error-status) {
			animation: none;
		}
	}

	@media (min-width: 821px) and (max-width: 1040px) {
		.dashboard-grid {
			grid-template-columns: minmax(21.5rem, 42%) minmax(0, 1fr);
		}
	}

	@media (max-height: 720px) and (min-width: 821px) {
		.dashboard-grid {
			height: auto;
			min-height: 42rem;
		}
	}
</style>
