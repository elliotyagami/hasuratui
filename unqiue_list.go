package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
)

func unqiue(file string) {
	// get all files in directory
	// check error
	line, _ := ioutil.ReadFile(file)
	// turn the byte slice into string format
	strLine := string(line)
	// split the lines by a space, can also change this
	lines := strings.Split(strLine, " ")
	// remove the duplicates from lines slice (from func we created)
	// fmt.Println(lines)

	// get the actual file
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
	// err check
	if err != nil {
		log.Println(err)
	}
	var uniques []string
	// delete old one
	os.Remove(file)
	// create it again
	f, err = os.Create(file)
	// go through your lines
	for _, v := range lines {
		skip := false
		for _, u := range uniques {
			if v == u {
				skip = true
				break
			}
		}
		if !skip {
			uniques = append(uniques, v)
		}
	}
	// write to the file without the duplicates
	for e := range uniques {
		// write to the file without the duplicates
		f.Write([]byte(lines[e] + " ")) // added a space here, but you can change this
	}
	// close file
	defer f.Close()
}
