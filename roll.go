package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var r = rand.New(rand.NewSource(time.Now().UnixNano()))

func AttackRoll(toHit, AC int) (bool, bool) {
	natural := r.Intn(20) + 1
	if natural == 20 {
		return true, true
	}
	return natural+toHit >= AC, false
}

func AdvantageAttackRoll(toHit, AC int) (bool, bool, bool) {
	hit1, crit1 := AttackRoll(toHit, AC)
	hit2, crit2 := AttackRoll(toHit, AC)
	return hit1 || hit2, crit1 || crit2, crit1 && crit2
}

type Die struct {
	Count int
	Sides int
	Fixed int
}

func (d Die) Roll() int {
	if d.Fixed != 0 {
		return d.Fixed
	}
	total := 0
	for i := 0; i < d.Count; i++ {
		total += r.Intn(d.Sides) + 1 // turn [0,sides) into (0,sides]
	}
	return total
}

func (d Die) Crit() int {
	critPortion := d.Count * d.Sides // Fixed is dropped.
	return d.Roll() + critPortion
}

type Dice []Die

func (ds Dice) Roll() int {
	total := 0
	for _, d := range ds {
		total += d.Roll()
	}
	return total
}
func (ds Dice) Crit() int {
	total := 0
	for _, d := range ds {
		total += d.Crit()
	}
	return total
}

func ParseDice(rstr string) (Dice, error) {
	tokens := strings.Split(rstr, "+")
	var dice Dice
	for _, dieSet := range tokens {
		dsstr := strings.TrimSpace(dieSet)
		tokens := strings.Split(dsstr, "d")

		// fixed value
		if len(tokens) == 1 {
			fstr := strings.TrimSpace(tokens[0])
			val, err := strconv.Atoi(fstr)
			if err != nil {
				return nil, fmt.Errorf("could not parse %q as int: %v", fstr, err)
			}
			dice = append(dice, Die{
				Fixed: val,
			})
			continue
		}

		count := 1 // omitting the count means 1
		if tokens[0] != "" {
			var err error
			cstr := strings.TrimSpace(tokens[0])
			count, err = strconv.Atoi(cstr)
			if err != nil {
				return nil, fmt.Errorf("could not parse %q as int: %v", cstr, err)
			}
		}
		sstr := strings.TrimSpace(tokens[1])
		sides, err := strconv.Atoi(sstr)
		if err != nil {
			return nil, fmt.Errorf("could not parse %q as int: %v", sstr, err)
		}
		dice = append(dice, Die{
			Count: count,
			Sides: sides,
		})
	}
	return dice, nil
}
