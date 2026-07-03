<script lang="ts">
import Modal from './Modal.svelte';
import {
	hoverPreviewAudioEnabled,
	hoverPreviewVolume,
	previewSettingsOpen,
	toggleHoverPreviewAudio,
} from '../../lib/stores/ui';

function close() {
	previewSettingsOpen.set(false);
}
</script>

<Modal
	id="preview-settings-modal"
	open={$previewSettingsOpen}
	on:close={close}
>
	<div class="preview-settings-inner">
		<div class="preview-settings-header">
			<div class="preview-settings-title">Preview settings</div>
			<button
				type="button"
				class="close-btn"
				id="preview-settings-close"
				aria-label="Close"
				on:click={close}
			>
				✕
			</button>
		</div>
		<div class="preview-settings-body">
			<div class="preview-settings-row">
				<span class="preview-settings-label">Hover preview audio</span>
				<button
					type="button"
					class="btn"
					class:active={$hoverPreviewAudioEnabled}
					id="preview-settings-audio-toggle"
					on:click={toggleHoverPreviewAudio}
				>
					{$hoverPreviewAudioEnabled ? 'On' : 'Off'}
				</button>
			</div>
			<label class="preview-settings-row preview-settings-volume-row" for="preview-settings-volume">
				<span class="preview-settings-label">Preview volume</span>
				<input
					id="preview-settings-volume"
					type="range"
					min="0"
					max="1"
					step="0.05"
					bind:value={$hoverPreviewVolume}
				/>
			</label>
		</div>
	</div>
</Modal>
