export const CURRENT_PHASE = 5;
export const CURRENT_TITAN_PHASE = 5;
export const TITAN_PHASE_COUNT = 11;
export const TITAN_CUSTOM_ITEM_MIN_ID = 200000;

const TITAN_PHASE_STORAGE_KEY = 'wowsims-titan-phase';

export function getPreferredTitanPhase(): number {
	try {
		const raw = window.localStorage.getItem(TITAN_PHASE_STORAGE_KEY);
		const n = raw ? parseInt(raw, 10) : CURRENT_TITAN_PHASE;
		if (n >= 1 && n <= TITAN_PHASE_COUNT) {
			return n;
		}
	} catch (_e) {
		// Ignore storage errors (private mode, etc).
	}
	return CURRENT_TITAN_PHASE;
}

export function setPreferredTitanPhase(phase: number): void {
	if (phase < 1 || phase > TITAN_PHASE_COUNT) {
		return;
	}
	try {
		window.localStorage.setItem(TITAN_PHASE_STORAGE_KEY, String(phase));
	} catch (_e) {
		// Ignore storage errors (private mode, etc).
	}
}

// Github pages serves our site under the /wotlk directory (because the repo name is wotlk)
export const REPO_NAME = 'wotlk';

// Get 'elemental_shaman', the pathname part after the repo name
const pathnameParts = window.location.pathname.split('/');
const repoPartIdx = pathnameParts.findIndex(part => part == REPO_NAME);
export const SPEC_DIRECTORY = repoPartIdx == -1 ? '' : pathnameParts[repoPartIdx + 1];
