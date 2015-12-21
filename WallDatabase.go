package main

// following along somewhat with https://golang.org/doc/articles/wiki/

import (
  "fmt"
  "io/ioutil"
  "net/http"
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
)

func main() {
  
}

/*
Basic flow:
  User starts program
  Program hosts web server on :8080
  Serve HTML page displaying current wallpaper image and checkbox list of tags
    also a text form for new tags
    For wallpapers which are already tagged, some checkboxes should default to true
    Bonus points for not using javascript
  User selects tags
  Tags are POSTed back to web server listener
  Filename, tags are added to database
  Web server starts serving next wallpaper image HTML
  Web page auto-refreshes
*/
