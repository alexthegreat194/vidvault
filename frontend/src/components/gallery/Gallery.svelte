<script lang="ts">
import VideoCard from './VideoCard.svelte';
import {
	allVideos,
	filteredVideos,
	folderMeta,
	filters,
	isVideoDiscoveryLoading,
} from '../../lib/stores/library';
import { galleryListView } from '../../lib/stores/ui';
import { refresh } from '../../lib/stores/library';
import { clearNewStatus } from '../../lib/api';
import { toast } from '../../lib/toast';

$: list = $filteredVideos;
$: f = $filters;
</script>

<div id="gallery" class:list-view={$galleryListView}>
	{#if !list.length}
		{#if $isVideoDiscoveryLoading && $allVideos.length === 0}
			<div class="empty">
				<strong>Loading videos...</strong>Please wait while files are discovered.
			</div>
		{:else}
			<div class="empty">
				<strong>Nothing here</strong>
				{#if f.activeFolder !== '__all__' && $folderMeta[f.activeFolder]?.has_other_files}
					<span class="empty-warn">⚠ This folder contains non-video files that aren't shown here.</span>
				{:else}
					Try a different search or folder
				{/if}
			</div>
		{/if}
	{:else}
		{#each list as video, i (video.path)}
			<VideoCard {video} index={i} listView={$galleryListView} />
		{/each}
		{#if f.activeIncomingFilter}
			<div
				style="grid-column: 1 / -1; display: flex; justify-content: center; padding-top: 8px;"
			>
				<button
					class="btn"
					on:click={async () => {
						const paths = list.map((v) => v.path).filter(Boolean);
						if (!paths.length) return;
						const ok = await clearNewStatus(paths, false);
						if (!ok) {
							toast('Failed to mark filtered videos as reviewed', 'error');
							return;
						}
						await refresh();
						toast('Marked filtered videos as reviewed', 'success');
					}}
				>
					Mark as resolved
				</button>
			</div>
		{/if}
	{/if}
</div>
