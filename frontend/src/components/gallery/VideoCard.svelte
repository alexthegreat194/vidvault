<script lang="ts">
import { createEventDispatcher, onDestroy } from 'svelte';
import { escHtml, formatBytes, formatModified, videoUrl } from '../../lib/format';
import { DEFAULT_LOOP_ALL_VIDEOS, HOVER_PREVIEW_SEEK_SEC, CARD_DRAG_GHOST_OFFSET } from '../../lib/constants';
import {
	hoverPreviewAudioEnabled,
	hoverPreviewVolume,
	showCtxMenu,
	toggleHoverPreviewAudio,
} from '../../lib/stores/ui';
import { tagMap } from '../../lib/stores/library';
import { setFavorite } from '../../lib/api';
import { toast } from '../../lib/toast';
import { openPlayer } from '../../lib/stores/player';
import {
	enterSelectMode,
	selectMode,
	selectedPaths,
	toggleSelectPath,
} from '../../lib/stores/selection';
import type { Video } from '../../lib/types';

export let video: Video;
export let index = 0;

const dispatch = createEventDispatcher<{
	dragstart: { event: DragEvent; path: string };
}>();

let cardEl: HTMLDivElement;
let vidEl: HTMLVideoElement;
let hoverPreview = false;
let observer: IntersectionObserver | undefined;
let srcSet = false;

const PREVIEW_AUDIO_SVG_OFF =
	'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"/><line x1="23" y1="9" x2="17" y2="15"/><line x1="17" y1="9" x2="23" y2="15"/></svg>';
const PREVIEW_AUDIO_SVG_ON =
	'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true"><polygon points="11 5 6 9 2 9 2 15 6 15 11 19 11 5"/><path d="M15.54 8.46a5 5 0 0 1 0 7.07"/><path d="M19.07 4.93a10 10 0 0 1 0 14.14"/></svg>';

$: selected = $selectedPaths.has(video.path);
$: tagChips = (video.tags || [])
	.map((id) => $tagMap[id])
	.filter(Boolean);

function applyAudio() {
	if (!vidEl) return;
	vidEl.volume = $hoverPreviewVolume;
	vidEl.muted = !$hoverPreviewAudioEnabled;
}

function seekAndPlay() {
	if (!vidEl?.src || !hoverPreview) return;
	const doPlay = () => {
		if (!hoverPreview) return;
		const dur = vidEl.duration;
		if (dur && !Number.isNaN(dur) && Number.isFinite(dur)) {
			const t = Math.min(HOVER_PREVIEW_SEEK_SEC, Math.max(0, dur - 0.05));
			try {
				vidEl.currentTime = t;
			} catch {
				/* ignore */
			}
		}
		applyAudio();
		vidEl.play().catch(() => {});
	};
	if (vidEl.readyState >= HTMLMediaElement.HAVE_CURRENT_DATA) {
		doPlay();
	} else {
		vidEl.addEventListener('loadeddata', doPlay, { once: true });
	}
}

function stopPreview() {
	hoverPreview = false;
	if (!vidEl?.src) return;
	vidEl.pause();
	const dur = vidEl.duration;
	if (dur && !Number.isNaN(dur) && Number.isFinite(dur)) {
		try {
			vidEl.currentTime = Math.min(HOVER_PREVIEW_SEEK_SEC, Math.max(0, dur - 0.05));
		} catch {
			/* ignore */
		}
	}
}

function setupObserver() {
	observer?.disconnect();
	observer = new IntersectionObserver(
		(entries) => {
			if (entries[0].isIntersecting && vidEl && !srcSet) {
				srcSet = true;
				vidEl.src = videoUrl(video.path);
				vidEl.addEventListener(
					'loadeddata',
					() => {
						vidEl.classList.add('loaded');
						applyAudio();
						if (hoverPreview) seekAndPlay();
					},
					{ once: true },
				);
				observer?.disconnect();
			}
		},
		{ threshold: 0.1 },
	);
	if (cardEl) observer.observe(cardEl);
}

$: if (cardEl) setupObserver();

$: if (vidEl) {
	vidEl.loop = DEFAULT_LOOP_ALL_VIDEOS;
	applyAudio();
}

$: if ($hoverPreviewAudioEnabled !== undefined || $hoverPreviewVolume !== undefined) {
	applyAudio();
}

onDestroy(() => observer?.disconnect());

async function handleFav(e: MouseEvent) {
	e.preventDefault();
	e.stopPropagation();
	const nextValue = !video.is_favorite;
	const ok = await setFavorite(video.hash, nextValue);
	if (!ok) {
		toast('Favorite update failed', 'error');
		return;
	}
	video.is_favorite = nextValue;
}

function handleCardClick(e: MouseEvent) {
	const target = e.target as HTMLElement;
	if (target.closest('.drag-handle')) return;
	if (target.closest('.fav-toggle')) return;
	if (target.closest('.preview-audio-btn')) return;
	if (target.closest('.check-overlay')) {
		if (!$selectMode) enterSelectMode();
		toggleSelectPath(video.path);
		return;
	}
	if ($selectMode) {
		toggleSelectPath(video.path);
	} else {
		openPlayer(video.path);
	}
}

function handleContextMenu(e: MouseEvent) {
	e.preventDefault();
	showCtxMenu(e, video);
}

function handleDragStart(e: DragEvent) {
	cardEl?.classList.add('dragging');
	e.dataTransfer?.setData('text/plain', video.path);
	if (e.dataTransfer) e.dataTransfer.effectAllowed = 'move';
	dispatch('dragstart', { event: e, path: video.path });
}

function handleDragEnd() {
	cardEl?.classList.remove('dragging');
}

let dragGhost: HTMLDivElement | undefined;

function ensureDragGhost(): HTMLDivElement {
	if (dragGhost) return dragGhost;
	dragGhost = document.createElement('div');
	dragGhost.className = 'card-drag-ghost';
	dragGhost.setAttribute('aria-hidden', 'true');
	dragGhost.innerHTML =
		'<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" aria-hidden="true">' +
		'<rect x="2.5" y="5" width="19" height="14" rx="2" stroke="rgba(255,255,255,.65)" stroke-width="1.5" fill="rgba(255,255,255,.08)"/>' +
		'<polygon points="10,9.5 10,14.5 15,12" fill="rgba(255,255,255,.58)"/>' +
		'</svg>';
	document.body.appendChild(dragGhost);
	return dragGhost;
}

function handleDragStartWithGhost(e: DragEvent) {
	handleDragStart(e);
	try {
		e.dataTransfer?.setDragImage(
			ensureDragGhost(),
			CARD_DRAG_GHOST_OFFSET,
			CARD_DRAG_GHOST_OFFSET,
		);
	} catch {
		/* unsupported */
	}
}
</script>

<!-- svelte-ignore a11y_no_static_element_interactions a11y_click_events_have_key_events -->
<div
	bind:this={cardEl}
	class="card"
	class:selected
	class:dragging={false}
	style="animation-delay: {Math.min(index, 30) * 15}ms"
	draggable="true"
	data-path={video.path}
	on:click={handleCardClick}
	on:contextmenu={handleContextMenu}
	on:mouseenter={() => {
		hoverPreview = true;
		seekAndPlay();
	}}
	on:mouseleave={stopPreview}
	on:dragstart={handleDragStartWithGhost}
	on:dragend={handleDragEnd}
>
	<div class="thumb">
		<video bind:this={vidEl} preload="metadata" muted></video>
		<button
			class="fav-toggle"
			class:active={video.is_favorite}
			title={video.is_favorite ? 'Remove favorite' : 'Mark favorite'}
			aria-label={video.is_favorite ? 'Remove favorite' : 'Mark favorite'}
			on:click={handleFav}
		>
			★
		</button>
		<div class="play-icon">
			<svg viewBox="0 0 36 36" fill="none">
				<circle cx="18" cy="18" r="17" fill="rgba(0,0,0,.55)" stroke="rgba(255,255,255,.25)" stroke-width="1" />
				<polygon points="14,11 27,18 14,25" fill="white" />
			</svg>
		</div>
		<div class="drag-handle" title="Drag to move">
			<svg xmlns="http://www.w3.org/2000/svg" width="10" height="10" viewBox="0 0 10 18">
				<circle cx="3" cy="2" r="1.5" fill="white" />
				<circle cx="7" cy="2" r="1.5" fill="white" />
				<circle cx="3" cy="9" r="1.5" fill="white" />
				<circle cx="7" cy="9" r="1.5" fill="white" />
				<circle cx="3" cy="16" r="1.5" fill="white" />
				<circle cx="7" cy="16" r="1.5" fill="white" />
			</svg>
		</div>
		<div class="check-overlay">
			<svg viewBox="0 0 10 8" fill="none" width="10" height="8">
				<polyline points="1,4 3.5,7 9,1" stroke="#000" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" />
			</svg>
		</div>
		<button
			type="button"
			class="preview-audio-btn"
			class:active={$hoverPreviewAudioEnabled}
			title={$hoverPreviewAudioEnabled
				? 'Preview sound on for all cards (click to mute)'
				: 'Preview sound off (click to enable for all cards)'}
			aria-label="Toggle hover preview sound for all cards"
			aria-pressed={$hoverPreviewAudioEnabled ? 'true' : 'false'}
			on:click|stopPropagation|preventDefault={toggleHoverPreviewAudio}
		>
			{@html $hoverPreviewAudioEnabled ? PREVIEW_AUDIO_SVG_ON : PREVIEW_AUDIO_SVG_OFF}
		</button>
	</div>
	<div class="card-meta">
		<span class="card-name" title={video.name}>{video.name}</span>
		<span class="card-path">{video.folder || '/'}</span>
		<div class="card-badges">
			<span class="card-ext">{video.ext.slice(1)}</span>
			{#if video.is_new}
				<span class="card-ext">new</span>
			{/if}
			{#if tagChips.length}
				<span class="card-tags card-tags--badges">
					{#each tagChips as tag (tag.id)}
						<span class="tag-chip" style="--tag-color: {tag.color || '#888888'}">{tag.name}</span>
					{/each}
				</span>
			{/if}
		</div>
		<span class="card-path">{formatBytes(video.size || 0)} • {formatModified(video.modified)}</span>
	</div>
</div>
