package main

import (
  "fmt"
  //"io/ioutil"
  // https://golang.org/pkg/flag/
  "flag"
  // https://godoc.org/github.com/mattn/go-sqlite3
  _ "github.com/mattn/go-sqlite3"
  "errors"
  "strings"
)

// Command flag vars
/*
help
version
add
edit
get
dbfile
createdb
wallpaperdir
test
random
*/

var helpf = flag.Bool("help", false, "Display help message")
var versionf = flag.Bool("version", false, "Display version number")
var addf userDefinition
var editf userDefinition
var getf = flag.String("get", "", "Get list of filenames corresponding to tag")
var dbfilef = flag.String("dbfile", "go-walls.db", "Path of database file")
var createdbf = flag.Bool("createdb", false, "Create database file")
var wallpaperdirf = flag.String("wallpaperdir", "", "Path to wallpapers")
var randomf = flag.String("random", "", "Returns random wallpaper with given tag")

// add and edit struct
type tagList []string

// Could use a better name
type userDefinition struct {
	wallpaperfilename string
	tags tagList
}

func (u *userDefinition) String() string {
	// this is how String() is handled in the pkg/flag example
	return fmt.Sprint(*u)
}

func (u *userDefinition) Set(value string) error {
	// Handle the case of a user trying to double up add and edit, etc
	if u.wallpaperfilename != "" {
		return errors.New("userDefinition flag already set")
	}
	
	counter := 0
	for _, elem := range strings.Split(value, " ") {
		// first element is the filename
		if counter == 0 {
			u.wallpaperfilename = elem 
		} else { // all other elements are tags
			u.tags = append(u.tags, elem)
		}
		counter++
	}
	return nil
}

func init() {
	flag.Var(&addf, "add", "filename of wallpaper followed by 0+ tags")
	flag.Var(&editf, "edit", "filename of wallpaper followed by new tags")
}

func main() {
	flag.Parse()
}
