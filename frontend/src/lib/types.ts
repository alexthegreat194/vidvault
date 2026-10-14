export interface Video {
	name: string;
	path: string;
	folder: string;
	ext: string;
	size: number;
	modified: string;
	hash: string;
	is_favorite: boolean;
	is_new: boolean;
	batch_id?: string;
	tags?: string[];
}

export interface UploadBatch {
	id: string;
	name: string;
	created_at: string;
	needs_sorting: boolean;
	source: 'upload' | 'migrated' | 'discovered';
	video_count: number;
	new_count: number;
}

export interface FolderInfo {
	name: string;
	has_other_files: boolean;
}

export interface Tag {
	id: string;
	name: string;
	color?: string;
}

export interface TagsResponse {
	tags: Tag[];
	assignments: Record<string, string[]>;
}

export interface Collection {
	id: string;
	name: string;
	video_hashes: string[];
	created_at: string;
	updated_at: string;
}

export interface CollectionsResponse {
	collections: Collection[];
}

export interface DuplicateGroup {
	hash: string;
	members: Video[];
	keepPath: string;
}

export interface AuthStatus {
	pinRequired: boolean;
	authenticated: boolean;
}

export interface ScanProgress {
	found: number;
	hashed: number;
	cached: number;
}

export interface StreamEvent {
	type: string;
	payload?: Video | ScanProgress | { count?: number; path?: string; message?: string };
}

export interface UploadResult {
	name: string;
	error?: string;
	video?: Video;
}

export interface MutationResult {
	ok: boolean;
	message: string;
	path?: string;
	old_path?: string;
	video?: Video;
}

export interface UploadFileItem {
	file: File;
	id: string;
	status: 'pending' | 'uploading' | 'done' | 'error';
	error: string;
}

export type ToastType = 'success' | 'error' | '';
