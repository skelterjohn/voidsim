package main

import (
	"bytes"
	"fmt"
)

func MeleeAttack(pbp *bytes.Buffer, attacker, defender Group, attackerBonus, defenderBonus Effect) (Group, error) {
	if attackerBonus.Occupied {
		fmt.Fprintf(pbp, " %s units are OCCUPIED and cannot attack\n", attacker.Name)
		return defender, nil
	}
	a := attacker.Attributes.Melee.Apply(attackerBonus)
	defender, err := DoAttack(pbp, attacker, a, defender, defenderBonus)
	if err != nil {
		return Group{}, fmt.Errorf("could not make melee attack: %v", err)
	}
	return defender, err
}

func RangedAttack(pbp *bytes.Buffer, attacker, defender Group, attackerBonus, defenderBonus Effect) (Group, error) {
	if attackerBonus.Occupied {
		fmt.Fprintf(pbp, " %s units are OCCUPIED and cannot attack", attacker.Name)
		return defender, nil
	}
	a := attacker.Attributes.Ranged.Apply(attackerBonus)
	defender, err := DoAttack(pbp, attacker, a, defender, defenderBonus)
	if err != nil {
		return Group{}, fmt.Errorf("could not make ranged attack: %v", err)
	}
	return defender, err
}

func DoAttack(pbp *bytes.Buffer, attacker Group, attack Attack, defender Group, defenderBonus Effect) (Group, error) {
	numAttackers := attacker.Count()
	if numAttackers == 0 {
		fmt.Fprintf(pbp, "%s units cannot attack -- they're all dead!\n", attacker.Name)
		return defender, nil
	}
	fmt.Fprintf(pbp, "%s units attack summary:\n", attacker.Name)
	if defender.Count() == 0 {
		fmt.Fprintf(pbp, " %s units were already dead\n", defender.Name)
		return defender, nil
	}
	defenderUnits := defender.Units.Split()
	advantage := numAttackers/defender.Count() > 1 // read: at least twice as many attackers as defenders

	if advantage {
		fmt.Fprintf(pbp, " %s units have ADVANTAGE\n", attacker.Name)
	}

	toHit := attack.ToHit
	AC := defender.Attributes.AC + defenderBonus.AC

	whichDefender := 0

	totalDamage := 0
	killed := 0

	for i := 0; i < numAttackers; i++ {
		if len(defenderUnits) == 0 {
			break
		}

		if whichDefender >= len(defenderUnits) {
			whichDefender = 0
		}

		var hit, crit, superCrit bool
		if !advantage {
			hit, crit = AttackRoll(toHit, AC)
		} else {
			hit, crit, superCrit = AdvantageAttackRoll(toHit, AC)
		}

		if superCrit { // instant kill of one unit
			defenderUnits[whichDefender].HP = 0
			fmt.Fprintf(pbp, " supercrit-insta-death")
		} else {
			dice, err := ParseDice(attack.Damage)
			if err != nil {
				return Group{}, fmt.Errorf("could not parse attacker damage dice %q: %v", attack.Damage, err)
			}
			var damage int
			if crit {
				damage = dice.Crit()
				fmt.Fprintf(pbp, " crit=%d", damage)
			} else if hit {
				damage = dice.Roll()
				fmt.Fprintf(pbp, " hit=%d", damage)
			} else {
				fmt.Fprintf(pbp, " miss")
			}
			if damage != 0 && defenderBonus.Resistant {
				damage /= 2
				fmt.Fprintf(pbp, "/2=%d", damage)
			}
			defenderUnits[whichDefender].HP -= damage
			totalDamage += damage
		}
		if defenderUnits[whichDefender].HP <= 0 {
			defenderUnits = append(defenderUnits[:whichDefender], defenderUnits[whichDefender+1:]...)
			killed++
		} else {
			whichDefender++
		}
	}
	fmt.Fprintln(pbp)

	fmt.Fprintf(pbp, " %s units killed: %d\n", defender.Name, killed)
	fmt.Fprintf(pbp, " %d damage inflicted by %s units\n", totalDamage, attacker.Name)
	fmt.Fprintf(pbp, " %s status: %d remaining with %d total health\n", defender.Name, len(defenderUnits), Units(defenderUnits).Health())
	if len(defenderUnits) == 0 {
		fmt.Fprintf(pbp, " %s units were WIPED OUT\n", defender.Name)
	}

	defender.Units = defenderUnits

	return defender, nil
}
