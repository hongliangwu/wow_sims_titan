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
	368: '附魔披风 - 强效敏捷',
	369: '护腕 - 主要智力',
	684: '手套 - 主要力量',
	849: '附魔披风 - 次级敏捷',
	851: '附魔靴子 - 精神',
	983: '超级敏捷',
	1071: '主要耐力',
	1075: '强效耐力',
	1099: '主要敏捷',
	1103: '卓越敏捷',
	1119: '卓越智力',
	1128: '强效智力',
	1144: '胸部 - 主要精神',
	1147: '强效精神',
	1262: '超级奥术抗性',
	1354: '超级火焰抗性',
	1400: '超级自然抗性',
	1441: '附魔披风 - 强效暗影抗性',
	1446: '超级暗影抗性',
	1593: '护腕 - 突袭',
	1597: '强效突袭',
	1600: '打击',
	1603: '粉碎',
	1606: '强效效能',
	1891: '护腕 - 属性',
	1950: '胸部 - 防御',
	1951: '泰坦织纹',
	1952: '防御',
	1953: '强效防御',
	2326: '强效法术能量',
	2332: '超级法术能量',
	2381: '强效法力回复',
	2564: '敏捷',
	2583: '力量降临',
	2605: '赞达拉魔印',
	2622: '附魔披风 - 躲闪',
	2647: '护腕 - 蛮力',
	2648: '附魔披风 - 钢织',
	2649: '附魔靴子 - 耐力',
	2650: '护腕 - 法术能量',
	2654: '智力',
	2656: '附魔靴子 - 活力',
	2657: '附魔靴子 - 敏捷',
	2658: '附魔靴子 - 稳固',
	2659: '胸部 - 卓越生命',
	2661: '胸部 - 卓越属性',
	2666: '主要智力',
	2667: '残暴',
	2669: '主要法术能量',
	2670: '双手武器 - 主要敏捷',
	2716: '天灾之韧',
	2717: '天灾之威',
	2721: '天灾之力',
	2747: '神秘魔线',
	2748: '符文魔线',
	2928: '法术能量',
	2931: '属性',
	2933: '胸部 - 主要韧性',
	2935: '手套 - 法术打击',
	2937: '手套 - 主要法术能量',
	2938: '附魔披风 - 法术穿透',
	2939: '附魔靴子 - 猫之迅捷',
	2940: '附魔靴子 - 野猪之速',
	2978: '高级守卫铭文',
	2982: '高级纪律铭文',
	2986: '高级复仇铭文',
	2988: '自然护甲片',
	2991: '高级骑士铭文',
	2995: '高级宝珠铭文',
	2997: '高级利刃铭文',
	2998: '耐久铭文',
	2999: '防御者秘药',
	3002: '力量秘药',
	3003: '凶猛秘药',
	3004: '角斗士秘药',
	3010: '蛇皮腿甲片',
	3012: '虚空蛇皮腿甲片',
	3013: '虚空裂片腿甲片',
	3096: '流浪者秘药',
	3150: '胸部 - 法力回复',
	3222: '强效敏捷',
	3229: '韧性',
	3230: '超级冰霜抗性',
	3231: '精准',
	3232: '海象人的活力',
	3233: '卓越法力',
	3234: '精准',
	3236: '强效生命',
	3243: '法术穿透',
	3244: '强效活力',
	3245: '卓越韧性',
	3246: '卓越法术能量',
	3252: '超级属性',
	3253: '武器大师',
	3256: '暗影护甲',
	3294: '强效护甲',
	3296: '智慧',
	3297: '超级生命',
	3325: '尤芒腿甲片',
	3326: '奈幽腿甲片',
	3327: '尤芒腿甲加固',
	3328: '奈幽腿甲加固',
	3329: '北地护甲片',
	3330: '厚北地护甲片',
	3605: '柔性衬垫',
	3606: '氮气推进器',
	3718: '闪光魔线',
	3719: '璀璨魔线',
	3720: '蔚蓝魔线',
	3721: '蓝宝石魔线',
	3731: '钛制武器链',
	3756: '毛皮衬垫 - 攻击强度',
	3757: '毛皮衬垫 - 耐力',
	3758: '毛皮衬垫 - 法术能量',
	3759: '毛皮衬垫 - 火焰抗性',
	3760: '毛皮衬垫 - 冰霜抗性',
	3761: '毛皮衬垫 - 暗影抗性',
	3762: '毛皮衬垫 - 自然抗性',
	3763: '毛皮衬垫 - 奥术抗性',
	3788: '精准',
	3791: '耐力',
	3793: '凯旋铭文',
	3794: '支配铭文',
	3795: '凯旋秘药',
	3796: '支配秘药',
	3806: '次级风暴铭文',
	3807: '次级峭壁铭文',
	3808: '高级利斧铭文',
	3809: '高级峭壁铭文',
	3810: '高级风暴铭文',
	3811: '高级巅峰铭文',
	3812: '冰霜之魂秘药',
	3813: '剧毒防护秘药',
	3814: '遁影秘药',
	3815: '蚀月秘药',
	3816: '烈焰之魂秘药',
	3817: '折磨秘药',
	3818: '坚毅守护者秘药',
	3819: '祥和治愈秘药',
	3820: '燃烧之谜秘药',
	3822: '霜皮腿甲片',
	3823: '冰鳞腿甲片',
	3824: '突袭',
	3825: '速度',
	3826: '冰行者',
	3827: '大屠杀',
	3828: '强效残暴',
	3829: '强效突袭',
	3830: '卓越法术能量',
	3831: '强效速度',
	3832: '强力属性',
	3833: '超级效能',
	3834: '强效法术能量',
	3835: '大师利斧铭文',
	3836: '大师峭壁铭文',
	3837: '大师巅峰铭文',
	3838: '大师风暴铭文',
	3839: '突袭',
	3840: '强效法术能量',
	3842: '野蛮角斗士秘药',
	3844: '卓越精神',
	3845: '强效突袭',
	3849: '钛合金镀层',
	3850: '主要耐力',
	3852: '高级角斗士铭文',
	3853: '大地腿甲片',
	3854: '法杖 - 强效法术能量',
	3855: '法杖 - 法术能量',
	3859: '弹性蛛网织层',
	3860: '网状护甲织层',
	3872: '神圣魔线',
	3873: '大师魔线',
	3875: '次级利斧铭文',
	3876: '次级巅峰铭文',
	3878: '思维放大碟',
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
		const cnName = enchantNameCn[enchant.effectId];
		if (cnName) return cnName;
		const statsDesc = formatEnchantStatsCn(enchant.stats);
		if (statsDesc) return statsDesc;
	}
	const descriptionsMap = await fetchEnchantDescriptions();
	return descriptionsMap[enchant.effectId] || enchant.name;
}

export function getEnchantDescriptionSync(enchant: Enchant): string {
	if (isCn()) {
		const cnName = enchantNameCn[enchant.effectId];
		if (cnName) return cnName;
		const statsDesc = formatEnchantStatsCn(enchant.stats);
		if (statsDesc) return statsDesc;
	}
	return enchant.name;
}

// Returns a string uniquely identifying the enchant.
export function getUniqueEnchantString(enchant: Enchant): string {
	return enchant.effectId + '-' + enchant.type;
}
