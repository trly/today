<script>
	import AllDayEvent from './AllDayEvent.svelte';

	let { events = [], loading = false, error = '', onRetry = undefined } = $props();

	const headingId = $props.id();
	const formatEventCount = (count) => {
		if (count === 0) {
			return 'No all-day commitments';
		}

		return `${count} all-day ${count === 1 ? 'commitment' : 'commitments'}`;
	};

	let eventCountLabel = $derived(formatEventCount(events.length));
</script>

<section class="all-day-panel" aria-labelledby={headingId}>
	<div class="section-heading">
		<h2 class="sr-only" id={headingId}>All-day commitments</h2>
		<p class="eyebrow" role={error ? 'alert' : 'status'}>
			{loading ? 'Loading all-day commitments' : error ? 'All-day commitments unavailable' : eventCountLabel}
		</p>
	</div>

	<ul class="all-day-list" aria-label="All-day commitments">
		{#if loading}
			<li class="loading-state" aria-hidden="true">Loading…</li>
		{:else if error}
			<li class="empty-state empty-state-action">
				<p>All-day commitments could not load.</p>
				{#if onRetry}
					<button type="button" onclick={onRetry}>Try again</button>
				{/if}
			</li>
		{:else if events.length === 0}
			<li class="empty-state">Nothing all-day today</li>
		{:else}
			{#each events as event (event.id)}
				<AllDayEvent {event} />
			{/each}
		{/if}
	</ul>
</section>

<style>
	.all-day-panel {
		display: grid;
		grid-template-rows: auto minmax(0, 1fr);
		min-width: 0;
		min-height: 0;
		padding: var(--space-panel);
		overflow: hidden;
	}

	.section-heading {
		display: grid;
		gap: 0.32rem;
		margin-bottom: 1rem;
	}

	.eyebrow {
		margin: 0;
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

	.all-day-list {
		display: grid;
		align-content: start;
		gap: 0;
		margin: 0;
		padding: 0;
		overflow: auto;
		overscroll-behavior: contain;
		list-style: none;
	}

	.empty-state,
	.loading-state {
		padding: 0.72rem 0;
		border-top: 1px solid var(--color-border);
	}

	.empty-state {
		font-size: var(--text-caption);
		line-height: var(--leading-body);
		color: var(--color-ink-muted);
	}

	.empty-state p {
		margin: 0;
	}

	.empty-state-action {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 0.8rem;
	}

	.empty-state button {
		appearance: none;
		min-width: var(--size-touch);
		min-height: var(--size-touch);
		padding: 0.58rem 0.78rem;
		border: 1px solid var(--color-border);
		border-radius: var(--radius-pill);
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
			color var(--duration-state) var(--ease-out-quart),
			transform var(--duration-state) var(--ease-out-quart);
	}

	.empty-state button:hover,
	.empty-state button:focus-visible {
		border-color: var(--color-moss-border-hover);
		color: var(--color-moss-deep);
	}

	.empty-state button:focus-visible {
		outline: var(--focus-ring);
		outline-offset: 2px;
	}

	.empty-state button:active {
		transform: translateY(1px);
	}

	.loading-state {
		font-size: var(--text-caption);
		line-height: var(--leading-body);
		color: var(--color-ink-muted);
	}

	@media (max-width: 820px) {
		.all-day-panel {
			overflow: visible;
		}

		.all-day-list {
			overflow: visible;
		}
	}
</style>
