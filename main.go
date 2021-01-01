package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-yaml/yaml"
)

var (
	noSave = flag.Bool("nosave", false, "Do not save results of combat")
)

const usage = `Usage: voidsim GROUP_1 GROUP_2

GROUP_1 and GROUP_2 readable as yaml files from the directory where you run voidsim.
The results of the combat will be prepended to the files, leaving history intact.
To undo a round, simply remove it from the top of the file.

You will be prompted for a number of temporary modifications. The first question
will be if you want any, and if not you can hit return and do a fast round.
`

func Usage() {
	fmt.Fprintf(os.Stderr, usage)
	os.Exit(1)
}

func AskString(prompt string) (string, error) {
	fmt.Printf("%s: ", prompt)
	reader := bufio.NewReader(os.Stdin)
	resp, err := reader.ReadString('\n')
	return strings.TrimSpace(resp), err
}

func AskYesOrNo(prompt string, defaultYes bool) (bool, error) {
	if defaultYes {
		prompt += " (Y/n)"
	} else {
		prompt += " (y/N)"
	}
	resp, err := AskString(prompt)
	if err != nil {
		return false, fmt.Errorf("could not AskString: %v", err)
	}
	if len(resp) == 0 {
		return defaultYes, nil
	}
	if strings.ToLower(resp) == "yes" || strings.ToLower(resp) == "y" {
		return true, nil
	}
	if strings.ToLower(resp) == "no" || strings.ToLower(resp) == "n" {
		return false, nil
	}
	return false, fmt.Errorf("could not interpret %q as y/n", resp)
}

func AskNumber(prompt string, defaultValue int) (int, error) {
	resp, err := AskString(prompt + fmt.Sprintf(" (%d)", defaultValue))
	if err != nil {
		return 0, fmt.Errorf("could not AskString: %v", err)
	}

	if resp == "" {
		return defaultValue, nil
	}

	n, err := strconv.ParseInt(resp, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("could not interpret %q as a number: %v", resp, err)
	}
	return int(n), nil
}

func AskBonus(name string) (Effect, error) {
	var e Effect

	if bonus, err := AskYesOrNo(fmt.Sprintf("Do you want to give a bonus to %s?", name), false); err != nil {
		return e, fmt.Errorf("could not determine if we need to give a bonus to %s: %v", name, err)
	} else {
		if !bonus {
			return e, nil
		}
	}

	var err error
	e.Attack.ToHit, err = AskNumber("To hit", 0)
	if err != nil {
		return e, fmt.Errorf("could not ask about tohit bonus: %v", err)
	}
	e.Attack.Damage, err = AskString("Damage")
	if err != nil {
		return e, fmt.Errorf("could not ask about damage bonus: %v", err)
	}
	e.AC, err = AskNumber("AC", 0)
	if err != nil {
		return e, fmt.Errorf("could not ask about AC bonus: %v", err)
	}
	e.Resistant, err = AskYesOrNo("Resistance", false)
	if err != nil {
		return e, fmt.Errorf("could not ask about resistance: %v", err)
	}

	return e, nil
}

func Fight(path1, path2 string) error {
	var group1, group2 Group
	var err error
	group1, err = LoadGroup(path1)
	if err != nil {
		return fmt.Errorf("could not load group in %s: %v", path1, err)
	}
	group2, err = LoadGroup(path2)
	if err != nil {
		return fmt.Errorf("could not load group in %s: %v", path2, err)
	}

	melee, err := AskYesOrNo("Is this melee?", true)
	if err != nil {
		return fmt.Errorf("could not determine if this was melee: %v", err)
	}

	attackFoo := MeleeAttack
	if !melee {
		attackFoo = RangedAttack
	}

	effect1, err := AskBonus(group1.Name)
	if err != nil {
		return fmt.Errorf("could not ask about bonus for group 1, %s: %v", group1.Name, err)
	}
	effect2, err := AskBonus(group2.Name)
	if err != nil {
		return fmt.Errorf("could not ask about bonus for group 1, %s: %v", group2.Name, err)
	}

	fmt.Println("===Let's fight!===")

	group2Result, err := attackFoo(group1, group2, effect1, effect2)
	if err != nil {
		return fmt.Errorf("could not have 1 attack 2 for melee: %v", err)
	}
	group1Result, err := attackFoo(group2, group1, effect2, effect1)
	if err != nil {
		return fmt.Errorf("could not have 2 attack 1 for melee: %v", err)
	}
	group1 = group1Result
	group2 = group2Result

	comment := fmt.Sprintf("Combat at %v between %q and %q", time.Now(), path1, path2)

	if *noSave {
		fmt.Println("NOT SAVING COMBAT RESULTS")
		return nil

	}

	if err := group1.Write(path1, comment); err != nil {
		return fmt.Errorf("could not write group back to %s: %v", path1, err)
	}
	if err := group2.Write(path2, comment); err != nil {
		return fmt.Errorf("could not write group back to %s: %v", path2, err)

	}
	return nil
}

func debugPrint(label string, i interface{}) {
	ydata, err := yaml.Marshal(i)
	if err != nil {
		fmt.Printf("Could not debugPrint %q\n", label)
	} else {
		fmt.Printf("%s:\n%s\n", label, ydata)
	}
}

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		Usage()
	}
	if err := Fight(args[0], args[1]); err != nil {
		fmt.Fprintf(os.Stderr, "Problems fighting: %v", err)
		os.Exit(1)
	}
}
