import type {
	AuthStatus,
	Collection,
	CollectionsResponse,
	FolderInfo,
	MutationResult,
	ScanProgress,
	Tag,
	TagsResponse,
	UploadResult,
	Video,
} from './types';

async function jsonFetch<T>(url: string, init?: RequestInit): Promise<T> {
	const res = await fetch(url, init);
	if (!res.ok) throw new Error(await res.text());
	return res.json() as Promise<T>;
}

export async function fetchAuthStatus(): Promise<AuthStatus> {
	const res = await fetch('/api/auth/status');
	return res.json();
}

export async function submitPin(pin: string): Promise<{ ok: boolean }> {
	const res = await fetch('/api/auth/pin', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ pin }),
	});
	return res.json();
}

export async function lockSession(): Promise<boolean> {
	const res = await fetch('/api/auth/lock', { method: 'POST' });
	return res.ok;
}

export async function fetchFolders(): Promise<FolderInfo[]> {
	return jsonFetch('/api/folders');
}

export async function fetchTags(): Promise<TagsResponse> {
	return jsonFetch('/api/tags');
}

export async function fetchCollections(): Promise<CollectionsResponse> {
	return jsonFetch('/api/collections');
}

export async function fetchVideos(): Promise<Video[]> {
	return jsonFetch('/api/videos');
}

export async function fetchTagAssignmentsForHashes(
	hashes: string[],
): Promise<{ videos: { hash: string; tags: string[] }[] }> {
	const res = await fetch('/api/tags/videos', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ hashes }),
	});
	if (!res.ok) throw new Error('failed to refresh tag assignments');
	return res.json();
}

export async function moveVideoApi(
	path: string,
	dest_folder: string,
): Promise<MutationResult> {
	try {
		const res = await fetch('/api/move', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ path, dest_folder }),
		});
		if (!res.ok) return { ok: false, message: await res.text() };
		const data = await res.json();
		return {
			ok: true,
			message: '',
			path: data.path,
			old_path: data.old_path,
			video: data.video,
		};
	} catch {
		return { ok: false, message: 'request failed' };
	}
}

export async function mkdirApi(folder: string): Promise<Response> {
	return fetch('/api/mkdir', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ folder }),
	});
}

export async function rmdirApi(folder: string): Promise<Response> {
	return fetch('/api/rmdir', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ folder }),
	});
}

export async function deleteVideoPath(path: string): Promise<MutationResult> {
	try {
		const res = await fetch('/api/delete', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ path }),
		});
		if (!res.ok) return { ok: false, message: await res.text() };
		const data = await res.json();
		return { ok: true, message: '', path: data.path || path };
	} catch {
		return { ok: false, message: 'request failed' };
	}
}

export async function renameVideoPath(
	path: string,
	name: string,
): Promise<MutationResult> {
	try {
		const res = await fetch('/api/rename', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ path, name }),
		});
		if (!res.ok) return { ok: false, message: await res.text() };
		const data = await res.json();
		return {
			ok: true,
			message: '',
			path: data.path,
			old_path: data.old_path,
			video: data.video,
		};
	} catch {
		return { ok: false, message: 'request failed' };
	}
}

export async function setFavorite(hash: string, favorite: boolean): Promise<boolean> {
	if (!hash) return false;
	try {
		const res = await fetch('/api/favorites/set', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ hash, favorite }),
		});
		return res.ok;
	} catch {
		return false;
	}
}

export async function createTag(name: string): Promise<Tag | null> {
	try {
		const res = await fetch('/api/tags/create', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ name }),
		});
		if (!res.ok) return null;
		return res.json();
	} catch {
		return null;
	}
}

export async function deleteTag(tagID: string): Promise<boolean> {
	try {
		const res = await fetch('/api/tags/delete', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ tag_id: tagID }),
		});
		return res.ok;
	} catch {
		return false;
	}
}

export async function setTagAssignmentsBulk(
	hashes: string[],
	tagID: string,
	assigned: boolean,
): Promise<boolean> {
	try {
		const uniqueHashes = [...new Set(hashes.filter(Boolean))];
		if (!uniqueHashes.length) return true;
		const res = await fetch('/api/tags/assign/bulk', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ hashes: uniqueHashes, tag_id: tagID, assigned }),
		});
		return res.ok;
	} catch {
		return false;
	}
}

export async function createCollection(name: string): Promise<Collection | null> {
	try {
		const res = await fetch('/api/collections/create', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ name }),
		});
		if (!res.ok) return null;
		return res.json();
	} catch {
		return null;
	}
}

export async function renameCollection(id: string, name: string): Promise<boolean> {
	try {
		const res = await fetch('/api/collections/rename', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ id, name }),
		});
		return res.ok;
	} catch {
		return false;
	}
}

export async function deleteCollection(id: string): Promise<boolean> {
	try {
		const res = await fetch('/api/collections/delete', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ id }),
		});
		return res.ok;
	} catch {
		return false;
	}
}

export async function setCollectionVideo(
	id: string,
	hash: string,
	assigned: boolean,
): Promise<boolean> {
	try {
		const res = await fetch('/api/collections/videos/set', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ id, hash, assigned }),
		});
		return res.ok;
	} catch {
		return false;
	}
}

export async function bulkAddCollectionVideos(id: string, hashes: string[]): Promise<boolean> {
	try {
		const res = await fetch('/api/collections/videos/bulk', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ id, hashes }),
		});
		return res.ok;
	} catch {
		return false;
	}
}

export async function createCollectionFromFolder(
	folder: string,
	name: string,
): Promise<{ ok: boolean; error?: string; collection?: Collection }> {
	try {
		const res = await fetch('/api/collections/from-folder', {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ folder, name }),
		});
		if (!res.ok) return { ok: false, error: await res.text() };
		return { ok: true, collection: await res.json() };
	} catch {
		return { ok: false, error: 'request failed' };
	}
}

export async function clearNewStatus(paths: string[], all = false): Promise<boolean> {
	const res = await fetch('/api/videos/new/clear', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ all, paths }),
	});
	return res.ok;
}

export async function forgetVideos(
	paths: string[],
	hashes: string[] = [],
	all = false,
): Promise<boolean> {
	const res = await fetch('/api/videos/new/forget', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ all, paths, hashes }),
	});
	return res.ok;
}

export async function uploadFile(file: File, folder: string): Promise<UploadResult[]> {
	const fd = new FormData();
	fd.append('file', file);
	fd.append('folder', folder);
	const res = await fetch('/api/upload', { method: 'POST', body: fd });
	return res.json();
}

export interface StreamVideosOptions {
	rescan?: boolean;
	onProgress?: (progress: ScanProgress) => void;
	onRemoved?: (path: string) => void;
}

export async function streamVideos(
	onVideo: (video: Video) => void,
	opts: StreamVideosOptions = {},
): Promise<Video[]> {
	const url = opts.rescan ? '/api/videos/stream?rescan=1' : '/api/videos/stream';
	const res = await fetch(url);
	if (!res.ok || !res.body) {
		const videos = await fetchVideos();
		const seen = new Set<string>();
		for (const v of videos) {
			if (!v?.path || seen.has(v.path)) continue;
			seen.add(v.path);
			onVideo(v);
		}
		return videos;
	}
	const reader = res.body.getReader();
	const decoder = new TextDecoder();
	let buffer = '';
	const streamed: Video[] = [];
	const seen = new Set<string>();
	while (true) {
		const { value, done } = await reader.read();
		if (done) break;
		buffer += decoder.decode(value, { stream: true });
		const lines = buffer.split('\n');
		buffer = lines.pop() || '';
		for (const line of lines) {
			const trimmed = line.trim();
			if (!trimmed) continue;
			try {
				const event = JSON.parse(trimmed);
				if (event.type === 'video' && event.payload?.path) {
					const v = event.payload as Video;
					if (seen.has(v.path)) continue;
					seen.add(v.path);
					streamed.push(v);
					onVideo(v);
				} else if (event.type === 'scan_progress' && event.payload) {
					opts.onProgress?.(event.payload as ScanProgress);
				} else if (event.type === 'video_removed' && event.payload?.path) {
					opts.onRemoved?.(event.payload.path as string);
				}
			} catch {
				/* skip bad line */
			}
		}
	}
	return streamed;
}
