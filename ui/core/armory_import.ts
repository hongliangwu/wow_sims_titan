import { Class, EquipmentSpec, ItemSpec, Race, Spec } from './proto/common.js';
import { Database } from './proto_utils/database.js';
import { classNames } from './proto_utils/names.js';
import type { IndividualSimUI } from './individual_sim_ui.js';
import { TypedEvent } from './typed_event.js';

export interface ArmoryImportItem {
	slot: string;
	id: number;
	enchant?: number;
	gems?: number[];
}

export interface ArmoryImportPayload {
	source?: string;
	classId?: number;
	raceId?: number;
	items: ArmoryImportItem[];
}

const CLASS_BY_ID: Record<number, Class> = {
	1: Class.ClassWarrior,
	2: Class.ClassPaladin,
	3: Class.ClassHunter,
	4: Class.ClassRogue,
	5: Class.ClassPriest,
	6: Class.ClassDeathknight,
	7: Class.ClassShaman,
	8: Class.ClassMage,
	9: Class.ClassWarlock,
	11: Class.ClassDruid,
};

const RACE_BY_ID: Record<number, Race> = {
	1: Race.RaceHuman,
	2: Race.RaceOrc,
	3: Race.RaceDwarf,
	4: Race.RaceNightElf,
	5: Race.RaceUndead,
	6: Race.RaceTauren,
	7: Race.RaceGnome,
	8: Race.RaceTroll,
	10: Race.RaceBloodElf,
	11: Race.RaceDraenei,
};

const SLOT_ORDER = [
	'HEAD',
	'NECK',
	'SHOULDER',
	'BACK',
	'CHEST',
	'WRIST',
	'HANDS',
	'WAIST',
	'LEGS',
	'FEET',
	'FINGER_1',
	'FINGER_2',
	'TRINKET_1',
	'TRINKET_2',
	'MAIN_HAND',
	'OFF_HAND',
	'RANGED',
];

// Runs on the HTTPS armory page, so it must be self-contained (no localhost script).
export const ARMORY_BOOKMARKLET = `(async()=>{
function findVue(el){if(!el)return null;if(el.__vue__)return el.__vue__;for(var i=0;i<el.children.length;i++){var v=findVue(el.children[i]);if(v)return v;}return null;}
function collect(vm,o){if(!vm||o.n>8000)return;o.n=(o.n||0)+1;if(vm.$store&&vm.$store.state)o.s=vm.$store.state;if(vm.equipment&&vm.equipment.equipped_items)o.e=vm.equipment;(vm.$children||[]).forEach(function(c){collect(c,o);});}
function pageType(){var p=location.pathname;if(p.indexOf("classictrr")>=0)return"classictrr";if(p.indexOf("classicann")>=0)return"classicann";if(p.indexOf("classic1x")>=0)return"classic1x";if(p.indexOf("/character/classic")>=0)return"classic";return"index";}
async function apiLoad(){
  var parts=location.hash.replace(/^#\\/?/,"").split("/");
  var realm=decodeURIComponent(parts[0]||""), name=decodeURIComponent(parts[1]||"");
  if(!realm||!name) throw new Error("请在英雄榜角色页使用此书签");
  var base="https://webapi.blizzard.cn/wow-armory-server/api/"+pageType();
  var idx=await(await fetch(base+"/index?realm_slug="+encodeURIComponent(realm)+"&role_name="+encodeURIComponent(name),{credentials:"include"})).json();
  if(idx.code===20000) throw new Error("请先登录战网/网易账号后再使用此书签");
  if(idx.code!==0) throw new Error(idx.message||("英雄榜错误 "+idx.code));
  var eq=await(await fetch(base+"/do?api=equipment&token="+encodeURIComponent(idx.data.token),{credentials:"include"})).json();
  if(eq.code!==0) throw new Error(eq.message||"读取装备失败");
  return {hero:idx.data.character_summary, equipment:eq.data};
}
function convert(hero,equipment){
  var by={}, slots=["HEAD","NECK","SHOULDER","BACK","CHEST","WRIST","HANDS","WAIST","LEGS","FEET","FINGER_1","FINGER_2","TRINKET_1","TRINKET_2","MAIN_HAND","OFF_HAND","RANGED"];
  (equipment.equipped_items||[]).forEach(function(it){
    var slot=it.slot&&it.slot.type||"";
    if(!slot||slot==="SHIRT"||slot==="TABARD") return;
    var ench=0;
    (it.enchantments||[]).forEach(function(e){ if(e.enchantment_slot&&e.enchantment_slot.type==="PERMANENT") ench=e.enchantment_id||0; });
    by[slot]={slot:slot,id:(it.item&&it.item.id)||0,enchant:ench,gems:(it.sockets||[]).map(function(s){return (s.item&&s.item.id)||0;})};
  });
  var items=[];
  slots.forEach(function(s){ if(by[s]&&by[s].id) items.push(by[s]); });
  var cls=hero&&hero.character_class||{}, race=hero&&hero.race||{};
  return {source:"blizzard-cn-armory",classId:cls.id||0,raceId:race.id||0,items:items};
}
var hero=null, equipment=null;
try{
  var acc={n:0};
  collect(findVue(document.body),acc);
  if(acc.e) equipment=acc.e;
  if(acc.s&&acc.s.indexHero) hero=acc.s.indexHero;
}catch(e){}
if(!equipment||!equipment.equipped_items||!hero){
  var api=await apiLoad();
  hero=hero||api.hero;
  equipment=equipment||api.equipment;
}
if(!equipment||!equipment.equipped_items) throw new Error("没有找到装备数据，请等英雄榜页面加载完成后再试");
var payload=convert(hero,equipment);
if(!payload.items.length) throw new Error("角色没有可导入的装备");
var json=JSON.stringify(payload);
try{ await navigator.clipboard.writeText(json); }catch(e){}
var paths={1:"/wotlk/warrior/",2:"/wotlk/retribution_paladin/",3:"/wotlk/hunter/",4:"/wotlk/rogue/",5:"/wotlk/shadow_priest/",6:"/wotlk/deathknight/",7:"/wotlk/enhancement_shaman/",8:"/wotlk/mage/",9:"/wotlk/warlock/",11:"/wotlk/feral_druid/"};
var url="http://localhost:3333"+(paths[payload.classId]||"/wotlk/")+"?armory="+encodeURIComponent(json);
window.open(url);
alert("已抓取装备并打开模拟器。若未自动导入，请在模拟器「导入 → 英雄榜」中粘贴。");
})().catch(function(e){alert(e&&e.message?e.message:String(e));});`;

export function armoryBookmarkletHref(): string {
	return 'javascript:' + encodeURIComponent(ARMORY_BOOKMARKLET);
}

export function parseArmoryCharacterUrl(input: string): { apiType: string; realm: string; name: string } | null {
	const trimmed = input.trim();
	if (!trimmed || trimmed.startsWith('{')) {
		return null;
	}
	if (!/wow\.blizzard\.cn\/character/i.test(trimmed) && !/#\/[^/]+\/[^/?#]+/.test(trimmed)) {
		return null;
	}

	const typeMatch = trimmed.match(/\/character\/(classictrr|classicann|classic1x|classic|val)(?:\/|#|$)/i);
	let apiType = (typeMatch?.[1] || 'classictrr').toLowerCase();
	if (apiType == 'val') {
		apiType = 'classic1x';
	}

	const hashMatch = trimmed.match(/#\/([^/?#]+)\/([^/?#]+)/);
	if (!hashMatch) {
		return null;
	}
	return {
		apiType,
		realm: decodeURIComponent(hashMatch[1]),
		name: decodeURIComponent(hashMatch[2]),
	};
}

export function isArmoryCharacterUrl(input: string): boolean {
	return parseArmoryCharacterUrl(input) != null;
}

export function parseArmoryPayload(data: string): ArmoryImportPayload {
	const parsed = JSON.parse(data) as ArmoryImportPayload;
	if (!parsed || !Array.isArray(parsed.items)) {
		throw new Error('无效的英雄榜导入数据');
	}
	return parsed;
}

export function classFromArmory(payload: ArmoryImportPayload): Class {
	if (payload.classId && CLASS_BY_ID[payload.classId]) {
		return CLASS_BY_ID[payload.classId];
	}
	return Class.ClassUnknown;
}

export function raceFromArmory(payload: ArmoryImportPayload): Race {
	if (payload.raceId && RACE_BY_ID[payload.raceId]) {
		return RACE_BY_ID[payload.raceId];
	}
	return Race.RaceUnknown;
}

export function equipmentSpecFromArmory(payload: ArmoryImportPayload): EquipmentSpec {
	const bySlot: Record<string, ArmoryImportItem> = {};
	for (const item of payload.items) {
		if (item?.slot && item.id) {
			bySlot[item.slot] = item;
		}
	}

	const equipmentSpec = EquipmentSpec.create();
	for (const slot of SLOT_ORDER) {
		const item = bySlot[slot];
		if (!item) {
			continue;
		}
		equipmentSpec.items.push(
			ItemSpec.create({
				id: item.id,
				enchant: item.enchant || 0,
				gems: (item.gems || []).map(id => id || 0),
			}),
		);
	}
	return equipmentSpec;
}

export async function applyArmoryImport<SpecType extends Spec>(
	simUI: IndividualSimUI<SpecType>,
	data: string,
): Promise<void> {
	if (isArmoryCharacterUrl(data)) {
		throw new Error('官方英雄榜必须登录才能读装备，模拟器带不上登录态，无法只靠链接导入。请把「英雄榜导入」拖到书签栏，在已登录的角色页点击该书签。');
	}
	const payload = parseArmoryPayload(data);
	const charClass = classFromArmory(payload);
	const race = raceFromArmory(payload);
	const equipmentSpec = equipmentSpecFromArmory(payload);
	if (equipmentSpec.items.length == 0) {
		throw new Error('英雄榜数据里没有装备');
	}

	const playerClass = simUI.player.getClass();
	if (charClass != Class.ClassUnknown && charClass != playerClass) {
		throw new Error(`职业不匹配：当前是 ${classNames.get(playerClass)}，英雄榜是 ${classNames.get(charClass)}`);
	}

	await Database.loadLeftoversIfNecessary(equipmentSpec);
	const gear = simUI.sim.db.lookupEquipmentSpec(equipmentSpec);

	const expectedItemIds = equipmentSpec.items.map(item => item.id);
	const foundItemIds = gear.asSpec().items.map(item => item.id);
	const missingItems = expectedItemIds.filter(expectedId => expectedId && !foundItemIds.includes(expectedId));

	const eventID = TypedEvent.nextEventID();
	TypedEvent.freezeAllAndDo(() => {
		if (race != Race.RaceUnknown) {
			simUI.player.setRace(eventID, race);
		}
		simUI.player.setGear(eventID, gear);
	});

	if (missingItems.length == 0) {
		alert('英雄榜导入成功！');
	} else {
		alert('导入成功，但以下物品 ID 不在模拟器数据库中：\n\n' + missingItems.join(', '));
	}
}

export function tryConsumeArmoryQueryParam(): string | null {
	const params = new URLSearchParams(window.location.search);
	const raw = params.get('armory');
	if (!raw) {
		return null;
	}
	params.delete('armory');
	const next = params.toString();
	const url = window.location.pathname + (next ? '?' + next : '') + window.location.hash;
	window.history.replaceState({}, '', url);
	return raw;
}
