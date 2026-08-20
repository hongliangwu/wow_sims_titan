package paladin

import (
	"time"

	"github.com/wowsims/wotlk/sim/core"
)

const (
	spellIDHolyVerdictTalent  = 1299093
	spellIDHolyVerdictBuff    = 1298723
	spellIDRighteousness      = 1299090
	spellIDPiety              = 1298725
	spellIDWrathOfLight       = 1299075
	spellIDHolyVerdictLockout = 1299086
	spellIDInvokeAura         = 1298724

	// SpellAuraOptions CumulativeAura: 正义/虔诚 each cap at 5.
	holyVerdictPietyThreshold = 5
	// SpellMisc DurationIndex 18 → SpellDuration 20000ms.
	holyVerdictBuffDuration         = time.Second * 20
	holyVerdictRighteousnessDuration = time.Second * 20
	holyVerdictPietyDuration         = time.Second * 20
	// SpellMisc DurationIndex 28 → 5000ms (5 ticks at 1s).
	holyVerdictInvokeDuration = time.Second * 5
	// SpellMisc DurationIndex 4 → 120000ms, same as Forbearance.
	holyVerdictLockoutDuration = time.Minute * 2
	// SpellCooldowns RecoveryTime 10000ms, StartRecoveryTime 1500ms.
	holyVerdictCooldown = time.Second * 10
)

func (paladin *Paladin) registerHolyVerdictSpell() {
	if !paladin.Talents.HolyVerdict {
		return
	}

	damageBonus := 0.50
	if paladin.Talents.ImprovedHolyVerdict {
		// 1299096: "使你的圣光裁决的伤害加成提高$s1%" (SPELLMOD_EFFECT1 +10).
		damageBonus += 0.10
	}
	echoPct := 0.20 // 1298728 EffectIndex1 Dummy BP=20

	paladin.RighteousnessAura = paladin.RegisterAura(core.Aura{
		Label:     "Righteousness",
		ActionID:  core.ActionID{SpellID: spellIDRighteousness},
		Duration:  holyVerdictRighteousnessDuration,
		MaxStacks: holyVerdictPietyThreshold,
	})

	pietyAura := paladin.RegisterAura(core.Aura{
		Label:     "Piety",
		ActionID:  core.ActionID{SpellID: spellIDPiety},
		Duration:  holyVerdictPietyDuration,
		MaxStacks: holyVerdictPietyThreshold,
	})

	var echoDamage float64
	wrathOfLight := paladin.RegisterSpell(core.SpellConfig{
		ActionID:    core.ActionID{SpellID: spellIDWrathOfLight},
		SpellSchool: core.SpellSchoolHoly,
		ProcMask:    core.ProcMaskEmpty,
		Flags:       core.SpellFlagNoOnCastComplete | core.SpellFlagIgnoreModifiers | core.SpellFlagMeleeMetrics,

		DamageMultiplier: 1,
		ThreatMultiplier: 1,
		CritMultiplier:   1,

		ApplyEffects: func(sim *core.Simulation, target *core.Unit, spell *core.Spell) {
			spell.CalcAndDealDamage(sim, target, echoDamage, spell.OutcomeAlwaysHit)
		},
	})

	paladin.HolyVerdictBuffAura = paladin.RegisterAura(core.Aura{
		Label:    "Holy Verdict",
		ActionID: core.ActionID{SpellID: spellIDHolyVerdictBuff},
		Duration: holyVerdictBuffDuration,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			paladin.PseudoStats.DamageDealtMultiplier *= 1 + damageBonus
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			paladin.PseudoStats.DamageDealtMultiplier /= 1 + damageBonus
		},
		OnSpellHitDealt: func(aura *core.Aura, sim *core.Simulation, spell *core.Spell, result *core.SpellResult) {
			if !result.Landed() || spell == wrathOfLight || !spell.IsMelee() {
				return
			}
			echoDamage = result.Damage * echoPct
			wrathOfLight.Cast(sim, result.Target)
		},
	})

	paladin.HolyVerdictLockoutAura = paladin.RegisterAura(core.Aura{
		Label:    "Holy Verdict Lockout",
		ActionID: core.ActionID{SpellID: spellIDHolyVerdictLockout},
		Duration: holyVerdictLockoutDuration,
	})

	var invokePA *core.PendingAction
	paladin.HolyVerdictInvokeAura = paladin.RegisterAura(core.Aura{
		Label:    "Invoke Holy Light",
		ActionID: core.ActionID{SpellID: spellIDInvokeAura},
		Duration: holyVerdictInvokeDuration,
		OnGain: func(aura *core.Aura, sim *core.Simulation) {
			invokePA = core.StartPeriodicAction(sim, core.PeriodicActionOptions{
				Period:   time.Second,
				NumTicks: holyVerdictPietyThreshold,
				Priority: core.ActionPriorityDOT,
				OnAction: func(sim *core.Simulation) {
					paladin.tickInvokeHolyLight(sim, pietyAura)
				},
			})
		},
		OnExpire: func(aura *core.Aura, sim *core.Simulation) {
			if invokePA != nil {
				invokePA.Cancel(sim)
				invokePA = nil
			}
		},
	})

	paladin.HolyVerdict = paladin.RegisterSpell(core.SpellConfig{
		ActionID: core.ActionID{SpellID: spellIDHolyVerdictTalent},
		Flags:    core.SpellFlagAPL,

		Cast: core.CastConfig{
			DefaultCast: core.Cast{
				GCD: core.GCDDefault,
			},
			CD: core.Cooldown{
				Timer:    paladin.NewTimer(),
				Duration: holyVerdictCooldown,
			},
		},

		ExtraCastCondition: func(sim *core.Simulation, target *core.Unit) bool {
			return paladin.RighteousnessAura.GetStacks() >= holyVerdictPietyThreshold &&
				!paladin.HolyVerdictLockoutAura.IsActive() &&
				!paladin.HolyVerdictInvokeAura.IsActive() &&
				!paladin.HolyVerdictBuffAura.IsActive()
		},

		ApplyEffects: func(sim *core.Simulation, _ *core.Unit, _ *core.Spell) {
			paladin.HolyVerdictInvokeAura.Activate(sim)
		},
	})
}

func (paladin *Paladin) addHolyVerdictRighteousness(sim *core.Simulation) {
	if paladin.RighteousnessAura == nil {
		return
	}
	paladin.RighteousnessAura.Activate(sim)
	if paladin.RighteousnessAura.GetStacks() < paladin.RighteousnessAura.MaxStacks {
		paladin.RighteousnessAura.AddStack(sim)
	}
}

func (paladin *Paladin) tickInvokeHolyLight(sim *core.Simulation, pietyAura *core.Aura) {
	if paladin.RighteousnessAura.GetStacks() <= 0 {
		paladin.HolyVerdictInvokeAura.Deactivate(sim)
		return
	}

	paladin.RighteousnessAura.RemoveStack(sim)
	if paladin.RighteousnessAura.GetStacks() == 0 {
		paladin.RighteousnessAura.Deactivate(sim)
	}

	pietyAura.Activate(sim)
	if pietyAura.GetStacks() < pietyAura.MaxStacks {
		pietyAura.AddStack(sim)
	}

	if pietyAura.GetStacks() >= holyVerdictPietyThreshold {
		pietyAura.Deactivate(sim)
		paladin.HolyVerdictInvokeAura.Deactivate(sim)
		paladin.HolyVerdictBuffAura.Activate(sim)
		paladin.HolyVerdictLockoutAura.Activate(sim)
	}
}
