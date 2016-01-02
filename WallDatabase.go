package WallDatabase

import (
  "fmt"
  "os"
  "io/ioutil"
  // https://godoc.org/github.com/mattn/go-sqlite3
  _ "github.com/mattn/go-sqlite3"
  "errors"
  "database/sql"
  "crypto/md5"
  "log"
)

type Wallpaper struct {
	ID int
	filename string
	md5 string
	tags []string
}

type WallDatabase *sql.DB

type Wallpapers []Wallpaper

/*
I don't think this is the best name for this function.
It does return a new INSTANCE of the database, but it also
creates an entirely new DATABASE if needed.
Could possibly make both a New() and an Open()
*/
func New(dbfile string) WallDatabase {
	if ex, _ := exists(*db); !ex {
		createDatabase(dbfile)
	}
	
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		panic(err)
	}
	
	return db
}

func createDatabase(dbfile string) {
	// I believe this creates the file on the disk
	// as well as opening it
	db, err := sql.Open("sqlite3", dbfile)
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

func WallDatabase getWallpapers(tag string) Wallpapers {
	// Print to stdout all wallpapers corresponding to a tag line by line
	getStmt := `
	SELECT Wallpaper.filename
		FROM Wallpaper, IsTagged, Tag
		WHERE Wallpaper.ID = IsTagged.wallpaper
		AND Tag.ID = IsTagged.tag
		AND Tag.tag = ?
	`
	rows, err := db.Query(getStmt, tag)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	var returnedWallpapers Wallpapers
	for rows.Next() {
		var filename string
		if err := rows.Scan(&filename); err != nil {
			log.Fatal(err)
		}
		/* I should remove the single quotes from the printed line,
		   but that would currently break compatibility with Wallpaper.pl
		   because of the way the regex is implemented. */
		fmt.Append(returnedWallpapers, filename)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return returnedWallpapers
}

func addWallpaper(db *sql.DB) {
	// addf contains the userDefinition struct
	// check whether given filename exists
	
	if ex, _ := exists(addf.wallpaperfilename); !ex {
		// check against wallpaperdirf as well
		if ex, _ := exists(*wallpaperdirf + addf.wallpaperfilename); !ex {
			panic(fmt.Sprintf("Cannot find wallpaper: %s%s", *wallpaperdirf, addf.wallpaperfilename))
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
	log.Printf("WallpaperID: %d", wallpaperID)
	
	// Add tags to database
	for _, tag := range(addf.tags) {
		// check if tag exists in database
		var tagID int
		err := db.QueryRow("SELECT ID FROM Tag WHERE tag = ?", tag).Scan(&tagID)
		if err == sql.ErrNoRows {
			// tag not found
			// add tag to Tag
			// get Tag.ID of added tag
			db.Exec("INSERT INTO Tag VALUES(NULL, ?)", tag)
			db.QueryRow("SELECT ID FROM Tag WHERE tag = ?", tag).Scan(&tagID)
			log.Printf("Created tag '%s' with ID %d", tag, tagID)
		} else if err != nil {
			log.Fatal(err)
		}
		db.Exec("INSERT INTO IsTagged VALUES(?, ?)", wallpaperID, tagID)
		log.Printf("Tagged %s as '%s'", addf.wallpaperfilename, tag)
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