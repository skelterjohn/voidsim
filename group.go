package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-yaml/yaml"
)

/*
name: Sword Thrower
units:
- count: 10
  HP: 133
attributes:
  max_HP: 133
  AC: 17
  melee:
	to_hit: 8
	damage: 7d6
  ranged:
	to_hit: 8
	damage: 7d6
*/

type Attack struct {
	ToHit  int    `yaml:"to_hit"`
	Damage string `yaml:"damage"`
}

func (a Attack) Apply(e Effect) Attack {
	ea := Attack{
		ToHit:  a.ToHit + e.Attack.ToHit,
		Damage: a.Damage,
	}
	if e.Attack.Damage != "" {
		ea.Damage += "+" + e.Attack.Damage
	}
	return ea
}

type Unit struct {
	Count int `yaml:"count"`
	HP    int `yaml:"HP"`
}

type Units []Unit

func (us Units) Split() []Unit {
	rus := []Unit{}
	for _, u := range us {
		for i := 0; i < u.Count; i++ {
			rus = append(rus, Unit{
				Count: 1,
				HP:    u.HP,
			})
		}
	}
	return rus
}

func (us Units) Health() int {
	th := 0
	for _, u := range us {
		th += u.Count * u.HP
	}
	return th
}

type Group struct {
	Name       string `yaml:"name"`
	Units      Units  `yaml:"units"`
	Attributes struct {
		MaxHP  int    `yaml:"max_HP"`
		AC     int    `yaml:"AC"`
		Ranged Attack `yaml:"ranged"`
		Melee  Attack `yaml:"melee"`
	} `yaml:"attributes"`
}

func (g Group) Count() int {
	total := 0
	for _, u := range g.Units {
		total += u.Count
	}
	return total
}

type Effect struct {
	AC        int
	Attack    Attack
	Resistant bool
	Occupied  bool
}

func LoadGroup(path string) (Group, error) {
	f, err := os.Open(path)
	if err != nil {
		return Group{}, fmt.Errorf("could not open file for group: %v", err)
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return Group{}, fmt.Errorf("could not read contents for group: %v", err)
	}
	var g Group
	if err := yaml.Unmarshal(data, &g); err != nil {
		return Group{}, fmt.Errorf("could not unmarshal document for group: %v", err)
	}
	return g, nil
}

func (g Group) Write(path string, comments []string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open file for group: %v", err)
	}
	historicalData, err := ioutil.ReadAll(f)
	if err != nil {
		return fmt.Errorf("could not read contents for group: %v", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("could not close file: %v", err)
	}
	gdata, err := yaml.Marshal(g)
	if err != nil {
		return fmt.Errorf("could not encode group: %v", err)
	}
	buf := &bytes.Buffer{}
	fmt.Fprint(buf, "---\n#")
	for _, c := range comments {
		fmt.Fprintf(buf, "# %s\n", c)
	}
	fmt.Fprintf(buf, "%s", gdata)
	fmt.Fprintf(buf, "%s", historicalData)

	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("could not stat file: %v", err)
	}

	ioutil.WriteFile(path, buf.Bytes(), fi.Mode())

	return nil
}
