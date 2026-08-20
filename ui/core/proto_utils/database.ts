import {
	EquipmentSpec,
	GemColor,
	ItemSlot,
	ItemSpec,
	ItemSwap,
	PresetEncounter,
	PresetTarget,
	SimDatabase,
} from '../proto/common.js';
import {
	GlyphID,
	IconData,
	UIDatabase,
	UIEnchant as Enchant,
	UIGem as Gem,
	UIItem as Item,
	UINPC as Npc,
	UIZone as Zone,
} from '../proto/ui.js';

import {
	getEligibleEnchantSlots,
	getEligibleItemSlots,
} from './utils.js';
import { gemEligibleForSocket, gemMatchesSocket, isEpicQualityGem } from './gems.js';
import { EquippedItem } from './equipped_item.js';
import { Gear, ItemSwapGear } from './gear.js';
import { CHARACTER_LEVEL } from '../constants/mechanics.js';
import { getLanguageCode, getWowheadLocaleId } from '../constants/lang.js';
import { distinct } from '../utils.js';
import { registerTitanItemEffects, registerTitanItemIds, registerTitanItemSets } from './item_tooltip.js';

const dbUrlJson = '/wotlk/assets/database/db.json';
const dbUrlBin = '/wotlk/assets/database/db.bin';
const leftoversUrlJson = '/wotlk/assets/database/leftover_db.json';
const leftoversUrlBin = '/wotlk/assets/database/leftover_db.bin';
const titanSetsUrlJson = '/wotlk/assets/database/titan_sets.json';
// When changing this value, don't forget to change the html <link> for preloading!
const READ_JSON = true;

const titanSpellIconData: Record<number, { id: number, name: string, icon: string }> = {
	1299093: { id: 1299093, name: '圣光裁决', icon: 'ability_paladin_judgementred' },
	1299096: { id: 1299096, name: '强化圣光裁决', icon: 'spell_holy_holyguidance' },
	1298728: { id: 1298728, name: '祈求圣光', icon: 'spell_holy_holybolt' },
	1298723: { id: 1298723, name: '圣光裁决', icon: 'ability_paladin_judgementred' },
	1298724: { id: 1298724, name: '祈求圣光', icon: 'spell_holy_holybolt' },
	1298725: { id: 1298725, name: '虔诚', icon: 'spell_holy_holybolt' },
	1299090: { id: 1299090, name: '正义', icon: 'spell_holy_righteousnessaura' },
	1299075: { id: 1299075, name: '圣光之怒', icon: 'spell_holy_searinglight' },
	1299086: { id: 1299086, name: '无法祈求圣光', icon: 'spell_holy_removekurse' },
};


export class Database {
	private static loadPromise: Promise<Database> | null = null;
	static get(): Promise<Database> {
		if (Database.loadPromise == null) {
			if (READ_JSON) {
				Database.loadPromise = Promise.all([
					fetch(dbUrlJson).then(response => response.json()),
					fetch(titanSetsUrlJson).then(response => (response.ok ? response.json() : [])).catch(() => []),
				]).then(([json, titanCatalog]) => {
					const sets = Array.isArray(titanCatalog) ? titanCatalog : titanCatalog?.sets;
					const itemIds = Array.isArray(titanCatalog) ? [] : titanCatalog?.itemIds;
					const effects = Array.isArray(titanCatalog) ? undefined : titanCatalog?.effects;
					registerTitanItemSets(sets, json.items);
					registerTitanItemIds(itemIds);
					registerTitanItemEffects(effects);
					return new Database(UIDatabase.fromJson(json));
				});
			} else {
				Database.loadPromise = fetch(dbUrlBin)
					.then(response => response.arrayBuffer())
					.then(buffer => new Database(UIDatabase.fromBinary(new Uint8Array(buffer))));
			}
		}
		return Database.loadPromise;
	}

	static getLeftovers(): Promise<UIDatabase> {
		if (READ_JSON) {
			return fetch(leftoversUrlJson)
				.then(response => response.json())
				.then(json => UIDatabase.fromJson(json));
		} else {
			return fetch(leftoversUrlBin)
				.then(response => response.arrayBuffer())
				.then(buffer => UIDatabase.fromBinary(new Uint8Array(buffer)));
		}
	}

	// Checks if any items in the equipment are missing from the current DB. If so, loads the leftover DB.
	static async loadLeftoversIfNecessary(equipment: EquipmentSpec): Promise<Database> {
		const db = await Database.get();
		if (db.loadedLeftovers) {
			return db;
		}

		const shouldLoadLeftovers = equipment.items.some(item => item.id != 0 && !db.items.has(item.id));
		if (shouldLoadLeftovers) {
			const leftoverDb = await Database.getLeftovers();
			db.loadProto(leftoverDb);
			db.loadedLeftovers = true;
		}
		return db;
	}

	private readonly items = new Map<number, Item>();
	private readonly enchantsBySlot: Partial<Record<ItemSlot, Enchant[]>> = {};
	private readonly gems = new Map<number, Gem>();
	private readonly npcs = new Map<number, Npc>();
	private readonly zones = new Map<number, Zone>();
	private readonly presetEncounters = new Map<string, PresetEncounter>();
	private readonly presetTargets = new Map<string, PresetTarget>();
	private readonly itemIcons: Record<number, Promise<IconData>> = {};
	private readonly spellIcons: Record<number, Promise<IconData>> = {};
	private readonly localizedItemIcons: Record<string, Promise<IconData>> = {};
	private readonly localizedSpellIcons: Record<string, Promise<IconData>> = {};
	private readonly glyphIds: Array<GlyphID> = [];
	private loadedLeftovers: boolean = false;

	private constructor(db: UIDatabase) {
		this.loadProto(db);
	}

	// Add all data from the db proto into this database.
	private loadProto(db: UIDatabase) {
		db.items.forEach(item => this.items.set(item.id, item));
		db.enchants.forEach(enchant => {
			const slots = getEligibleEnchantSlots(enchant);
			slots.forEach(slot => {
				if (!this.enchantsBySlot[slot]) {
					this.enchantsBySlot[slot] = [];
				}
				this.enchantsBySlot[slot]!.push(enchant);
			});
		});
		db.gems.forEach(gem => this.gems.set(gem.id, gem));

		db.npcs.forEach(npc => this.npcs.set(npc.id, npc));
		db.zones.forEach(zone => this.zones.set(zone.id, zone));
		db.encounters.forEach(encounter => this.presetEncounters.set(encounter.path, encounter));
		db.encounters.map(e => e.targets).flat().forEach(target => this.presetTargets.set(target.path, target));

		db.items.forEach(item => this.itemIcons[item.id] = Promise.resolve(IconData.create({
			id: item.id,
			name: item.name,
			icon: item.icon,
		})));
		db.gems.forEach(gem => this.itemIcons[gem.id] = Promise.resolve(IconData.create({
			id: gem.id,
			name: gem.name,
			icon: gem.icon,
		})));
		db.itemIcons.forEach(data => this.itemIcons[data.id] = Promise.resolve(data));
		db.spellIcons.forEach(data => this.spellIcons[data.id] = Promise.resolve(data));
		db.glyphIds.forEach(id => this.glyphIds.push(id));
	}

	getAllItems(): Array<Item> {
		return Array.from(this.items.values());
	}

	getItems(slot: ItemSlot): Array<Item> {
		return this.getAllItems().filter(item => getEligibleItemSlots(item).includes(slot));
	}

	getEnchants(slot: ItemSlot): Array<Enchant> {
		return this.enchantsBySlot[slot] || [];
	}

	getGems(socketColor?: GemColor): Array<Gem> {
		const selectable = (g: Gem) => !isEpicQualityGem(g);
		if (!socketColor)
			return Array.from(this.gems.values()).filter(selectable);

		let ret = new Array();
		for (let g of this.gems.values()){
			if (selectable(g) && gemEligibleForSocket(g, socketColor))
				ret.push(g);
		}
		return ret;
	}

	getNpc(npcId: number): Npc | null {
		return this.npcs.get(npcId) || null;
	}
	getZone(zoneId: number): Zone | null {
		return this.zones.get(zoneId) || null;
	}

	getMatchingGems(socketColor: GemColor): Array<Gem> {
		let ret = new Array();
		for (let g of this.gems.values()){
			if (!isEpicQualityGem(g) && gemMatchesSocket(g, socketColor))
				ret.push(g);
		}
		return ret;
	}

	lookupGem(itemID: number): Gem | null {
		return this.gems.get(itemID) || null;
	}

	lookupItemSpec(itemSpec: ItemSpec): EquippedItem | null {
		const item = this.items.get(itemSpec.id);
		if (!item)
			return null;

		let enchant: Enchant | null = null;
		if (itemSpec.enchant) {
			const slots = getEligibleItemSlots(item);
			for (let i = 0; i < slots.length; i++) {
				enchant = (this.enchantsBySlot[slots[i]] || [])
					.find(enchant => [enchant.effectId, enchant.itemId, enchant.spellId].includes(itemSpec.enchant)) || null;
				if (enchant) {
					break;
				}
			}
		}

		const gems = itemSpec.gems.map(gemId => this.lookupGem(gemId));

		return new EquippedItem(item, enchant, gems);
	}

	lookupEquipmentSpec(equipSpec: EquipmentSpec): Gear {
		// EquipmentSpec is supposed to be indexed by slot, but here we assume
		// it isn't just in case.
		const gearMap: Partial<Record<ItemSlot, EquippedItem | null>> = {};

		equipSpec.items.forEach(itemSpec => {
			const item = this.lookupItemSpec(itemSpec);
			if (!item)
				return;

			const itemSlots = getEligibleItemSlots(item.item);

			const assignedSlot = itemSlots.find(slot => !gearMap[slot]);
			if (assignedSlot == null)
				throw new Error('No slots left to equip ' + Item.toJsonString(item.item));

			gearMap[assignedSlot] = item;
		});

		return new Gear(gearMap);
	}

	lookupItemSwap(itemSwap: ItemSwap): ItemSwapGear {
		return new ItemSwapGear({
			[ItemSlot.ItemSlotMainHand]: itemSwap.mhItem ? this.lookupItemSpec(itemSwap.mhItem): null,
			[ItemSlot.ItemSlotOffHand]: itemSwap.ohItem ? this.lookupItemSpec(itemSwap.ohItem): null,
			[ItemSlot.ItemSlotRanged]: itemSwap.rangedItem ? this.lookupItemSpec(itemSwap.rangedItem): null,
		});
	}

	enchantSpellIdToEffectId(enchantSpellId: number): number {
		const enchant = Object.values(this.enchantsBySlot).flat().find(enchant => enchant.spellId == enchantSpellId);
		return enchant ? enchant.effectId : 0;
	}

	glyphItemToSpellId(itemId: number): number {
		return this.glyphIds.find(gid => gid.itemId == itemId)?.spellId || 0;
	}
	glyphSpellToItemId(spellId: number): number {
		return this.glyphIds.find(gid => gid.spellId == spellId)?.itemId || 0;
	}

	getPresetEncounter(path: string): PresetEncounter | null {
		return this.presetEncounters.get(path) || null;
	}
	getPresetTarget(path: string): PresetTarget | null {
		return this.presetTargets.get(path) || null;
	}
	getAllPresetEncounters(): Array<PresetEncounter> {
		return Array.from(this.presetEncounters.values());
	}
	getAllPresetTargets(): Array<PresetTarget> {
		return Array.from(this.presetTargets.values());
	}

	static async getItemIconData(itemId: number): Promise<IconData> {
		const db = await Database.get();
		const dbItem = db.items.get(itemId);
		if (itemId >= 200000 && dbItem) {
			return IconData.create({
				id: itemId,
				name: dbItem.name,
				icon: dbItem.icon,
			});
		}
		const lang = getLanguageCode();
		if (lang) {
			const key = `${lang}:${itemId}`;
			if (!db.localizedItemIcons[key]) {
				db.localizedItemIcons[key] = Database.getWowheadItemTooltipData(itemId);
			}
			return await db.localizedItemIcons[key];
		}
		if (!db.itemIcons[itemId]) {
			db.itemIcons[itemId] = Database.getWowheadItemTooltipData(itemId);
		}
		return await db.itemIcons[itemId];
	}

	static async getSpellIconData(spellId: number): Promise<IconData> {
		const titanSpell = titanSpellIconData[spellId];
		if (titanSpell) {
			return IconData.create(titanSpell);
		}
		const db = await Database.get();
		const lang = getLanguageCode();
		if (lang) {
			const key = `${lang}:${spellId}`;
			if (!db.localizedSpellIcons[key]) {
				db.localizedSpellIcons[key] = Database.getWowheadSpellTooltipData(spellId);
			}
			return await db.localizedSpellIcons[key];
		}
		if (!db.spellIcons[spellId]) {
			db.spellIcons[spellId] = Database.getWowheadSpellTooltipData(spellId);
		}
		return await db.spellIcons[spellId];
	}

	private static async getWowheadItemTooltipData(id: number): Promise<IconData> {
		return Database.getWowheadTooltipData(id, 'item');
	}
	private static async getWowheadSpellTooltipData(id: number): Promise<IconData> {
		return Database.getWowheadTooltipData(id, 'spell');
	}
	private static async getWowheadTooltipData(id: number, tooltipPostfix: string): Promise<IconData> {
		const locale = getWowheadLocaleId();
		const localeQuery = locale ? `&locale=${locale}` : '';
		const url = `https://nether.wowhead.com/wotlk/tooltip/${tooltipPostfix}/${id}?lvl=${CHARACTER_LEVEL}${localeQuery}`;
		try {
			const response = await fetch(url);
			const json = await response.json();
			return IconData.create({
				id: id,
				name: json['name'],
				icon: json['icon'],
			});
		} catch (e) {
			console.error('Error while fetching url: ' + url + '\n\n' + e);
			return IconData.create();
		}
	}

	public static mergeSimDatabases(db1: SimDatabase, db2: SimDatabase): SimDatabase {
		return SimDatabase.create({
			items: distinct(db1.items.concat(db2.items), (a, b) => a.id == b.id),
			enchants: distinct(db1.enchants.concat(db2.enchants), (a, b) => a.effectId == b.effectId),
			gems: distinct(db1.gems.concat(db2.gems), (a, b) => a.id == b.id),
		})
	}
}
