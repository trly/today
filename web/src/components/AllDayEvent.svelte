<script>
	let { event } = $props();

	let calendarColor = $derived(event.calendarColor || event.calendar_color || 'var(--color-moss)');
	let calendarName = $derived(event.calendar || event.calendarName || event.calendar_name || event.meta);
</script>

<li style:--calendar-color={calendarColor}>
	<span>{event.title}</span>
	{#if calendarName}
		<small>{calendarName}</small>
	{/if}
</li>

<style>
	li {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		column-gap: 0.55rem;
		row-gap: 0.18rem;
		padding: 0.72rem 0;
		border-top: 1px solid var(--color-border);
	}

	li::before {
		content: '';
		grid-row: 1 / span 2;
		width: 0.48rem;
		height: 0.48rem;
		margin-top: 0.24rem;
		border-radius: var(--radius-pill);
		background: var(--calendar-color, var(--color-moss));
	}

	li:last-child {
		border-bottom: 1px solid var(--color-border);
	}

	span {
		grid-column: 2;
		overflow-wrap: anywhere;
		font-size: var(--text-body);
		line-height: 1.28;
		font-weight: 650;
		letter-spacing: -0.01em;
	}

	small {
		grid-column: 2;
		overflow-wrap: anywhere;
		font-size: var(--text-caption);
		line-height: var(--leading-body);
		color: var(--color-ink-muted);
	}
</style>
