package main

import (
	"io/ioutil"
	"strings"
)

type StringArray []string

func (s *StringArray) Set(t string) error {
	(*s) = append(*s, t)
	return nil
}

func (s *StringArray) String() string {
	return "[]"
}

type StringArrayFromFile struct {
	List StringArray
}

func (s *StringArrayFromFile) Set(t string) error {
	buf, err := ioutil.ReadFile(t)
	if err != nil {
		return err
	}

	s.List = append(s.List, strings.Split(string(buf), "\n")...)
	return nil
}

func (s *StringArrayFromFile) String() string {
	return "[]"
}
