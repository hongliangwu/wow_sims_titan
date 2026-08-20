import {
	UIEnchant as Enchant,
} from '../proto/ui.js';
import { getLanguageCode } from '../constants/lang.js';
import { Database } from './database.js';

let descriptionsPromise: Promise<Record<number, string>> | null = null;
function fetchEnchantDescriptions(): Promise<Record<number, string>> {
	if (descriptionsPromise == null) {
		descriptionsPromise = fetch('/wotlk/assets/enchants/descriptions.json')
			.then(response => response.json())
			.then(json => {
				const descriptionsMap: Record<number, string> = {};
				for (let idStr in json) {
					descriptionsMap[parseInt(idStr)] = json[idStr];
				}
				return descriptionsMap;
			});
	}
	return descriptionsPromise;
}

async function getLocalizedEnchantName(enchant: Enchant): Promise<string> {
	if (!getLanguageCode()) {
		return '';
	}
	if (enchant.spellId) {
		const data = await Database.getSpellIconData(enchant.spellId);
		if (data.name) {
			return data.name;
		}
	}
	if (enchant.itemId) {
		const data = await Database.getItemIconData(enchant.itemId);
		if (data.name) {
			return data.name;
		}
	}
	return '';
}

export async function getEnchantDescription(enchant: Enchant): Promise<string> {
	const localized = await getLocalizedEnchantName(enchant);
	if (localized) {
		return localized;
	}
	const descriptionsMap = await fetchEnchantDescriptions();
	return descriptionsMap[enchant.effectId] || enchant.name;
}

// Returns a string uniquely identifying the enchant.
export function getUniqueEnchantString(enchant: Enchant): string {
	return enchant.effectId + '-' + enchant.type;
}
