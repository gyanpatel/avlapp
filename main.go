package main

import (
	"database/sql"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var dbCon *sql.DB

//go:embed templates/*
var templates embed.FS

func init() {
	_, err := os.Stat("sqllite_file.db")

	if err == os.ErrNotExist {
		file, err := os.Create("sqllite_file.db")
		defer file.Close()

		if err != nil {
			log.Fatal("start up failure", err)
		}
	}

	db, err := sql.Open("sqlite3", "./sqllite_file.db")
	dbCon = db
	tx, err := dbCon.Begin()
	if err != nil {
		log.Fatal("db con failure ", err)
	}
	stmt1, err := tx.Prepare(`create table if not exists player_master ( name text not null primary key,
							 contact text);
								);
							 `)
	if err != nil {
		log.Fatal("stmt prepare failure ", err)
	}
	defer stmt1.Close()
	_, err = stmt1.Exec()
	if err != nil {
		log.Fatal("stmt exec failure ", err)
	}
	//stmt2, err := tx.Prepare(`  drop table player_avl `)
	stmt2, err := tx.Prepare(`create table if not exists player_avl ( name text ,
	   comments text,
	   avl_yn text,
	   date text,
	   PRIMARY KEY (name, date)
	   );
	`)
	if err != nil {
		log.Fatal("stmt prepare failure ", err)
	}
	defer stmt2.Close()
	_, err = stmt2.Exec()
	if err != nil {
		log.Fatal("stmt exec failure ", err)
	}
	stmt3, err := tx.Prepare(`create table if not exists match_sch ( team text ,
		date text,
		venue_typ text,
		vs text,
		postcode text,
		teamc text,
		teamvc text,
		comments text,
		flex1 text,
		flex2 text,
		PRIMARY KEY (team, date)
		);
	 `)
	if err != nil {
		log.Fatal("stmt3 prepare failure ", err)
	}
	defer stmt3.Close()
	_, err = stmt3.Exec()
	if err != nil {
		log.Fatal("stmt3 exec failure ", err)
	}
	stmt4, err := tx.Prepare(`create table if not exists playing_xi ( team text ,
		date text,
		xi text,
		PRIMARY KEY (team, date)
		);
	 `)
	if err != nil {
		log.Fatal("stmt4 prepare failure ", err)
	}
	defer stmt4.Close()
	_, err = stmt4.Exec()
	if err != nil {
		log.Fatal("stmt4 exec failure ", err)
	}
	err = tx.Commit()
	if err != nil {
		log.Fatal("stmt tx commit ", err)
	}
}
func main() {
	//http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets/"))))
	assets, err := fs.Sub(fs.FS(templates), "templates/assets")
	if err != nil {
		log.Println("application could not load the static files ")
	}
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))

	http.HandleFunc("/", handleLogin)
	http.HandleFunc("/loginval", handleLoginVal)
	http.HandleFunc("/home", handleHome)
	http.HandleFunc("/view", handleViewAvail)
	http.HandleFunc("/viewteam", handleViewXi)
	http.HandleFunc("/save11", handleSaveXi)
	http.HandleFunc("/managesch", handleManageSch)
	http.HandleFunc("/viewsch", handleViewSch)
	http.HandleFunc("/manage", handleManage)
	http.HandleFunc("/addplayer", handleManagePlayer)
	http.HandleFunc("/saveavail", handleSaveAvail)
	http.HandleFunc("/savematch", handleSaveMatch)
	http.HandleFunc("/removeplayer", handleRemovePlayer)
	http.HandleFunc("/viewplayers", handleViewPlayers)

	http.ListenAndServe(":8085", nil)

}

//<td><a href="http://maps.google.co.uk/m?q=rg129lw">rg129lw</a> </td>
