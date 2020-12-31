package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-yaml/yaml"
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

func AskNumber(prompt string) (int, error) {
	resp, err := AskString(prompt)
	if err != nil {
		return 0, fmt.Errorf("could not AskString: %v", err)
	}
	n, err := strconv.ParseInt(resp, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("could not interpret %q as a number: %v", resp, err)
	}
	return int(n), nil
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

	debugPrint("group1", group1)
	debugPrint("group2", group2)

	comment := fmt.Sprintf("Combat at %v between %q and %q", time.Now(), path1, path2)

	if err := group1.Write(path1, comment); err != nil {
		return fmt.Errorf("could not write group back to %s: %v", path1, err)
	}
	if err := group2.Write(path1, comment); err != nil {
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

	if len(os.Args) != 3 {
		Usage()
	}
	if err := Fight(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintf(os.Stderr, "Problems fighting: %v", err)
		os.Exit(1)
	}
}
