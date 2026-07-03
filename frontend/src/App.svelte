<script lang="ts">
import { onMount } from 'svelte';
import './app.css';
import { setOnPinLock, setPinGateUnlocked } from './lib/auth';
import {
	bootstrap,
	pinGateUnlocked,
	setPinOverlayVisible,
	unloadLibraryAfterPinLock,
} from './lib/stores/library';
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

onMount(() => {
	setOnPinLock(() => {
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

<div class="shell">
	<Sidebar />
	<main>
		<Gallery />
	</main>
</div>

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
