package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-yaml/yaml"
)

type Group struct {
	Name     string `yaml:"name"`
	Quantity int    `yaml:"quantity"`
	HPPer    int    `yaml:"hp_per"`
	Damage   struct {
		Ranged int `yaml:"ranged"`
		Melee  int `yaml:"melee"`
	}
	Combat struct {
		AC    int `yaml:"AC"`
		ToHit int `yaml:"to_hit"`
	}
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

func (g Group) Write(path, comment string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open file for group: %v", err)
	}
	data, err := ioutil.ReadAll(f)
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
	fmt.Fprintf(buf, "---\n# %s\n", comment)
	fmt.Fprintf(buf, "%s\n", gdata)
	fmt.Fprintf(buf, "%s\n", data)

	fi, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("could not stat file: %v", err)
	}

	ioutil.WriteFile(path, buf.Bytes(), fi.Mode())

	return nil
}
