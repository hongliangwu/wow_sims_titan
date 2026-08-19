export const wowheadSupportedLanguages: Record<string, string> = {
	'en': 'English',
	'cn': '简体中文',
	'de': 'Deutsch',
	'es': 'Español',
	'fr': 'Français',
	'it': 'Italiano',
	'ko': '한국어',
	'pt': 'Português Brasileiro',
	'ru': 'Русский',
};

// Wowhead nether tooltip locale IDs (locale=4 is 简体中文).
export const wowheadLocaleIds: Record<string, number> = {
	'en': 0,
	'ko': 1,
	'fr': 2,
	'de': 3,
	'cn': 4,
	'es': 6,
	'ru': 7,
	'pt': 8,
	'it': 9,
};

// v2 ignores the old default of 'en' stored in wowsims-language.
const LANGUAGE_STORAGE_KEY = 'wowsims-language-v2';
export const DEFAULT_LANGUAGE = 'cn';

export function normalizeLanguageCode(lang: string): string {
	const raw = (lang || '').trim().toLowerCase();
	const code = raw.includes('-') ? raw.split('-')[0] : raw.substring(0, 2);
	if (code === 'zh') {
		return 'cn';
	}
	if (Object.keys(wowheadSupportedLanguages).includes(code)) {
		return code;
	}
	return DEFAULT_LANGUAGE;
}

export function getPreferredLanguage(): string {
	try {
		return normalizeLanguageCode(localStorage.getItem(LANGUAGE_STORAGE_KEY) || DEFAULT_LANGUAGE);
	} catch {
		return DEFAULT_LANGUAGE;
	}
}

export function getBrowserLanguageCode(): string {
	return normalizeLanguageCode(navigator.language || '');
}

export function getLanguageCode(): string {
	return cachedLanguageCode_;
}

export function getWowheadLanguagePrefix(): string {
	return cachedWowheadLanguagePrefix_;
}

export function getWowheadLocaleId(): number {
	const lang = getLanguageCode() || 'en';
	return wowheadLocaleIds[lang] ?? 0;
}

export function setLanguageCode(newLang: string) {
	applyLanguageCode(newLang);
	try {
		localStorage.setItem(LANGUAGE_STORAGE_KEY, normalizeLanguageCode(newLang));
	} catch {
		// Ignore storage failures (private mode, etc).
	}
}

function applyLanguageCode(newLang: string) {
	const normalized = normalizeLanguageCode(newLang);
	// Use '' instead of 'en' because wowhead doesn't like having the en/ prefix.
	cachedLanguageCode_ = normalized == 'en' ? '' : normalized;
	cachedWowheadLanguagePrefix_ = cachedLanguageCode_ ? cachedLanguageCode_ + '/' : '';
	try {
		document.documentElement.lang = normalized === 'cn' ? 'zh-CN' : normalized;
	} catch {
		// Ignore if document is unavailable.
	}
}

let cachedLanguageCode_: string = '';
let cachedWowheadLanguagePrefix_: string = '';

try {
	const stored = localStorage.getItem(LANGUAGE_STORAGE_KEY);
	applyLanguageCode(stored || DEFAULT_LANGUAGE);
} catch {
	applyLanguageCode(DEFAULT_LANGUAGE);
}
