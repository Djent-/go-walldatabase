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

type Walldatabase *sql.DB

/*
type WallDatabase struct{
	database *sql.DB
	wallpapers Wallpapers
}
*/

type Wallpapers []Wallpaper

func OpenDB(dbfile string) WallDatabase {
	if ex, _ := exists(*db); !ex {
		NewDB(dbfile)
	}
	
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		panic(err)
	}
	
	return db
}

func NewDB(dbfile string) {
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

func (w WallDatabase) Get(tag string) Wallpapers {
	getStmt := `
	SELECT Wallpaper.filename
		FROM Wallpaper, IsTagged, Tag
		WHERE Wallpaper.ID = IsTagged.wallpaper
		AND Tag.ID = IsTagged.tag
		AND Tag.tag = ?
	`
	rows, err := w.Query(getStmt, tag)
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
		// TODO: remove single quotes from around filename
		returnedWallpapers = append(returnedWallpapers, filename)
	}
	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	return returnedWallpapers
}

func (w WallDatabase) Add(wp Wallpaper) error {
	// check whether given wallpaper exists on disk
	// this is done in Wallpaper.Set() but I feel I should do it again
	if ex, _ := exists(wp.filename); !ex {
		log.Fatal(fmt.Sprintf("Cannot find wallpaper: %s", wp.filename))
	}
	
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
	err = w.QueryRow("SELECT md5 FROM Wallpaper WHERE md5 = ?", wp.md5).Scan(&found)
	switch {
		case err == sql.ErrNoRows:
			break
		case err != nil:
			log.Fatal(err)
		default:
			return error.New("Wallpaper already tracked. Use Edit().")
	}
	
	// Add file to database
	w.Exec("INSERT INTO Wallpaper VALUES(NULL, ?, ?)", wallpaper.filename, md5hash)
	var wallpaperID int
	w.QueryRow("SELECT ID FROM Wallpaper WHERE md5 = ?", md5hash).Scan(&wallpaperID)
	
	// Add tags to database
	for _, tag := range(wp.tags) {
		// check if tag exists in database
		var tagID int
		err := w.QueryRow("SELECT ID FROM Tag WHERE tag = ?", tag).Scan(&tagID)
		if err == sql.ErrNoRows {
			// tag not found
			// add tag to Tag
			// get Tag.ID of added tag
			w.Exec("INSERT INTO Tag VALUES(NULL, ?)", tag)
			w.QueryRow("SELECT ID FROM Tag WHERE tag = ?", tag).Scan(&tagID)
			log.Printf("Created tag '%s' with ID %d", tag, tagID)
		} else if err != nil {
			log.Fatal(err)
		}
		w.Exec("INSERT INTO IsTagged VALUES(?, ?)", wallpaperID, tagID)
		log.Printf("Tagged %s as '%s'", wp.filename, tag)
	}
	return nil
}

func NewWP(filename string, tags []string) Wallpaper {
	var wallpaper Wallpaper
	wallpaper.Set(filename, tags)
	return wallpaper
}

func (w Wallpaper) Set(filename string, tags []string) {
	// check whether given filename exists
	if ex, _ := exists(filename); !ex {
		log.Fatal(fmt.Sprintf("Cannot find wallpaper: %s", filename))
	}
	w.filename = filename
	w.tags = tags
	
	// MD5 hash it
	// ioutil.ReadFile returns []byte
	var filedata []byte
	var err error
	filedata, err = ioutil.ReadFile(wallpaper.filename)
	if err != nil { panic(err) }
	
	// convert the md5 from [16]byte to string
	md5hash := fmt.Sprintf("%x", md5.Sum(filedata))
	w.md5 = md5hash
}

func (w WallDatabase) ReadWP(filename string) Wallpaper {
	// Go into the Wallpaper table and SELECT *
	// Go into the IsTagged table and get the Tag IDs associated
	// Go into the Tag table and get the tag names
	
}