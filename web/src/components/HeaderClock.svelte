<script>
	import { timeFormatter } from './timeFormat.js';

	let { currentTime } = $props();

	const clockTime = $derived(currentTime.toTimeString().slice(0, 5));
	const formattedTime = $derived(timeFormatter.format(currentTime));
	const separatorIndex = $derived(formattedTime.indexOf(':'));
</script>

<div class="clock-card">
	<time datetime={clockTime} aria-label={formattedTime}>
		{#if separatorIndex === -1}
			{formattedTime}
		{:else}
			<span aria-hidden="true">{formattedTime.slice(0, separatorIndex)}</span><span
				class="clock-separator"
				aria-hidden="true">:</span
			><span aria-hidden="true">{formattedTime.slice(separatorIndex + 1)}</span>
		{/if}
	</time>
</div>

<style>
	.clock-card {
		display: grid;
		align-content: center;
		justify-items: start;
		gap: clamp(0.55rem, 2cqi, 0.8rem);
		min-width: 0;
		padding: 0.1rem 0;
	}

	time {
		font-family: var(--font-ui);
		font-size: clamp(var(--text-clock), 10.5cqi, 4.1rem);
		font-weight: var(--weight-heading);
		line-height: var(--leading-tight);
		letter-spacing: var(--tracking-display);
		font-variant-numeric: tabular-nums;
		font-feature-settings: "tnum" 1;
		color: var(--color-ink-strong);
		white-space: nowrap;
	}

	.clock-separator {
		color: var(--color-moss-deep);
	}

	@media (max-width: 520px) {
		.clock-card {
			justify-items: center;
		}
	}
</style>
