package main

import (
  "fmt"
  "os"
  //"io"
  "io/ioutil"
  // https://golang.org/pkg/flag/
  "flag"
  // https://godoc.org/github.com/mattn/go-sqlite3
  _ "github.com/mattn/go-sqlite3"
  "errors"
  "strings"
  "database/sql"
  "crypto/md5"
  "log"
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
	dbh := useDatabase()
	dbh.Exec("DELETE FROM Wallpaper")
	switch {
		case addf.wallpaperfilename != "":
			addWallpaper(dbh)
		case *getf != "":
			getWallpaper(dbh)
	}
	
}

func useDatabase() (*sql.DB) {
	// dbfilef defaults to go-walls.db
	if ex, _ := exists(*dbfilef); !ex {
		createDatabase()
	}
	
	db, err := sql.Open("sqlite3", *dbfilef)
	if err != nil {
		panic(err)
	}
	
	return db
}

func createDatabase() {
	// I believe this creates the file on the disk
	// as well as opening it
	db, err := sql.Open("sqlite3", *dbfilef)
	if err != nil {
		panic(err)
	}
	
	// Create tables
	
	// Wallpaper table
		// ID
		// filename
		// MD5 hash
		
	wallpaperTableStmt := `
	CREATE TABLE Wallpaper
		(ID INTEGER PRIMARY KEY,
		filename TEXT,
		md5 TEXT);
	`
	
	_, err = db.Exec(wallpaperTableStmt)
	if err != nil {
		panic(err)
	}
	
	// Tag table
		// ID
		// tag name
		
	tagTableStmt := `
	CREATE TABLE Tag
		(ID INTEGER PRIMARY KEY,
			tag TEXT);
	`
	
	_, err = db.Exec(tagTableStmt)
	if err != nil {
		panic(err)
	}
	
	// IsTagged table
		// Wallpaper ID
		// Tag ID
		
	istaggedTableStmt := `
	CREATE TABLE IsTagged
		(wallpaper INTEGER,
		tag INTEGER,
		FOREIGN KEY(wallpaper) REFERENCES Wallpaper(ID),
		FOREIGN KEY(tag) REFERENCES Tag(ID));
	`
	_, err = db.Exec(istaggedTableStmt)
	if err != nil {
		panic(err)
	}
	
	db.Close()
}

// icza on stackoverflow:
// https://stackoverflow.com/questions/10510691/how-to-check-whether-a-file-or-directory-denoted-by-a-path-exists-in-golang
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {return true, nil}
	if os.IsNotExist(err) {return false, nil}
	return true, err
}

func getWallpaper(db *sql.DB) {

}

func addWallpaper(db *sql.DB) {
	// addf contains the userDefinition struct
	// check whether given filename exists
	
	if ex, _ := exists(addf.wallpaperfilename); !ex {
		// check against wallpaperdirf as well
		if ex, _ := exists(*wallpaperdirf + addf.wallpaperfilename); !ex {
			panic("Cannot find wallpaper")
		} else {
			addf.wallpaperfilename = *wallpaperdirf + addf.wallpaperfilename
		}
	}
	
	// MD5 hash it
	// ioutil.ReadFile returns []byte
	var filedata []byte
	var err error
	filedata, err = ioutil.ReadFile(addf.wallpaperfilename)
	if err != nil { panic(err) }
	
	// convert the md5 from [16]byte to string
	md5hash := fmt.Sprintf("%x", md5.Sum(filedata))
	log.Printf("Hashed the wallpaper: " + md5hash)
	
	// Check whether the file is already in the database
		/*
			In WallDatabase.pl, I do this by:
				quoting the wallpaper filename
				searching the Wallpaper table for that string
				if there are no results, file is new
			In WallDatabase.go, I want to use the MD5 hash instead.
			This will give me the option of updating the database
			in case of file renames or moves.
		*/
	var found string
	//This line querys the database, setting found to the md5 hash
	err = db.QueryRow("SELECT md5 FROM Wallpaper WHERE md5 = ?", md5hash).Scan(&found)
	switch {
		case err == sql.ErrNoRows:
			break
		case err != nil:
			log.Fatal(err)
		default:
			log.Fatal("Wallpaper already tracked. Use --edit.")
	}
	// debug
	// log.Printf("Made it past the switch statement.")
	
	// Add file to database
	db.Exec("INSERT INTO Wallpaper VALUES(NULL, ?, ?)", addf.wallpaperfilename, md5hash)
	var wallpaperID int
	db.QueryRow("SELECT ID FROM Wallpaper WHERE md5 = ?", md5hash).Scan(&wallpaperID)
	log.Printf("WallpaperID: " + string(wallpaperID))
	
	// Add tags to database
	for _, tag := range(addf.tags) {
		// check if tag exists in database
		var tagID int
		err := db.QueryRow("SELECT tag FROM Tag WHERE tag = ?", tag).Scan(&tagID)
		switch {
			case err == sql.ErrNoRows:
				// tag not found
				// add tag to Tag
				// get Tag.ID of added tag
				db.Exec("INSERT INTO Tag VALUES(NULL, ?)", tag)
				db.QueryRow("SELECT ID FROM Tag WHERE tag = ?", tag).Scan(&tagID)
			case err != nil:
				log.Fatal(err)
			default: 
				break
		}
		db.Exec("INSERT INTO IsTagged VALUES(?, ?)", wallpaperID, tagID)
		log.Printf("Tagged " + addf.wallpaperfilename + " as " + tag)
	}
}

/*
Currently thinking abotu another possible place to implement structures.
Imagine a wallpaper struct with methods to add themselves to the database.
We would just have to parse the command line info, make a struct,
have it check whether it exists in the database. The struct would be able
to add itself, check itself, update itself, etc. This would also allow
for greater extensibility in retrieving and using retrieved wallpapers.
We'd also be able to do wallpaper.SetCurrent() or something. The problem
is, do I want to load every sql row as a struct, or just when I need to 
do stuff with it?
*/