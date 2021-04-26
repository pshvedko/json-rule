package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/pshvedko/json-rule/rule"
)

func main() {
	var err error

	var d *os.File
	d, err = os.Open("skud.json")
	if err != nil {
		log.Fatal(err)
	}

	var j interface{}
	err = json.NewDecoder(d).Decode(&j)
	if err != nil {
		log.Fatal(err)
	}

	var f *os.File
	f, err = os.Open("rule.json")
	if err != nil {
		log.Fatal(err)
	}

	var r rule.Rule
	err = json.NewDecoder(f).Decode(&r)
	if err != nil {
		log.Fatal(err)
	}

	var c rule.Condition
	c, err = r.Condition()
	if err != nil {
		log.Fatal(err)
	}
	log.Print(c)

	t := time.Now()
	var v interface{}
	var n time.Duration
	for n < 1_000 {
		v, err = c.Exec(j)
		if err != nil {
			log.Fatal(err)
		}
		n++
	}
	log.Print(time.Now().Sub(t)/n, v)
}
