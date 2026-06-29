export function movedVideoRelPath(srcPath: string, destFolder: string): string {
	const slash = srcPath.lastIndexOf('/');
	const base = slash >= 0 ? srcPath.slice(slash + 1) : srcPath;
	let dest = destFolder == null ? '' : String(destFolder).trim();
	if (dest === '.' || dest === '/') dest = '';
	const rel = dest
		? dest.replace(/^[/\\]+/, '').replace(/[/\\]+$/, '') + '/' + base
		: base;
	return rel.replace(/\\/g, '/');
}

export function folderFromRelPath(relPath: string): string {
	const i = relPath.lastIndexOf('/');
	if (i < 0) return '/';
	const dir = relPath.slice(0, i);
	return dir === '' ? '/' : dir.replace(/\\/g, '/');
}
