<script lang="ts">
import Modal from './Modal.svelte';
import { ALLOWED_EXTS } from '../../lib/constants';
import { getUploadFolderOptions, upsertDiscoveredVideo, refreshDerivedState } from '../../lib/stores/library';
import { uploadModalOpen } from '../../lib/stores/ui';
import { uploadFile } from '../../lib/api';
import { toast } from '../../lib/toast';
import type { UploadFileItem, UploadResult, Video } from '../../lib/types';

let uploadFiles: UploadFileItem[] = [];
let selectedFolder = '/';
let newFolderInput = '';
let showNewFolder = false;
let uploading = false;

$: folders = getUploadFolderOptions();
$: validCount = uploadFiles.filter((f) => f.status !== 'error').length;
$: summary = uploadFiles.length
	? `${validCount} of ${uploadFiles.length} file${uploadFiles.length > 1 ? 's' : ''} ready`
	: '';

function close() {
	uploadModalOpen.set(false);
	uploadFiles = [];
}

function addFiles(files: FileList | File[]) {
	for (const f of files) {
		const ext = f.name.split('.').pop()?.toLowerCase() || '';
		const ok = ALLOWED_EXTS.has(ext);
		uploadFiles = [
			...uploadFiles,
			{
				file: f,
				id: Math.random().toString(36).slice(2),
				status: ok ? 'pending' : 'error',
				error: ok ? '' : 'Unsupported type',
			},
		];
	}
}

function removeFile(id: string) {
	uploadFiles = uploadFiles.filter((f) => f.id !== id);
}

function handleFolderChange(e: Event) {
	const value = (e.target as HTMLSelectElement).value;
	if (value === '__new__') {
		showNewFolder = true;
	} else {
		showNewFolder = false;
		newFolderInput = '';
		selectedFolder = value;
	}
}

async function startUpload() {
	let dest = selectedFolder;
	if (dest === '__new__') dest = newFolderInput.trim() || '/';
	const todo = uploadFiles.filter((f) => f.status === 'pending');
	uploading = true;
	for (const item of todo) {
		item.status = 'uploading';
		uploadFiles = [...uploadFiles];
		try {
			const results = (await uploadFile(item.file, dest)) as UploadResult[];
			const r = results[0];
			if (r?.error) {
				item.status = 'error';
				item.error = r.error;
			} else {
				item.status = 'done';
				if (r?.video) {
					upsertDiscoveredVideo(r.video as Video);
				}
				refreshDerivedState(false);
			}
		} catch {
			item.status = 'error';
			item.error = 'failed';
		}
		uploadFiles = [...uploadFiles];
	}
	uploading = false;
	toast('Upload complete', 'success');
}
</script>

<Modal id="upload-modal" open={$uploadModalOpen} on:close={close}>
	<div class="upload-inner">
		<div class="upload-header">
			<div class="upload-title">Upload videos</div>
			<button class="close-btn" id="upload-close" on:click={close}>✕</button>
		</div>
		<div class="upload-body">
			<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
			<div
				class="drop-zone"
				id="drop-zone"
				on:click={() => document.getElementById('upload-input')?.click()}
				on:dragover|preventDefault={() => {}}
				on:drop|preventDefault={(e) => addFiles([...e.dataTransfer!.files])}
			>
				<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
					<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" /><polyline points="17 8 12 3 7 8" /><line x1="12" y1="3" x2="12" y2="15" />
				</svg>
				<strong>Drop videos here</strong>
				<span>or click to browse — webm, mp4, mkv, mov, avi…</span>
				<input type="file" id="upload-input" multiple accept=".webm,.mp4,.mpv,.mkv,.mov,.avi,.m4v,.ogv" on:change={(e) => { addFiles((e.target as HTMLInputElement).files!); (e.target as HTMLInputElement).value = ''; }} />
			</div>
			<div class="upload-folder-row">
				<label for="upload-folder-select">Destination</label>
				<select id="upload-folder-select" value={selectedFolder} on:change={handleFolderChange}>
					{#each folders as f}
						<option value={f}>{f === '/' ? 'Root (/)' : f}</option>
					{/each}
					<option value="__new__">+ new folder…</option>
				</select>
				<input id="new-folder-input" class:visible={showNewFolder} placeholder="new/folder/name" bind:value={newFolderInput} />
			</div>
			<div class="file-queue" id="file-queue">
				{#each uploadFiles as item (item.id)}
					<div class="file-item" data-id={item.id}>
						<span class="file-item-name">{item.file.name}</span>
						<span class="file-item-size">{(item.file.size / 1048576).toFixed(1)} MB</span>
						<span class="file-item-status {item.status}">{item.status === 'error' ? item.error : item.status}</span>
						<button class="file-remove" on:click={() => removeFile(item.id)}>×</button>
					</div>
				{/each}
			</div>
		</div>
		<div class="upload-footer">
			<span class="upload-summary" id="upload-summary">{summary}</span>
			<div style="display:flex;gap:8px">
				<button class="btn" id="upload-cancel-btn" on:click={close}>Cancel</button>
				<button class="btn active" id="upload-start-btn" disabled={validCount === 0 || uploading} on:click={startUpload}>Upload</button>
			</div>
		</div>
	</div>
</Modal>
