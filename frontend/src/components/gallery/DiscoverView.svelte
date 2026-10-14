<script lang="ts">
import { onMount, onDestroy } from 'svelte';
import { discoverMode } from '../../lib/stores/ui';
import { fetchRandomVideos, setFavorite } from '../../lib/api';
import { videoUrl } from '../../lib/format';
import { tagMap, allVideos } from '../../lib/stores/library';
import { toast } from '../../lib/toast';
import type { Video } from '../../lib/types';
import { get } from 'svelte/store';

const BATCH_SIZE = 8;
const PREFETCH_THRESHOLD = 3;

let videos: Video[] = [];
let seenHashes: string[] = [];
let loading = false;
let scrollEl: HTMLDivElement;
let videoEls: HTMLVideoElement[] = [];
let currentIndex = 0;
let discoverMuted = true;
let observer: IntersectionObserver;

const MUTE_SVG =
	'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"/><line x1="23" y1="9" x2="17" y2="15"/><line x1="17" y1="9" x2="23" y2="15"/></svg>';
const VOLUME_SVG =
	'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"/><path d="M15.54 8.46a5 5 0 0 1 0 7.07"/><path d="M19.07 4.93a10 10 0 0 1 0 14.14"/></svg>';

function applyMuteToAll() {
	for (const vid of videoEls) {
		if (!vid) continue;
		vid.muted = discoverMuted;
		if (!discoverMuted) vid.volume = 1;
	}
}

function toggleMute() {
	discoverMuted = !discoverMuted;
	applyMuteToAll();
}

function shuffleLocal(count: number): Video[] {
	const all = get(allVideos);
	const excludeSet = new Set(seenHashes);
	const candidates = all.filter((v) => !excludeSet.has(v.hash));
	for (let i = candidates.length - 1; i > 0; i--) {
		const j = Math.floor(Math.random() * (i + 1));
		[candidates[i], candidates[j]] = [candidates[j], candidates[i]];
	}
	return candidates.slice(0, count);
}

async function loadMore() {
	if (loading) return;
	loading = true;
	try {
		let batch: Video[];
		try {
			batch = await fetchRandomVideos(BATCH_SIZE, seenHashes);
		} catch {
			batch = shuffleLocal(BATCH_SIZE);
		}
		if (batch.length === 0) return;
		videos = [...videos, ...batch];
		seenHashes = [...seenHashes, ...batch.map((v) => v.hash)];
	} finally {
		loading = false;
	}
}

function handleClose() {
	discoverMode.set(false);
}

function setupObserver() {
	observer = new IntersectionObserver(
		(entries) => {
			for (const entry of entries) {
				const idx = Number(entry.target.getAttribute('data-idx'));
				if (isNaN(idx)) continue;
				const vid = videoEls[idx];
				if (!vid) continue;
				if (entry.isIntersecting) {
					currentIndex = idx;
					vid.muted = discoverMuted;
					if (!discoverMuted) vid.volume = 1;
					vid.play().catch(() => {});
					if (idx >= videos.length - PREFETCH_THRESHOLD) {
						loadMore();
					}
				} else {
					vid.pause();
				}
			}
		},
		{ root: scrollEl, threshold: 0.6 },
	);
}

function observeItems() {
	if (!scrollEl || !observer) return;
	scrollEl.querySelectorAll('.discover-item').forEach((el) => {
		observer.observe(el);
	});
}

$: if (videos.length && scrollEl && observer) {
	requestAnimationFrame(observeItems);
}

$: if (videoEls.length) applyMuteToAll();

async function toggleFav(video: Video) {
	const next = !video.is_favorite;
	const ok = await setFavorite(video.hash, next);
	if (!ok) {
		toast('Favorite update failed', 'error');
		return;
	}
	const idx = videos.indexOf(video);
	if (idx >= 0) {
		videos[idx] = { ...videos[idx], is_favorite: next };
		videos = videos;
	}
}

function scrollToIndex(idx: number) {
	const items = scrollEl?.querySelectorAll('.discover-item');
	if (items && items[idx]) {
		items[idx].scrollIntoView({ behavior: 'smooth' });
	}
}

function handleKeydown(e: KeyboardEvent) {
	if (e.key === 'Escape') handleClose();
	else if (e.key === 'ArrowDown' || e.key === 'ArrowRight') {
		e.preventDefault();
		const next = Math.min(currentIndex + 1, videos.length - 1);
		scrollToIndex(next);
	} else if (e.key === 'ArrowUp' || e.key === 'ArrowLeft') {
		e.preventDefault();
		const prev = Math.max(currentIndex - 1, 0);
		scrollToIndex(prev);
	}
}

onMount(() => {
	loadMore();
	setupObserver();
	document.addEventListener('keydown', handleKeydown);
});

onDestroy(() => {
	observer?.disconnect();
	document.removeEventListener('keydown', handleKeydown);
});
</script>

<div class="discover-view">
	<div class="discover-header">
		<span class="discover-title">DISCOVER</span>
		<div class="discover-header-actions">
			<button
				type="button"
				class="discover-mute"
				class:active={!discoverMuted}
				on:click={toggleMute}
				title={discoverMuted ? 'Unmute' : 'Mute'}
				aria-label={discoverMuted ? 'Unmute videos' : 'Mute videos'}
				aria-pressed={discoverMuted ? 'false' : 'true'}
			>
				{@html discoverMuted ? MUTE_SVG : VOLUME_SVG}
			</button>
			<button class="discover-close" on:click={handleClose} title="Close">✕</button>
		</div>
	</div>
	<div class="discover-scroll" bind:this={scrollEl}>
		{#each videos as video, i (video.path + i)}
			<div class="discover-item" data-idx={i}>
				<!-- svelte-ignore a11y_media_has_caption -->
				<video
					bind:this={videoEls[i]}
					src={videoUrl(video.path)}
					loop
					muted={discoverMuted}
					playsinline
					preload="auto"
				></video>
				<div class="discover-overlay">
					<span class="discover-name">{video.name}</span>
					<span class="discover-folder">{video.folder || '/'}</span>
					{#if video.tags?.length}
						<div class="discover-tags">
							{#each video.tags as tid}
								{#if $tagMap[tid]}
									<span class="tag-chip" style="--tag-color: {$tagMap[tid].color || '#888'}">{$tagMap[tid].name}</span>
								{/if}
							{/each}
						</div>
					{/if}
					<div class="discover-actions">
						<button
							class="discover-fav"
							class:active={video.is_favorite}
							on:click={() => toggleFav(video)}
							title={video.is_favorite ? 'Remove favorite' : 'Mark favorite'}
						>★</button>
					</div>
				</div>
			</div>
		{/each}
		{#if loading}
			<div class="discover-spinner">Loading more...</div>
		{/if}
	</div>
</div>
