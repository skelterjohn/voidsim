package main

import (
	"fmt"
)

func MeleeAttack(attacker, defender Group, attackerBonus, defenderBonus Effect) (Group, error) {
	a := attacker.Attributes.Melee.Apply(attackerBonus)
	defender, err := DoAttack(attacker.Count(), a, defender, defenderBonus)
	if err != nil {
		return Group{}, fmt.Errorf("could not make melee attack: %v", err)
	}
	return defender, err
}

func RangedAttack(attacker, defender Group, attackerBonus, defenderBonus Effect) (Group, error) {
	a := attacker.Attributes.Ranged.Apply(attackerBonus)
	defender, err := DoAttack(attacker.Count(), a, defender, defenderBonus)
	if err != nil {
		return Group{}, fmt.Errorf("could not make ranged attack: %v", err)
	}
	return defender, err
}

func DoAttack(numAttackers int, attack Attack, defender Group, defenderBonus Effect) (Group, error) {
	defenderUnits := defender.Units.Split()
	advantage := numAttackers/defender.Count() > 1 // read: at least twice as many attackers as defenders

	toHit := attack.ToHit
	AC := defender.Attributes.AC + defenderBonus.AC

	whichDefender := 0

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
		} else {
			dice, err := ParseDice(attack.Damage)
			if err != nil {
				return Group{}, fmt.Errorf("could not parse attacker damage dice %q: %v", attack.Damage, err)
			}
			var damage int
			if crit {
				damage = dice.Crit()
			} else if hit {
				damage = dice.Roll()
			}
			defenderUnits[whichDefender].HP -= damage
		}
		if defenderUnits[whichDefender].HP <= 0 {
			defenderUnits = append(defenderUnits[:whichDefender], defenderUnits[whichDefender+1:]...)
			killed++
		} else {
			whichDefender++
		}
	}

	fmt.Printf("%s units killed: %d\n", defender.Name, killed)
	fmt.Printf("%s units total health: %d\n", defender.Name, Units(defenderUnits).Health())
	if len(defenderUnits) == 0 {
		fmt.Printf("%s units were WIPED OUT\n", defender.Name)
	}

	defender.Units = defenderUnits

	return defender, nil
}
