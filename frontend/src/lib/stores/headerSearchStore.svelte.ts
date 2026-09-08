export type HeaderSearchMode = 'expanded' | 'compact';

function createHeaderSearchStore() {
	let mode = $state<HeaderSearchMode>('expanded');
	let locked = $state(false);

	return {
		get mode() {
			return mode;
		},
		get isExpanded() {
			return mode === 'expanded';
		},
		get isCompact() {
			return mode === 'compact';
		},
		get isLocked() {
			return locked;
		},
		setExpanded(manual = false) {
			mode = 'expanded';
			locked = manual;
		},
		setCompact() {
			mode = 'compact';
			locked = false;
		},
		setLocked(v: boolean) {
			locked = v;
		},
		reset() {
			mode = 'expanded';
			locked = false;
		}
	};
}

export const headerSearchStore = createHeaderSearchStore();
