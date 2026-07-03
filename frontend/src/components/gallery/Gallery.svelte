<script lang="ts">
import VirtualGallery from './VirtualGallery.svelte';
import {
	allVideos,
	discoveredVideos,
	filteredVideos,
	folderMeta,
	filters,
	isVideoDiscoveryLoading,
	refreshDerivedState,
} from '../../lib/stores/library';
import { galleryListView } from '../../lib/stores/ui';
import { clearNewStatus } from '../../lib/api';
import { toast } from '../../lib/toast';

$: list = $filteredVideos;
$: f = $filters;

function markFilteredResolved(paths: string[]) {
	discoveredVideos.update((videos) =>
		videos.map((v) => (paths.includes(v.path) ? { ...v, is_new: false } : v)),
	);
	refreshDerivedState(false);
}
</script>

{#if !list.length}
	<div id="gallery" class:list-view={$galleryListView}>
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
	</div>
{:else}
	<VirtualGallery videos={list} listView={$galleryListView} />
	{#if f.activeIncomingFilter}
		<div class="incoming-resolve-row">
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
					markFilteredResolved(paths);
					toast('Marked filtered videos as reviewed', 'success');
				}}
			>
				Mark as resolved
			</button>
		</div>
	{/if}
{/if}

<style>
	.incoming-resolve-row {
		display: flex;
		justify-content: center;
		padding-top: 8px;
	}
</style>
