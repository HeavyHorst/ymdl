package main

import (
	"strconv"
	"strings"
)

func durToSec(duration string) (int, error) {
	data := strings.Split(duration, ":")
	if len(data) == 3 {
		var h, m, s int
		var err error
		for k, v := range data {
			switch k {
			case 0:
				h, err = strconv.Atoi(v)
			case 1:
				m, err = strconv.Atoi(v)
			case 2:
				s, err = strconv.Atoi(v)
			}
		}
		return h*60*60 + m*60 + s, err
	} else if len(data) == 2 {
		var err error
		var m, s int
		for k, v := range data {
			switch k {
			case 0:
				m, err = strconv.Atoi(v)
			case 1:
				s, err = strconv.Atoi(v)
			}
		}
		return m*60 + s, err
	}
	return 0, errInvalidInput
}
