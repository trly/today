<script>
	import { onMount } from 'svelte';
	import DayEvents from './components/DayEvents.svelte';

	let currentTime = $state(new Date());
	let selectedDate = $state(new Date());
	let pollRequest = $state(0);

	const isSameLocalDate = (left, right) =>
		left.getFullYear() === right.getFullYear() &&
		left.getMonth() === right.getMonth() &&
		left.getDate() === right.getDate();

	const REFRESH_INTERVAL_MS = 5 * 60 * 1000;

	const handleDateChange = (date) => {
		selectedDate = date;
	};

	onMount(() => {
		let clockTimeout;
		let refreshTimeout;
		let lastRefreshEpoch = Date.now();

		const scheduleNextMinute = () => {
			window.clearTimeout(clockTimeout);

			if (document.hidden) {
				return;
			}

			const now = new Date();
			const millisecondsUntilNextMinute =
				60_000 - (now.getSeconds() * 1000 + now.getMilliseconds());
			clockTimeout = window.setTimeout(tickClock, millisecondsUntilNextMinute);
		};
		const scheduleNextRefresh = () => {
			window.clearTimeout(refreshTimeout);

			if (document.hidden) {
				return;
			}

			const elapsed = Date.now() - lastRefreshEpoch;
			const delay = Math.max(0, REFRESH_INTERVAL_MS - elapsed);
			refreshTimeout = window.setTimeout(triggerRefresh, delay);
		};
		const updateDateOnDayCrossing = (now) => {
			if (isSameLocalDate(selectedDate, currentTime) && !isSameLocalDate(currentTime, now)) {
				selectedDate = new Date(now.getFullYear(), now.getMonth(), now.getDate());
			}
		};
		const tickClock = () => {
			const now = new Date();
			updateDateOnDayCrossing(now);
			currentTime = now;
			scheduleNextMinute();
		};
		const triggerRefresh = () => {
			lastRefreshEpoch = Date.now();

			if (isSameLocalDate(selectedDate, new Date())) {
				pollRequest += 1;
			}

			scheduleNextRefresh();
		};
		const handleVisibilityChange = () => {
			if (document.hidden) {
				window.clearTimeout(clockTimeout);
				window.clearTimeout(refreshTimeout);
				return;
			}

			const now = new Date();
			updateDateOnDayCrossing(now);
			currentTime = now;
			lastRefreshEpoch = Date.now();
			scheduleNextMinute();
			scheduleNextRefresh();
		};

		document.addEventListener('visibilitychange', handleVisibilityChange);
		scheduleNextMinute();
		scheduleNextRefresh();

		return () => {
			document.removeEventListener('visibilitychange', handleVisibilityChange);
			window.clearTimeout(clockTimeout);
			window.clearTimeout(refreshTimeout);
		};
	});
</script>

<main class="dashboard-shell" aria-label="Today dashboard">
	<section class="dashboard-board" aria-label="Day overview">
		<DayEvents date={selectedDate} {currentTime} {pollRequest} onDateChange={handleDateChange} />
	</section>
</main>

<style>
	.dashboard-shell {
		position: relative;
		min-height: 100svh;
		padding: var(--space-shell);
		overflow: auto;
		isolation: isolate;
	}

	.dashboard-board {
		display: grid;
		height: calc(100svh - var(--space-shell) - var(--space-shell));
		min-height: calc(100svh - var(--space-shell) - var(--space-shell));
		border: 1px solid var(--color-border);
		border-radius: var(--radius-board);
		background:
			linear-gradient(90deg, var(--color-surface-raised), var(--color-surface)),
			repeating-linear-gradient(
				90deg,
				var(--color-grid) 0 1px,
				transparent 1px 6.5rem
			);
		box-shadow: var(--shadow-board);
		overflow: hidden;
	}

	@media (max-width: 820px) {
		.dashboard-board {
			height: auto;
			min-height: auto;
		}
	}

	@media (max-height: 720px) and (min-width: 821px) {
		.dashboard-board {
			height: auto;
			min-height: calc(100svh - var(--space-shell) - var(--space-shell));
		}
	}

	@media (max-width: 520px) {
		.dashboard-shell {
			padding: 0;
		}

		.dashboard-board {
			border-right: 0;
			border-left: 0;
			border-radius: 0;
		}
	}
</style>
