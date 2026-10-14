<script lang="ts">
import { onMount } from 'svelte';
import './app.css';
import { setOnPinLock, setPinGateUnlocked } from './lib/auth';
import {
	bootstrap,
	pinGateMode,
	pinGateUnlocked,
	setPinOverlayVisible,
	unloadLibraryAfterPinLock,
} from './lib/stores/library';
import { setPinGateMode } from './lib/auth';
import PinGate from './components/PinGate.svelte';
import AppHeader from './components/AppHeader.svelte';
import DupBanner from './components/DupBanner.svelte';
import IncomingBanner from './components/IncomingBanner.svelte';
import Sidebar from './components/Sidebar.svelte';
import Gallery from './components/gallery/Gallery.svelte';
import ContextMenu from './components/ContextMenu.svelte';
import SelectBar from './components/SelectBar.svelte';
import VideoPlayerModal from './components/modals/VideoPlayerModal.svelte';
import MoveModal from './components/modals/MoveModal.svelte';
import TagModal from './components/modals/TagModal.svelte';
import UploadModal from './components/modals/UploadModal.svelte';
import DuplicateModal from './components/modals/DuplicateModal.svelte';
import IncomingModal from './components/modals/IncomingModal.svelte';
import CollectionPickerModal from './components/modals/CollectionPickerModal.svelte';
import WatchCollectionModal from './components/modals/WatchCollectionModal.svelte';
import PreviewSettingsModal from './components/modals/PreviewSettingsModal.svelte';
import DiscoverView from './components/gallery/DiscoverView.svelte';
import { discoverMode, sidebarOpen } from './lib/stores/ui';

onMount(() => {
	setOnPinLock(() => {
		setPinGateMode(true);
		pinGateMode.set(true);
		setPinGateUnlocked(false);
		pinGateUnlocked.set(false);
		unloadLibraryAfterPinLock();
		setPinOverlayVisible(true, 'Session locked or PIN required.');
	});
	bootstrap();
});
</script>

<PinGate />
<AppHeader />
<DupBanner />
<IncomingBanner />

<!-- svelte-ignore a11y_no_static_element_interactions a11y_click_events_have_key_events -->
<div class="sidebar-backdrop" class:visible={$sidebarOpen} on:click={() => sidebarOpen.set(false)}></div>
<div class="shell">
	<Sidebar />
	<main>
		<Gallery />
	</main>
</div>

{#if $discoverMode}
	<DiscoverView />
{/if}

<ContextMenu />
<SelectBar />
<VideoPlayerModal />
<MoveModal />
<TagModal />
<UploadModal />
<DuplicateModal />
<IncomingModal />
<CollectionPickerModal />
<WatchCollectionModal />
<PreviewSettingsModal />
