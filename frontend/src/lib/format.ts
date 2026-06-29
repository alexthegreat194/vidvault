export function formatBytes(bytes: number): string {
	const n = Number(bytes) || 0;
	if (n < 1024) return n + ' B';
	const units = ['KB', 'MB', 'GB', 'TB'];
	let value = n;
	let idx = -1;
	while (value >= 1024 && idx < units.length - 1) {
		value /= 1024;
		idx++;
	}
	return value.toFixed(value >= 10 ? 0 : 1) + ' ' + units[idx];
}

export function formatModified(iso: string): string {
	if (!iso) return 'unknown';
	const d = new Date(iso);
	if (Number.isNaN(d.getTime())) return 'unknown';
	return d.toLocaleString();
}

export function escHtml(s: string): string {
	return s
		.replace(/&/g, '&amp;')
		.replace(/</g, '&lt;')
		.replace(/>/g, '&gt;')
		.replace(/"/g, '&quot;');
}

export function videoUrl(path: string): string {
	return '/video?path=' + encodeURIComponent(path) + '#t=2';
}
