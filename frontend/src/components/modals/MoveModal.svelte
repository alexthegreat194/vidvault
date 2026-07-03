<script lang="ts">
import { get } from 'svelte/store';
import Modal from './Modal.svelte';
import {
	allFolders,
	getMoveFolderOptions,
	patchVideoPath,
	refreshFoldersOnly,
	computeDuplicateGroups,
	updateDuplicateBannerState,
	updateIncomingBannerState,
} from '../../lib/stores/library';
import { moveModalOpen, moveSelectedFolder, moveTarget } from '../../lib/stores/ui';
import { moveVideoApi } from '../../lib/api';
import { selectedPaths, exitSelectMode } from '../../lib/stores/selection';
import { toast } from '../../lib/toast';

let newFolderName = '';

$: target = $moveTarget;
$: if (target) moveSelectedFolder.set(target.folder || '/');
$: folders = getMoveFolderOptions();

function close() {
	moveModalOpen.set(false);
	moveTarget.set(null);
	newFolderName = '';
}

function selectFolder(f: string) {
	moveSelectedFolder.set(f);
}

function addFolder() {
	const n = newFolderName.trim();
	if (!n) return;
	allFolders.update((list) => (list.includes(n) ? list : [...list, n]));
	newFolderName = '';
	moveSelectedFolder.set(n);
}

async function confirmMove() {
	const dest = $moveSelectedFolder;
	if (!dest) return;
	const paths =
		get(selectedPaths).size > 0
			? [...get(selectedPaths)]
			: target
				? [target.path]
				: [];
	if (!paths.length) return;
	let errs = 0;
	for (const p of paths) {
		const res = await moveVideoApi(p, dest);
		if (!res.ok) {
			errs++;
		} else if (res.video) {
			patchVideoPath(res.old_path || p, res.video);
		}
	}
	if (!errs) {
		try {
			await refreshFoldersOnly();
		} catch {
			/* best-effort */
		}
		computeDuplicateGroups();
		updateDuplicateBannerState(false);
		updateIncomingBannerState();
	}
	if (errs) toast(errs + ' move(s) failed', 'error');
	else {
		toast(
			'Moved ' +
				(paths.length > 1 ? paths.length + ' items' : '') +
				(paths.length > 1 ? ' to ' : '') +
				dest,
			'success',
		);
	}
	if (get(selectedPaths).size > 0) exitSelectMode();
	close();
}
</script>

<Modal id="move-modal" open={$moveModalOpen} on:close={close}>
	<div class="move-inner">
		<div class="move-header">
			<div>
				<div class="move-title">Move to folder</div>
				<div class="move-subtitle" id="move-subtitle">
					{get(selectedPaths).size > 1
						? get(selectedPaths).size + ' items selected'
						: target?.name || ''}
				</div>
			</div>
			<button class="close-btn" id="move-close" on:click={close}>✕</button>
		</div>
		<div class="move-body">
			<div class="move-new-folder">
				<input
					id="new-folder-name"
					placeholder="+ new folder name…"
					autocomplete="off"
					bind:value={newFolderName}
					on:keydown={(e) => e.key === 'Enter' && addFolder()}
				/>
				<button class="btn" id="add-folder-btn" on:click={addFolder}>Create</button>
			</div>
			<div id="folder-options">
				{#each folders as f (f)}
					<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
					<div
						class="folder-option"
						class:selected={$moveSelectedFolder === f}
						on:click={() => selectFolder(f)}
					>
						<svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
							<path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" />
						</svg>
						{f === '/' ? 'Root (/)' : f}
					</div>
				{/each}
			</div>
		</div>
		<div class="move-footer">
			<button class="btn" id="move-cancel-btn" on:click={close}>Cancel</button>
			<button class="btn active" id="move-confirm-btn" on:click={confirmMove}>Move</button>
		</div>
	</div>
</Modal>
