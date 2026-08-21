import {
	UIEnchant as Enchant,
} from '../proto/ui.js';
import { getLanguageCode } from '../constants/lang.js';

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

function isCn(): boolean {
	return getLanguageCode() === 'cn';
}

// Chinese name overrides ONLY for enchants with zero stats (procs/special effects).
// Keyed by effectId. Enchants with stats use formatEnchantStatsCn instead.
const enchantNameCn: Record<number, string> = {
	963: '武器 - 强力打击',
	1897: '武器 - 超级打击',
	1900: '武器 - 十字军',
	2523: '比兹尼克247x128精确瞄准镜',
	2613: '手套 - 威胁',
	2621: '附魔披风 - 隐蔽',
	2671: '武器 - 阳炎',
	2672: '武器 - 魂霜',
	2673: '武器 - 猫鼬',
	2723: '硬化氪金瞄准镜',
	2724: '稳定永恒瞄准镜',
	2929: '打击',
	3225: '武器 - 执行者',
	3238: '手套 - 采集',
	3239: '武器 - 冰破',
	3241: '武器 - 生命护卫',
	3247: '武器 - 天灾杀手',
	3251: '武器 - 巨人杀手',
	3273: '武器 - 死亡冰霜',
	3365: '符文 - 剑刃碎裂',
	3366: '符文 - 亡灵杀手',
	3367: '符文 - 法术碎裂',
	3368: '符文 - 堕落十字军',
	3369: '符文 - 灰烬冰霜',
	3370: '符文 - 锐利冰霜',
	3594: '符文 - 剑刃格挡',
	3595: '符文 - 法术格挡',
	3599: '个人电磁脉冲发生器',
	3601: '弹片腰带',
	3603: '手部火箭',
	3604: '超速加速器',
	3607: '太阳瞄准镜',
	3608: '寻心者瞄准镜',
	3722: '光纹刺绣',
	3728: '暗辉刺绣',
	3730: '剑卫刺绣',
	3748: '钛合金盾刺',
	3789: '武器 - 狂暴',
	3790: '武器 - 黑魔法',
	3843: '钻石切割折射瞄准镜',
	3847: '符文 - 石像鬼',
	3870: '武器 - 嗜血',
	3883: '符文 - 蛛魔壳甲',
};

// Format enchant stats as a Chinese description string.
export function formatEnchantStatsCn(stats: number[] | undefined): string {
	if (!stats) return '';
	const parts: string[] = [];
	const get = (idx: number) => stats[idx] || 0;
	const add = (idx: number, label: string) => {
		const v = get(idx);
		if (v) parts.push(`+${v} ${label}`);
	};
	add(0, '力量');
	add(1, '敏捷');
	add(2, '耐力');
	add(3, '智力');
	add(4, '精神');
	add(5, '攻击强度');
	add(6, '远程攻击强度');
	add(7, '法术强度');
	add(8, '每5秒法力回复');
	add(9, '护甲');
	add(10, '额外护甲');
	add(11, '格挡值');
	add(12, '命中等级');
	add(13, '爆击等级');
	add(14, '急速等级');
	add(15, '护甲穿透等级');
	add(16, '精准等级');
	add(17, '防御等级');
	add(18, '躲闪等级');
	add(19, '招架等级');
	add(20, '格挡等级');
	add(21, '韧性等级');
	add(22, '法术穿透');
	add(23, '火焰抗性');
	add(24, '冰霜抗性');
	add(25, '自然抗性');
	add(26, '暗影抗性');
	add(27, '奥术抗性');
	return parts.join('，');
}

export async function getEnchantDescription(enchant: Enchant): Promise<string> {
	if (isCn()) {
		const stats = enchant.stats;
		if (!stats || !stats.some(v => v)) {
			const cnName = enchantNameCn[enchant.effectId];
			if (cnName) return cnName;
		}
		const statsDesc = formatEnchantStatsCn(enchant.stats);
		if (statsDesc) return statsDesc;
	}
	const descriptionsMap = await fetchEnchantDescriptions();
	return descriptionsMap[enchant.effectId] || enchant.name;
}

export function getEnchantDescriptionSync(enchant: Enchant): string {
	if (isCn()) {
		const stats = enchant.stats;
		if (!stats || !stats.some(v => v)) {
			const cnName = enchantNameCn[enchant.effectId];
			if (cnName) return cnName;
		}
		const statsDesc = formatEnchantStatsCn(enchant.stats);
		if (statsDesc) return statsDesc;
	}
	return enchant.name;
}

// Returns a string uniquely identifying the enchant.
export function getUniqueEnchantString(enchant: Enchant): string {
	return enchant.effectId + '-' + enchant.type;
}
