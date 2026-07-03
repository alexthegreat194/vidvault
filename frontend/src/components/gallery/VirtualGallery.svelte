<script lang="ts">
import { onMount } from 'svelte';
import { createVirtualizer } from '@tanstack/svelte-virtual';
import VideoCard from './VideoCard.svelte';
import type { Video } from '../../lib/types';

export let videos: Video[] = [];
export let listView = false;

const GRID_MIN = 260;
const GRID_GAP = 12;
const LIST_ROW_HEIGHT = 68;
/** card-meta: padding, 2-line name, path, badges, size line, gaps */
const CARD_META_HEIGHT = 128;

let scrollEl: HTMLElement | null = null;
let containerEl: HTMLDivElement;
let containerWidth = 0;

function chunk<T>(items: T[], size: number): T[][] {
	if (size <= 0) return [];
	const rows: T[][] = [];
	for (let i = 0; i < items.length; i += size) {
		rows.push(items.slice(i, i + size));
	}
	return rows;
}

$: cols = listView
	? 1
	: Math.max(1, Math.floor((containerWidth + GRID_GAP) / (GRID_MIN + GRID_GAP)));
$: colWidth =
	cols > 0 ? (containerWidth - (cols - 1) * GRID_GAP) / cols : GRID_MIN;
$: gridRowHeight = listView
	? LIST_ROW_HEIGHT
	: Math.ceil(colWidth * (9 / 16)) + CARD_META_HEIGHT + GRID_GAP;
$: rows = listView ? videos.map((v) => [v]) : chunk(videos, cols);

$: virtualizer = createVirtualizer({
	count: rows.length,
	getScrollElement: () => scrollEl,
	estimateSize: () => gridRowHeight,
	overscan: 4,
});

onMount(() => {
	scrollEl = containerEl?.closest('main') ?? document.querySelector('main');
	const measure = () => {
		if (scrollEl) containerWidth = scrollEl.clientWidth;
	};
	measure();
	const ro = new ResizeObserver(() => measure());
	if (scrollEl) ro.observe(scrollEl);
	return () => ro.disconnect();
});
</script>

<div
	id="gallery"
	class:list-view={listView}
	class:virtual-gallery={true}
	bind:this={containerEl}
	style="--grid-cols: {cols}"
>
	<div class="virtual-spacer" style="height: {$virtualizer.getTotalSize()}px;">
		{#each $virtualizer.getVirtualItems() as row (row.key)}
			{@const rowVideos = rows[row.index]}
			{@const startIndex = row.index * cols}
			<div
				class="virtual-row"
				class:list-row={listView}
				style="transform: translateY({row.start}px); min-height: {row.size}px;"
			>
				{#each rowVideos as video, j (video.path)}
					<VideoCard {video} index={startIndex + j} />
				{/each}
			</div>
		{/each}
	</div>
</div>

<style>
	.virtual-spacer {
		position: relative;
		width: 100%;
	}
	.virtual-row {
		position: absolute;
		top: 0;
		left: 0;
		right: 0;
		width: 100%;
		box-sizing: border-box;
		display: grid;
		grid-template-columns: repeat(var(--grid-cols, 1), minmax(0, 1fr));
		gap: 12px;
		align-items: start;
	}
	.virtual-row.list-row {
		grid-template-columns: 1fr;
		gap: 4px;
	}
</style>
