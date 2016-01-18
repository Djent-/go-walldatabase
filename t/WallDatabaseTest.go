package main

import (
	"fmt"
	wdb "github.com/djent-/go-walldatabase"
)

func main() {
	fmt.Println("Testing WallDatabase.go")
	
	fmt.Println("Testing NewWP()")
	testwp := wdb.NewWP("test.jpg", []string{"stars", "night", "snow", "trees"})
	fmt.Printf("TestWP: %s\n", testwp.String())
	
	fmt.Println("Testing OpenDB()")
	testdb := wdb.OpenDB("go-walls.db")
	
	fmt.Println("Testing Add()")
	err := testdb.Add(testwp)
	if err != nil {
		fmt.Println(err) 
	} else {
		fmt.Println("Add() test passed")
	}
	
	fmt.Println("Testing Get()")
	testget := testdb.Get("stars")
	for _, wp := range(testget) {
		fmt.Println(wp)
	}
	
	fmt.Println("Testing FetchAllWallpapers()")
	wallpapers, err := testdb.FetchAllWallpapers()
	if err != nil {
		fmt.Println("FetchAllWallpapers() test failed")
		panic(err)
	}
	for _, wall := range(wallpapers) {
		fmt.Println(wall.String())
	}
	
	fmt.Println("Testing ReadWP()")
	testrwp, err := testdb.ReadWP("test.jpg")
	if err != nil {
		fmt.Println("ReadWP() test failed")
		panic(err)
	}
	fmt.Printf("Test ReadWP: %s\n", testrwp.String())
	
	fmt.Println("Testing OpenDB() (2)")
	testdb2 := wdb.OpenDB("c:\\users\\patrick\\documents\\walls.db")
	
	fmt.Println("Testing FetchAllWallpapers() (2)")
	wallpapers, err = testdb2.FetchAllWallpapers()
	if err != nil {
		fmt.Println("FetchAllWallpapers() test (2) failed")
		panic(err)
	}
	if len(wallpapers) == 0 {
		fmt.Printf("FetchAllWallpapers() test(2) failed: ")
	}
	fmt.Printf("len(wallpapers) = %d\n", len(wallpapers))
	for _, wall := range(wallpapers) {
		fmt.Println(wall.String())
	}
	
	fmt.Println("Testing ReadWP() (2)")
	filename := "c:\\Users\\Patrick\\Pictures\\Wallpapers\\1305281760696.jpg"
	testwp2, err := testdb2.ReadWP(filename)
	if err != nil {
		fmt.Println("ReadWP() test failed")
		// panic(err) //we know
	} else {
		fmt.Printf("Test ReadWP() passed: %s", testwp2.String())
	}
}