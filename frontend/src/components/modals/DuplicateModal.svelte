<script lang="ts">
import { get } from 'svelte/store';
import Modal from './Modal.svelte';
import { duplicateGroups, refresh } from '../../lib/stores/library';
import { dupModalOpen } from '../../lib/stores/ui';
import { deleteVideoPath } from '../../lib/api';
import { formatBytes, formatModified } from '../../lib/format';
import { toast } from '../../lib/toast';
import type { DuplicateGroup } from '../../lib/types';

let localGroups: DuplicateGroup[] = [];

$: if ($dupModalOpen) {
	localGroups = get(duplicateGroups).map((g) => ({ ...g, members: [...g.members] }));
}

function close() {
	dupModalOpen.set(false);
}

function setKeepPath(group: DuplicateGroup, path: string) {
	group.keepPath = path;
	localGroups = [...localGroups];
}

async function resolveGroup(hash: string) {
	const group = localGroups.find((g) => g.hash === hash);
	if (!group || group.members.length < 2) return;
	const toDelete = group.members.filter((m) => m.path !== group.keepPath);
	if (!toDelete.length) return;
	const keepName = group.members.find((m) => m.path === group.keepPath)?.name || 'selected copy';
	if (!confirm(`Delete ${toDelete.length} duplicate file(s) and keep "${keepName}"?`)) return;
	let fail = 0;
	for (const item of toDelete) {
		const res = await deleteVideoPath(item.path);
		if (!res.ok) fail++;
	}
	await refresh();
	localGroups = get(duplicateGroups).map((g) => ({ ...g, members: [...g.members] }));
	if (fail) toast(fail + ' duplicate delete(s) failed', 'error');
	else toast('Duplicate group resolved', 'success');
}

async function resolveAll() {
	const actions = localGroups.map((g) => ({
		toDelete: g.members.filter((m) => m.path !== g.keepPath),
	}));
	const totalDeletes = actions.reduce((acc, g) => acc + g.toDelete.length, 0);
	if (!totalDeletes) return;
	if (!confirm(`Resolve all groups by deleting ${totalDeletes} file(s) and keeping your selected copy in each group?`)) return;
	let fail = 0;
	for (const action of actions) {
		for (const item of action.toDelete) {
			const res = await deleteVideoPath(item.path);
			if (!res.ok) fail++;
		}
	}
	await refresh();
	localGroups = get(duplicateGroups).map((g) => ({ ...g, members: [...g.members] }));
	if (fail) toast(fail + ' duplicate delete(s) failed', 'error');
	else toast('All duplicate groups resolved', 'success');
}
</script>

<Modal id="dup-modal" open={$dupModalOpen} on:close={close}>
	<div class="dup-inner">
		<div class="dup-header">
			<div>
				<div class="dup-title">Duplicate manager</div>
				<div class="dup-subtitle" id="dup-subtitle">
					{localGroups.length
						? localGroups.length + ' group' + (localGroups.length === 1 ? '' : 's') + ' need review'
						: 'No duplicates right now.'}
				</div>
			</div>
			<button class="close-btn" id="dup-close" on:click={close}>✕</button>
		</div>
		<div class="dup-body" id="dup-groups">
			{#if !localGroups.length}
				<div class="dup-empty">Everything is unique. Nice and tidy.</div>
			{:else}
				{#each localGroups as group, idx (group.hash)}
					<section class="dup-group" data-hash={group.hash}>
						<div class="dup-group-header">
							<div class="dup-group-title">Group {idx + 1} ({group.members.length} copies)</div>
							<button class="btn dup-resolve-btn" on:click={() => resolveGroup(group.hash)}>Delete non-kept</button>
						</div>
						<div class="dup-items">
							{#each group.members as v (v.path)}
								<label class="dup-item" class:keep={group.keepPath === v.path}>
									<input
										type="radio"
										name="dup-keep-{group.hash}"
										value={v.path}
										checked={group.keepPath === v.path}
										on:change={() => setKeepPath(group, v.path)}
									/>
									<div class="dup-item-meta">
										<span class="dup-item-name">{v.name}</span>
										<span class="dup-item-path">{v.folder || '/'} • {formatBytes(v.size || 0)} • {formatModified(v.modified)}</span>
									</div>
								</label>
							{/each}
						</div>
					</section>
				{/each}
			{/if}
		</div>
		<div class="dup-footer">
			<button class="btn" id="dup-cancel-btn" on:click={close}>Close</button>
			<button class="btn active" id="dup-resolve-all-btn" disabled={!localGroups.length} on:click={resolveAll}>Resolve all groups</button>
		</div>
	</div>
</Modal>
