<script lang="ts">
import { onMount } from 'svelte';
import {
	adjacentInFiltered,
	filteredVideos,
	videoByPath,
} from '../../lib/stores/library';
import {
	closePlayer,
	openPlayer,
	playingPath,
} from '../../lib/stores/player';
import { DEFAULT_LOOP_ALL_VIDEOS } from '../../lib/constants';

let modalVid: HTMLVideoElement;
let open = false;

$: path = $playingPath;
$: video = path ? videoByPath(path) : undefined;
$: inFiltered = path ? $filteredVideos.some((v) => v.path === path) : false;

$: if (path && !inFiltered) {
	closePlayer();
}

$: open = Boolean(path && video && inFiltered);

$: if (open && video && modalVid) {
	modalVid.src = '/video?path=' + encodeURIComponent(video.path);
	modalVid.loop = DEFAULT_LOOP_ALL_VIDEOS;
}

function handleClose() {
	if (modalVid) {
		modalVid.pause();
		modalVid.src = '';
	}
	closePlayer();
}

function handlePrev() {
	if (!path) return;
	const prev = adjacentInFiltered(path, -1);
	if (prev) openPlayer(prev.path);
}

function handleNext() {
	if (!path) return;
	const next = adjacentInFiltered(path, 1);
	if (next) openPlayer(next.path);
}

function handleEnded() {
	handleNext();
}

function handleOverlayClick(e: MouseEvent) {
	if (e.target === e.currentTarget) handleClose();
}

function handleKeydown(e: KeyboardEvent) {
	if (!open) return;
	if (e.key === 'Escape') handleClose();
	if (e.key === 'ArrowLeft') handlePrev();
	if (e.key === 'ArrowRight') handleNext();
}

onMount(() => {
	document.addEventListener('keydown', handleKeydown);
	return () => document.removeEventListener('keydown', handleKeydown);
});
</script>

<div id="modal" class:open on:click={handleOverlayClick} role="presentation">
	{#if video}
		<div class="modal-inner">
			<div class="modal-header">
				<span class="modal-title" id="modal-title">{video.name}</span>
				<span class="modal-path" id="modal-path">{video.folder || '/'}</span>
				<div class="modal-nav">
					<button class="nav-btn" id="prev-btn" on:click|stopPropagation={handlePrev}>‹</button>
					<button class="nav-btn" id="next-btn" on:click|stopPropagation={handleNext}>›</button>
				</div>
				<button class="close-btn" id="close-btn" on:click|stopPropagation={handleClose}>✕</button>
			</div>
			<!-- svelte-ignore a11y_media_has_caption -->
			<video
				bind:this={modalVid}
				id="modal-video"
				controls
				autoplay
				loop
				on:ended={handleEnded}
			></video>
		</div>
	{/if}
</div>
