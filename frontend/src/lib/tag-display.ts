import { escHtml } from './format';
import type { Tag } from './types';

export function renderTagChip(tag: Tag): string {
	return (
		'<span class="tag-chip" style="--tag-color:' +
		escHtml(tag.color || '#888888') +
		'">' +
		escHtml(tag.name) +
		'</span>'
	);
}

export function renderVideoTagChipsInner(tagIDs: string[], tagMap: Record<string, Tag>): string {
	if (!Array.isArray(tagIDs) || !tagIDs.length) return '';
	const chips = tagIDs
		.map((id) => tagMap[id])
		.filter(Boolean)
		.map((tag) => renderTagChip(tag))
		.join('');
	return chips || '';
}
