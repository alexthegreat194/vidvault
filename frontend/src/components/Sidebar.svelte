<script lang="ts">
import SidebarBatches from './sidebar/SidebarBatches.svelte';
import SidebarTags from './sidebar/SidebarTags.svelte';
import SidebarCollections from './sidebar/SidebarCollections.svelte';
import SidebarFolders from './sidebar/SidebarFolders.svelte';
import { previewSettingsOpen, sidebarOpen, discoverMode } from '../lib/stores/ui';

function openDiscover() {
	discoverMode.set(true);
	sidebarOpen.set(false);
}
</script>

<aside class:sidebar-open={$sidebarOpen}>
	<div class="sidebar-section" style="padding:0 12px 8px 16px;">
		<button type="button" class="btn" style="width:100%; display:flex; align-items:center; gap:8px; justify-content:center; color:var(--accent); border-color:var(--accent);" on:click={openDiscover}>
			<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 3 21 3 21 8"/><line x1="4" y1="20" x2="21" y2="3"/><polyline points="21 16 21 21 16 21"/><line x1="15" y1="15" x2="21" y2="21"/><line x1="4" y1="4" x2="9" y2="9"/></svg>
			Discover
		</button>
	</div>
	<SidebarBatches />
	<SidebarTags />
	<SidebarCollections />
	<SidebarFolders />
	<button
		type="button"
		class="btn preview-settings-launcher"
		id="preview-settings-launcher"
		title="Preview settings"
		aria-label="Preview settings"
		on:click={() => previewSettingsOpen.set(true)}
	>
		<svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" aria-hidden="true">
			<circle cx="12" cy="12" r="3" />
			<path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
		</svg>
		<span class="preview-settings-launcher-label">settings</span>
	</button>
</aside>
