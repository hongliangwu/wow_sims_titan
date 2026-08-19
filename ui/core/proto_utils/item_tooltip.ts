import { Tooltip } from 'bootstrap';

import { getLanguageCode } from '../constants/lang.js';
import { TITAN_CUSTOM_ITEM_MIN_ID } from '../constants/other.js';
import {
	ArmorType,
	Class,
	GemColor,
	HandType,
	ItemQuality,
	ItemType,
	RangedWeaponType,
	Stat,
	WeaponType,
} from '../proto/common.js';
import { UIItem as Item } from '../proto/ui.js';

const qualityColors: Record<ItemQuality, string> = {
	[ItemQuality.ItemQualityJunk]: '#9d9d9d',
	[ItemQuality.ItemQualityCommon]: '#ffffff',
	[ItemQuality.ItemQualityUncommon]: '#1eff00',
	[ItemQuality.ItemQualityRare]: '#0070dd',
	[ItemQuality.ItemQualityEpic]: '#a335ee',
	[ItemQuality.ItemQualityLegendary]: '#ff8000',
	[ItemQuality.ItemQualityArtifact]: '#e6cc80',
	[ItemQuality.ItemQualityHeirloom]: '#e6cc80',
};

function escapeHtml(s: string): string {
	return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

function isCn(): boolean {
	return getLanguageCode() === 'cn';
}

function t(cn: string, en: string): string {
	return isCn() ? cn : en;
}

function slotLabel(item: Item): string {
	const cn = isCn();
	switch (item.type) {
		case ItemType.ItemTypeHead: return cn ? '头部' : 'Head';
		case ItemType.ItemTypeNeck: return cn ? '颈部' : 'Neck';
		case ItemType.ItemTypeShoulder: return cn ? '肩部' : 'Shoulder';
		case ItemType.ItemTypeBack: return cn ? '背部' : 'Back';
		case ItemType.ItemTypeChest: return cn ? '胸部' : 'Chest';
		case ItemType.ItemTypeWrist: return cn ? '手腕' : 'Wrist';
		case ItemType.ItemTypeHands: return cn ? '手' : 'Hands';
		case ItemType.ItemTypeWaist: return cn ? '腰部' : 'Waist';
		case ItemType.ItemTypeLegs: return cn ? '腿部' : 'Legs';
		case ItemType.ItemTypeFeet: return cn ? '脚' : 'Feet';
		case ItemType.ItemTypeFinger: return cn ? '手指' : 'Finger';
		case ItemType.ItemTypeTrinket: return cn ? '饰品' : 'Trinket';
		case ItemType.ItemTypeWeapon:
			switch (item.handType) {
				case HandType.HandTypeMainHand: return cn ? '主手' : 'Main Hand';
				case HandType.HandTypeOffHand: return cn ? '副手' : 'Off Hand';
				case HandType.HandTypeTwoHand: return cn ? '双手' : 'Two-Hand';
				default: return cn ? '单手' : 'One-Hand';
			}
		case ItemType.ItemTypeRanged: return cn ? '远程' : 'Ranged';
		default: return '';
	}
}

function typeLabel(item: Item): string {
	const cn = isCn();
	if (item.armorType) {
		switch (item.armorType) {
			case ArmorType.ArmorTypeCloth: return cn ? '布甲' : 'Cloth';
			case ArmorType.ArmorTypeLeather: return cn ? '皮甲' : 'Leather';
			case ArmorType.ArmorTypeMail: return cn ? '锁甲' : 'Mail';
			case ArmorType.ArmorTypePlate: return cn ? '板甲' : 'Plate';
		}
	}
	if (item.weaponType) {
		switch (item.weaponType) {
			case WeaponType.WeaponTypeAxe: return cn ? '斧' : 'Axe';
			case WeaponType.WeaponTypeDagger: return cn ? '匕首' : 'Dagger';
			case WeaponType.WeaponTypeFist: return cn ? '拳套' : 'Fist Weapon';
			case WeaponType.WeaponTypeMace: return cn ? '锤' : 'Mace';
			case WeaponType.WeaponTypeOffHand: return cn ? '副手物品' : 'Held In Off-hand';
			case WeaponType.WeaponTypePolearm: return cn ? '长柄武器' : 'Polearm';
			case WeaponType.WeaponTypeShield: return cn ? '盾牌' : 'Shield';
			case WeaponType.WeaponTypeStaff: return cn ? '法杖' : 'Staff';
			case WeaponType.WeaponTypeSword: return cn ? '剑' : 'Sword';
		}
	}
	if (item.rangedWeaponType) {
		switch (item.rangedWeaponType) {
			case RangedWeaponType.RangedWeaponTypeBow: return cn ? '弓' : 'Bow';
			case RangedWeaponType.RangedWeaponTypeCrossbow: return cn ? '弩' : 'Crossbow';
			case RangedWeaponType.RangedWeaponTypeGun: return cn ? '枪械' : 'Gun';
			case RangedWeaponType.RangedWeaponTypeWand: return cn ? '魔杖' : 'Wand';
			case RangedWeaponType.RangedWeaponTypeThrown: return cn ? '投掷武器' : 'Thrown';
			case RangedWeaponType.RangedWeaponTypeIdol: return cn ? '神像' : 'Idol';
			case RangedWeaponType.RangedWeaponTypeLibram: return cn ? '圣契' : 'Libram';
			case RangedWeaponType.RangedWeaponTypeTotem: return cn ? '图腾' : 'Totem';
			case RangedWeaponType.RangedWeaponTypeSigil: return cn ? '魔印' : 'Sigil';
		}
	}
	return '';
}

function socketCss(color: GemColor): string {
	switch (color) {
		case GemColor.GemColorMeta: return 'meta';
		case GemColor.GemColorRed: return 'red';
		case GemColor.GemColorYellow: return 'yellow';
		case GemColor.GemColorBlue: return 'blue';
		case GemColor.GemColorPrismatic: return 'prismatic';
		default: return 'prismatic';
	}
}

function socketLabel(color: GemColor): string {
	const cn = isCn();
	switch (color) {
		case GemColor.GemColorMeta: return cn ? '多彩插槽' : 'Meta Socket';
		case GemColor.GemColorRed: return cn ? '红色插槽' : 'Red Socket';
		case GemColor.GemColorYellow: return cn ? '黄色插槽' : 'Yellow Socket';
		case GemColor.GemColorBlue: return cn ? '蓝色插槽' : 'Blue Socket';
		case GemColor.GemColorPrismatic: return cn ? '棱彩插槽' : 'Prismatic Socket';
		default: return cn ? '插槽' : 'Socket';
	}
}

function classLabel(c: Class): string {
	const cn = isCn();
	switch (c) {
		case Class.ClassWarrior: return cn ? '战士' : 'Warrior';
		case Class.ClassPaladin: return cn ? '圣骑士' : 'Paladin';
		case Class.ClassHunter: return cn ? '猎人' : 'Hunter';
		case Class.ClassRogue: return cn ? '潜行者' : 'Rogue';
		case Class.ClassPriest: return cn ? '牧师' : 'Priest';
		case Class.ClassDeathknight: return cn ? '死亡骑士' : 'Death Knight';
		case Class.ClassShaman: return cn ? '萨满祭司' : 'Shaman';
		case Class.ClassMage: return cn ? '法师' : 'Mage';
		case Class.ClassWarlock: return cn ? '术士' : 'Warlock';
		case Class.ClassDruid: return cn ? '德鲁伊' : 'Druid';
		default: return '';
	}
}

type StatLine = { label: string; value: number };

function getStat(s: number[] | undefined, st: Stat): number {
	return s?.[st] || 0;
}

function whiteStatLines(s: number[] | undefined): StatLine[] {
	const cn = isCn();
	const lines: StatLine[] = [];
	const add = (st: Stat, labelCn: string, labelEn: string) => {
		const v = getStat(s, st);
		if (v) lines.push({ label: cn ? labelCn : labelEn, value: v });
	};
	add(Stat.StatArmor, '护甲', 'Armor');
	add(Stat.StatBonusArmor, '额外护甲', 'Bonus Armor');
	add(Stat.StatStrength, '力量', 'Strength');
	add(Stat.StatAgility, '敏捷', 'Agility');
	add(Stat.StatStamina, '耐力', 'Stamina');
	add(Stat.StatIntellect, '智力', 'Intellect');
	add(Stat.StatSpirit, '精神', 'Spirit');
	add(Stat.StatBlockValue, '格挡值', 'Block Value');
	return lines;
}

function plusGreenStatLines(s: number[] | undefined): StatLine[] {
	const cn = isCn();
	const lines: StatLine[] = [];
	const add = (st: Stat, labelCn: string, labelEn: string) => {
		const v = getStat(s, st);
		if (v) lines.push({ label: cn ? labelCn : labelEn, value: v });
	};
	add(Stat.StatSpellPower, '法术强度', 'Spell Power');
	add(Stat.StatMP5, '每5秒法力回复', 'Mana per 5 sec.');
	add(Stat.StatAttackPower, '攻击强度', 'Attack Power');
	if (getStat(s, Stat.StatRangedAttackPower) && getStat(s, Stat.StatRangedAttackPower) != getStat(s, Stat.StatAttackPower)) {
		add(Stat.StatRangedAttackPower, '远程攻击强度', 'Ranged Attack Power');
	}
	return lines;
}

function equipStatLines(s: number[] | undefined): StatLine[] {
	const cn = isCn();
	const lines: StatLine[] = [];
	const add = (st: Stat, labelCn: string, labelEn: string) => {
		const v = getStat(s, st);
		if (v) lines.push({ label: cn ? labelCn : labelEn, value: v });
	};

	const hit = getStat(s, Stat.StatMeleeHit) || getStat(s, Stat.StatSpellHit);
	if (hit) lines.push({ label: cn ? '命中等级' : 'hit rating', value: hit });
	const crit = getStat(s, Stat.StatMeleeCrit) || getStat(s, Stat.StatSpellCrit);
	if (crit) lines.push({ label: cn ? '爆击等级' : 'critical strike rating', value: crit });
	const haste = getStat(s, Stat.StatMeleeHaste) || getStat(s, Stat.StatSpellHaste);
	if (haste) lines.push({ label: cn ? '急速等级' : 'haste rating', value: haste });

	add(Stat.StatArmorPenetration, '护甲穿透等级', 'armor penetration rating');
	add(Stat.StatExpertise, '精准等级', 'expertise rating');
	add(Stat.StatDefense, '防御等级', 'defense rating');
	add(Stat.StatDodge, '躲闪等级', 'dodge rating');
	add(Stat.StatParry, '招架等级', 'parry rating');
	add(Stat.StatBlock, '格挡等级', 'block rating');
	add(Stat.StatResilience, '韧性等级', 'resilience rating');
	add(Stat.StatSpellPenetration, '法术穿透', 'spell penetration');
	add(Stat.StatFireResistance, '火焰抗性', 'Fire Resistance');
	add(Stat.StatFrostResistance, '冰霜抗性', 'Frost Resistance');
	add(Stat.StatNatureResistance, '自然抗性', 'Nature Resistance');
	add(Stat.StatShadowResistance, '暗影抗性', 'Shadow Resistance');
	add(Stat.StatArcaneResistance, '奥术抗性', 'Arcane Resistance');
	return lines;
}

function socketBonusLines(s: number[] | undefined): StatLine[] {
	return [...whiteStatLines(s), ...plusGreenStatLines(s), ...equipStatLines(s)];
}

function formatEquipLine(line: StatLine): string {
	const n = Math.round(line.value);
	if (isCn()) {
		return `装备：${escapeHtml(line.label)}提高${n}点。`;
	}
	return `Equip: Increases ${escapeHtml(line.label)} by ${n}.`;
}

function formatWhiteStat(line: StatLine): string {
	if (line.label === '护甲' || line.label === 'Armor') {
		return `${Math.round(line.value)}${isCn() ? '护甲' : ' Armor'}`;
	}
	if (line.label === '额外护甲' || line.label === 'Bonus Armor') {
		return `+${Math.round(line.value)} ${escapeHtml(line.label)}`;
	}
	const sign = line.value > 0 ? '+' : '';
	return `${sign}${Math.round(line.value)} ${escapeHtml(line.label)}`;
}

export function titanOriginLabel(phase?: number): string {
	if (phase && phase > 0) {
		return t(`时光 P${phase}`, `Titan P${phase}`);
	}
	return t('时光服', 'Titan Time');
}

export interface TitanSetPiece {
	id: number;
	name: string;
}
export interface TitanSetBonus {
	threshold: number;
	description: string;
}
export interface TitanItemSetInfo {
	id: number;
	name: string;
	items: TitanSetPiece[];
	bonuses: TitanSetBonus[];
}

const titanSetsByName = new Map<string, TitanItemSetInfo>();
const titanItemIds = new Set<number>();

export function registerTitanItemSets(sets: TitanItemSetInfo[] | undefined | null) {
	titanSetsByName.clear();
	for (const s of sets || []) {
		if (s?.name) {
			titanSetsByName.set(s.name, s);
		}
	}
}

export function registerTitanItemIds(ids: number[] | undefined | null) {
	titanItemIds.clear();
	for (const id of ids || []) {
		if (id) {
			titanItemIds.add(id);
		}
	}
}

export function isTitanCustomItemId(id: number): boolean {
	return id >= TITAN_CUSTOM_ITEM_MIN_ID || titanItemIds.has(id);
}

function renderSetBlock(item: Item, equippedIds: number[]): string {
	const set = titanSetsByName.get(item.setName);
	const equipped = new Set(equippedIds);
	if (!set) {
		return `<div class="titan-tt-set">${t('套装：', 'Set: ')}${escapeHtml(item.setName)}</div>`;
	}
	const worn = set.items.reduce((n, p) => n + (equipped.has(p.id) ? 1 : 0), 0);
	const lines: string[] = [
		`<div class="titan-tt-set">${t('套装：', 'Set: ')}${escapeHtml(set.name)} (${worn}/${set.items.length})</div>`,
	];
	for (const piece of set.items) {
		const on = equipped.has(piece.id);
		lines.push(`<div class="titan-tt-set-piece${on ? ' is-active' : ''}">${escapeHtml(piece.name)}</div>`);
	}
	for (const bonus of set.bonuses || []) {
		const on = worn >= bonus.threshold;
		lines.push(
			`<div class="titan-tt-set-bonus${on ? ' is-active' : ''}">` +
				`(${bonus.threshold}) ${escapeHtml(bonus.description)}` +
			`</div>`,
		);
	}
	return lines.join('');
}

export function buildItemTooltipHtml(item: Item, equippedIds: number[] = []): string {
	const color = qualityColors[item.quality] || '#ffffff';
	const rows: string[] = [];
	const origin = isTitanCustomItemId(item.id)
		? titanOriginLabel(item.phase)
		: (item.phase ? `Phase ${item.phase}` : '');

	rows.push(
		`<div class="titan-tt-head">` +
			`<div class="titan-tt-name" style="color:${color}">${escapeHtml(item.name)}</div>` +
			(origin ? `<div class="titan-tt-badge">[${escapeHtml(origin)}]</div>` : '') +
		`</div>`,
	);
	if (item.ilvl) {
		rows.push(`<div class="titan-tt-ilvl">${t('物品等级：', 'Item Level: ')}${item.ilvl}</div>`);
	}
	if (item.unique) {
		rows.push(`<div>${t('唯一装备', 'Unique-Equipped')}</div>`);
	}
	rows.push(`<div>${t('拾取后绑定', 'Binds when picked up')}</div>`);

	const slot = slotLabel(item);
	const kind = typeLabel(item);
	if (slot || kind) {
		rows.push(
			`<div class="titan-tt-row">` +
				`<span>${escapeHtml(slot)}</span>` +
				`<span>${escapeHtml(kind)}</span>` +
			`</div>`,
		);
	}
	if (item.weaponDamageMin || item.weaponDamageMax) {
		const dps = item.weaponSpeed ? ((item.weaponDamageMin + item.weaponDamageMax) / 2) / item.weaponSpeed : 0;
		rows.push(
			`<div class="titan-tt-row">` +
				`<span>${item.weaponDamageMin.toFixed(0)} - ${item.weaponDamageMax.toFixed(0)} ${t('伤害', 'Damage')}</span>` +
				`<span>${t('速度', 'Speed')} ${item.weaponSpeed.toFixed(2)}</span>` +
			`</div>` +
			`<div>(${dps.toFixed(1)} ${t('每秒伤害', 'damage per second')})</div>`,
		);
	}
	for (const line of whiteStatLines(item.stats)) {
		rows.push(`<div>${formatWhiteStat(line)}</div>`);
	}
	for (const line of plusGreenStatLines(item.stats)) {
		const sign = line.value > 0 ? '+' : '';
		rows.push(`<div class="titan-tt-green">${sign}${Math.round(line.value)} ${escapeHtml(line.label)}</div>`);
	}
	if (item.gemSockets?.length) {
		for (const sock of item.gemSockets) {
			rows.push(
				`<div class="titan-tt-socket-row">` +
					`<span class="titan-tt-gem titan-tt-gem-${socketCss(sock)}"></span>` +
					`${escapeHtml(socketLabel(sock))}` +
				`</div>`,
			);
		}
		const bonus = socketBonusLines(item.socketBonus);
		if (bonus.length) {
			rows.push(
				`<div class="titan-tt-socket-bonus">${t('插槽加成：', 'Socket Bonus: ')}` +
				bonus.map(b => `+${Math.round(b.value)} ${escapeHtml(b.label)}`).join(', ') +
				`</div>`,
			);
		}
	}
	if (item.classAllowlist?.length) {
		rows.push(`<div>${t('职业：', 'Classes: ')}${item.classAllowlist.map(classLabel).join(t('，', ', '))}</div>`);
	}
	for (const line of equipStatLines(item.stats)) {
		rows.push(`<div class="titan-tt-green">${formatEquipLine(line)}</div>`);
	}
	if (item.setName) {
		rows.push(renderSetBlock(item, equippedIds));
	}
	return rows.join('');
}

export function attachLocalItemTooltip(elem: HTMLElement, item: Item, equippedIds: number[] = []) {
	if (!isTitanCustomItemId(item.id)) {
		return;
	}
	clearLocalItemTooltip(elem);
	new Tooltip(elem, {
		html: true,
		sanitize: false,
		title: () => buildItemTooltipHtml(item, equippedIds),
		customClass: 'titan-item-tooltip',
		placement: 'right',
		trigger: 'hover',
		container: 'body',
		delay: { show: 80, hide: 40 },
	});
}

export function clearLocalItemTooltip(elem: HTMLElement) {
	Tooltip.getInstance(elem)?.dispose();
	elem.removeAttribute('title');
	elem.removeAttribute('data-bs-original-title');
}
