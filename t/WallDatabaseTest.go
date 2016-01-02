package main

import (
	"fmt"
	wdb "github.com/djent-/go-walldatabase"
	//"strings"
)

func main() {
	fmt.Println("Testing WallDatabase.go")
	testwp := wdb.NewWP("test.jpg", []string{"stars", "night", "snow", "trees"})
	fmt.Println(testwp)
	testdb := wdb.OpenDB("go-walls.db")
	err := testdb.Add(testwp)
	if err != nil { fmt.Println(err) }
	testget := testdb.Get("stars")
	for _, wp := range(testget) {
		fmt.Println(wp)
	}
}