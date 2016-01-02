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
  "strings"
)

type Wallpaper struct {
	filename string
	md5 string
	tags []string
}

type WallDatabase struct {
	db *sql.DB
}

/*
type WallDatabase struct{
	database *sql.DB
	wallpapers Wallpapers
}
*/

type Wallpapers []Wallpaper

func OpenDB(dbfile string) WallDatabase {
	if ex, _ := exists(dbfile); !ex {
		NewDB(dbfile)
	}
	
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		panic(err)
	}
	
	return WallDatabase{ db }
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
	rows, err := w.db.Query(getStmt, tag)
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
		wp, err := w.ReadWP(filename)
		if err != nil { log.Fatal(err) }
		returnedWallpapers = append(returnedWallpapers, wp)
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
	err := w.db.QueryRow("SELECT md5 FROM Wallpaper WHERE md5 = ?", wp.md5).Scan(&found)
	switch {
		case err == sql.ErrNoRows:
			break
		case err != nil:
			log.Fatal(err)
		case found != "":
			return errors.New("Wallpaper already tracked. Use Edit().")
	}
	
	// Add file to database
	w.db.Exec("INSERT INTO Wallpaper VALUES(NULL, ?, ?)", wp.filename, wp.md5)
	var wallpaperID int
	w.db.QueryRow("SELECT ID FROM Wallpaper WHERE md5 = ?", wp.md5).Scan(&wallpaperID)
	
	// Add tags to database
	for _, tag := range(wp.tags) {
		// check if tag exists in database
		var tagID int
		err := w.db.QueryRow("SELECT ID FROM Tag WHERE tag = ?", tag).Scan(&tagID)
		if err == sql.ErrNoRows {
			// tag not found
			// add tag to Tag
			// get Tag.ID of added tag
			w.db.Exec("INSERT INTO Tag VALUES(NULL, ?)", tag)
			w.db.QueryRow("SELECT ID FROM Tag WHERE tag = ?", tag).Scan(&tagID)
			// log.Printf("Created tag '%s' with ID %d", tag, tagID)
		} else if err != nil {
			log.Fatal(err)
		}
		w.db.Exec("INSERT INTO IsTagged VALUES(?, ?)", wallpaperID, tagID)
		// log.Printf("Tagged %s as '%s'", wp.filename, tag)
	}
	return nil
}

/*
Debating whether this should take a filename string as an argument
or a Wallpaper as an argument. You may not know the original filename
if the user renamed a file - then trying to delete the nonexistant file
from the database will be timeconsuming without the original filename.
Likewise, constructing a Wallpaper just to pass to Remove() is also
timeconsuming, except in that situation, it would be best to ReadWP()
and then Remove() the returned Wallpaper.
*/
func (w WallDatabase) Remove(wp Wallpaper) error {
	// Look up the Wallpaper's ID in Wallpaper table
	var wallpaperID int
	selectStmt := `
	SELECT ID FROM Wallpaper WHERE md5 = ?
	`
	err := w.db.QueryRow(selectStmt, wp.md5).Scan(&wallpaperID)
	if err != nil {
		return errors.New("Wallpaper to be removed to does exist in database")
	}
	// Remove the wallpaper from the Wallpaper table
	w.db.Exec("DELETE FROM Wallpaper WHERE md5 = ")
	// Remove associations between the wallpaper to be removed and any tags
	deleteStmt
	// Remove tags whose only association was with the removed wallpaper
	
}

/*
Here I have the option of either deleting the old wallpaper completely
from the database and then adding the new wallpaper, or going through
both wallpapers and depending on any changes between the new and the old,
update the SQLite database. I feel like the way I would implement the
wallpaper struct diff function, it would not be any faster than deleting
the old wallpaper completely and then adding a new one. But maybe not.
I may write two functions to use both ways and then time them for
various updates.
*/
func (w WallDatabase) Update(oldWallpaper, newWallpaper Wallpaper) error {
	// Remove old wallpaper
	
	// Add new wallpaper
	
}

func NewWP(filename string, tags []string) Wallpaper {
	// check whether given filename exists
	if ex, _ := exists(filename); !ex {
		log.Fatal(fmt.Sprintf("Cannot find wallpaper: %s", filename))
	}
	
	// MD5 hash it
	// ioutil.ReadFile returns []byte, error
	filedata, err := ioutil.ReadFile(filename)
	if err != nil { panic(err) }
	
	// convert the md5 from [16]byte to string
	md5hash := fmt.Sprintf("%x", md5.Sum(filedata))
	
	return Wallpaper{ filename: filename, tags: tags, md5: md5hash}
}

func (w Wallpaper) String() string {
	return fmt.Sprintf("%s : %s", w.filename, strings.Join(w.tags, ", "))
}

func (w WallDatabase) ReadWP(filename string) (Wallpaper, error) {
	// Go into the Wallpaper table and SELECT *
	// Go into the IsTagged table and get the Tag IDs associated
	// Go into the Tag table and get the tag names
	row := w.db.QueryRow("SELECT * FROM Wallpaper WHERE filename = ?", filename)
	/*
	if err != nil {
		return Wallpaper{}, err
	}
	*/
	var tags []string
	var tagIDs []int
	var md5 string
	var wallpaperID int
	row.Scan(&wallpaperID, &filename, &md5)
	
	// Get list of tag IDs from IsTagged
	var currentID int
	rows, err := w.db.Query("SELECT tag FROM IsTagged WHERE wallpaper = ?", wallpaperID)
	if err != nil {
		return Wallpaper{}, err
	}
	for rows.Next() {
		rows.Scan(&currentID)
		tagIDs = append(tagIDs, currentID)
	}
	
	// Turn tag IDs into tag names
	for _, tagID := range(tagIDs) {
		row := w.db.QueryRow("SELECT tag FROM Tag WHERE ID = ?", tagID)
		/*
		if err != nil {
			return Wallpaper{}, err
		}
		*/
		var currentTag string
		row.Scan(&currentTag)
		tags = append(tags, currentTag)
	}
	return NewWP(filename, tags), nil
}