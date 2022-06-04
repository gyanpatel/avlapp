package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/sessions"
)

var (
	store           = sessions.NewCookieStore([]byte("asdaskdhasdhgsajdgasdsadksakdhasidoajsdousahdopj"))
	authenticatedYN = "authenticated-bcc"
	sessoinKeyID    = "user-authenticated-bcc"
	sessionUser     = "username-acc"
)

func handleLogin(w http.ResponseWriter, r *http.Request) {
	e := ErrorMessages{LoginError: ""}
	t := template.Must(template.ParseFS(templates, "templates/login.html"))
	err := t.Execute(w, e)
	if err != nil {
		log.Println("ERROR:handleLogin Error occured Parsing - login ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
func handleLoginVal(w http.ResponseWriter, r *http.Request) {
	errp := r.ParseForm()
	if errp != nil {
		log.Println("ERROR:handleLoginVal Error occured r.ParseForm() ", errp)

	}
	params := r.PostForm
	accessCode := params.Get("AccessCode")
	var role string
	if accessCode == "23646" {
		role = "A"
	}
	if strings.Compare(accessCode, "1880") != 0 && strings.Compare(accessCode, "23646") != 0 {
		log.Println("ERROR : handleLoginVal", accessCode, "- login attempt failed ")
		e := ErrorMessages{LoginError: "Invalid login details"}
		t := template.Must(template.ParseFS(templates, "templates/login.html"))
		errt := t.Execute(w, e)
		if errt != nil {
			log.Println("ERROR:handleLoginVal Error occured Parsing - templates/login ", errt)
		}
		return
	} else {
		log.Println("Info : handleLoginVal", accessCode, "- login attempt successful ")
		log.Println("Info : handleLoginVal", accessCode, "Session timeout mins ", 60)
		session, _ := store.Get(r, sessoinKeyID)
		session.Options = &sessions.Options{
			Path:     "/",
			MaxAge:   60 * 720,
			HttpOnly: true,
		}
		session.Values[authenticatedYN] = true
		session.Values[sessionUser] = accessCode
		session.Values["ROLE"] = role
		errs := session.Save(r, w)
		if errs != nil {
			log.Println("ERROR:handleLoginVal Error occured session.Save ", errs)
		}
		http.Redirect(w, r, "/home", http.StatusSeeOther)
	}

}
func userSessionValidation(w http.ResponseWriter, r *http.Request) (string, string, error) {
	session, err := store.Get(r, sessoinKeyID)
	userName, _ := session.Values[sessionUser].(string)
	role, _ := session.Values["ROLE"].(string)

	log.Println("Info: userSessionValidation ", userName, role)

	if err != nil {
		log.Println("ERROR: userSessionValidation ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return "", role, err
	}
	if auth, ok := session.Values[authenticatedYN].(bool); !ok || !auth || len(userName) == 0 {
		e := ErrorMessages{LoginError: "Your sesssion has expired or you haven't logged in, please login ."}
		t := template.Must(template.ParseFS(templates, "templates/login.html"))
		errt := t.Execute(w, e)
		if errt != nil {
			log.Println("ERROR:userSessionValidation Error occured Parsing - login.html ", errt)
		}
		return "", role, fmt.Errorf(userName, " Your sesssion has expired, please login again")
	}
	return userName, role, nil
}

func handleHome(w http.ResponseWriter, r *http.Request) {
	_, role, _ := userSessionValidation(w, r)
	var playerAvail []PlayerAvail
	var playerAvailList PlayerAvailList
	rows, err := dbCon.Query(`select name,contact from player_master order by player_master.name`)
	for rows.Next() {
		var playerAvailRec PlayerAvail
		err = rows.Scan(&playerAvailRec.Name, &playerAvailRec.Contact)
		if err != nil {
			log.Println("scan error ", err)
		}
		playerAvail = append(playerAvail, playerAvailRec)
	}
	playerAvailList = PlayerAvailList{PlayerAvailList: playerAvail, Role: role}
	t := template.Must(template.ParseFS(templates, "templates/home.html", "templates/_menu.html", "templates/_footer.html"))
	err = t.Execute(w, playerAvailList)
	if err != nil {
		log.Println(err)
	}
}

func handleViewXi(w http.ResponseWriter, r *http.Request) {
	_, role, _ := userSessionValidation(w, r)
	var matchSch []MatchSch
	var matchSchList MatchSchList
	rows, err := dbCon.Query(`select playing_xi.team,playing_xi.date,playing_xi.xi,match_sch.teamc,match_sch.teamvc from  playing_xi left outer join match_sch on playing_xi.date = match_sch.date and playing_xi.team = match_sch.team   order by playing_xi.date desc`)
	for rows.Next() {
		var matchSchRec MatchSch
		var teamc, teamvc sql.NullString
		err = rows.Scan(&matchSchRec.Team, &matchSchRec.Date, &matchSchRec.Comment, &teamc, &teamvc)
		if err != nil {
			log.Println("scan error ", err)
		}
		var sel []string
		err = json.Unmarshal([]byte(matchSchRec.Comment), &sel)
		if err != nil {
			log.Println("json.Unmarshal ", err)
		}
		matchSchRec.TeamSel = append(matchSchRec.TeamSel, "")
		matchSchRec.TeamSel = append(matchSchRec.TeamSel, sel...)
		matchSchRec.TeamC = teamc.String
		matchSchRec.TeamVC = teamvc.String
		matchSch = append(matchSch, matchSchRec)
	}
	matchSchList = MatchSchList{MatchSchList: matchSch, Role: role}
	t := template.Must(template.ParseFS(templates, "templates/viewteam.html", "templates/_menu.html", "templates/_footer.html"))
	err = t.Execute(w, matchSchList)
	if err != nil {
		log.Println(err)
	}
}

func handleViewPlayers(w http.ResponseWriter, r *http.Request) {
	_, role, _ := userSessionValidation(w, r)
	var playerAvail []PlayerAvail
	var playerAvailList PlayerAvailList
	rows, err := dbCon.Query(`select name,contact from player_master order by player_master.name`)
	for rows.Next() {
		var playerAvailRec PlayerAvail
		err = rows.Scan(&playerAvailRec.Name, &playerAvailRec.Contact)
		if err != nil {
			log.Println("scan error ", err)
		}
		playerAvail = append(playerAvail, playerAvailRec)
	}
	//log.Println(playerAvail)
	playerAvailList = PlayerAvailList{PlayerAvailList: playerAvail, Role: role}
	t := template.Must(template.ParseFS(templates, "templates/viewplayers.html", "templates/_menu.html", "templates/_footer.html"))
	err = t.Execute(w, playerAvailList)
	if err != nil {
		log.Println(err)
	}
}

func handleViewSch(w http.ResponseWriter, r *http.Request) {
	_, role, _ := userSessionValidation(w, r)
	var matchSch []MatchSch
	var matchSchList MatchSchList
	nextweek := r.URL.Query().Get("date")
	var dateFilter int = 365
	if strings.Compare(nextweek, "nextweek") == 0 {
		dateFilter = 7
	}
	log.Println(dateFilter, nextweek)
	sqlQ := fmt.Sprintf(`select team,date,venue_typ,vs,postcode,teamc,teamvc,comments
	                 	   from match_sch 
	                      where date(match_sch.date) between date('now') and  date('now','+%d days');`, dateFilter)
	rows, err := dbCon.Query(sqlQ)
	for rows.Next() {
		var matchSchRec MatchSch
		err = rows.Scan(&matchSchRec.Team, &matchSchRec.Date, &matchSchRec.VenueType, &matchSchRec.Opposition, &matchSchRec.Postcode, &matchSchRec.TeamC, &matchSchRec.TeamVC, &matchSchRec.Comment)
		if err != nil {
			log.Println("scan error ", err)
		}
		matchSch = append(matchSch, matchSchRec)
	}
	matchSchList = MatchSchList{MatchSchList: matchSch, Role: role}
	t := template.Must(template.ParseFS(templates, "templates/viewsch.html", "templates/_menu.html", "templates/_footer.html"))
	err = t.Execute(w, matchSchList)
	if err != nil {
		log.Println(err)
	}
}

func handleViewAvail(w http.ResponseWriter, r *http.Request) {
	_, role, _ := userSessionValidation(w, r)
	var playerAvail []PlayerAvail
	var playerAvailList PlayerAvailList
	nextweek := r.URL.Query().Get("date")
	var dateFilter int = 365
	if strings.Compare(nextweek, "nextweek") == 0 {
		dateFilter = 7
	}
	log.Println(dateFilter, nextweek)
	sqlQ := fmt.Sprintf(`select player_avl.name,player_avl.comments,player_avl.avl_yn,player_avl.date,player_master.contact 
	from player_avl inner join player_master  on player_avl.name = player_master.name
	where date(player_avl.date) between date('now') and  date('now','+%d days') order by player_avl.date,player_avl.name;`, dateFilter)
	rows, err := dbCon.Query(sqlQ)
	for rows.Next() {
		var playerAvailRec PlayerAvail
		err = rows.Scan(&playerAvailRec.Name, &playerAvailRec.Comment, &playerAvailRec.AvlYn, &playerAvailRec.Date, &playerAvailRec.Contact)
		if err != nil {
			log.Println("scan error ", err)
		}
		day, _ := time.Parse("2006-01-02", playerAvailRec.Date)
		playerAvailRec.Day = day.Weekday().String()
		playerAvail = append(playerAvail, playerAvailRec)
	}
	playerAvailList = PlayerAvailList{PlayerAvailList: playerAvail, Role: role}
	t := template.Must(template.ParseFS(templates, "templates/view.html", "templates/_menu.html", "templates/_footer.html"))
	err = t.Execute(w, playerAvailList)
	if err != nil {
		log.Println(err)
	}
}
func handleManage(w http.ResponseWriter, r *http.Request) {
	_, role, _ := userSessionValidation(w, r)
	t := template.Must(template.ParseFS(templates, "templates/manage.html", "templates/_menu.html", "templates/_footer.html"))
	err := t.Execute(w, PlayerAvailList{PlayerAvailList: nil, Role: role})
	if err != nil {
		log.Println(err)
	}

}
func handleManageSch(w http.ResponseWriter, r *http.Request) {
	_, role, _ := userSessionValidation(w, r)
	var playerAvail []PlayerAvail
	var playerAvailList PlayerAvailList
	rows, err := dbCon.Query(`select name,contact from player_master order by player_master.name`)
	for rows.Next() {
		var playerAvailRec PlayerAvail
		err = rows.Scan(&playerAvailRec.Name, &playerAvailRec.Contact)
		if err != nil {
			log.Println("scan error ", err)
		}
		playerAvail = append(playerAvail, playerAvailRec)
	}
	playerAvailList = PlayerAvailList{PlayerAvailList: playerAvail, Role: role}
	t := template.Must(template.ParseFS(templates, "templates/managesch.html", "templates/_menu.html", "templates/_footer.html"))
	err = t.Execute(w, playerAvailList)
	if err != nil {
		log.Println(err)
	}

}

func handleSaveMatch(w http.ResponseWriter, r *http.Request) {
	userSessionValidation(w, r)
	//userName, selectedService, err := _, role, _ := userSessionValidation(w, r)
	//log.Println("handleSaveMatch", "method:", r.Method)
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Println("handleSaveMatch", "ERROR: Error occured r.ParseForm() ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		team := r.PostFormValue("team")
		matchdate := r.PostFormValue("matchdate")
		venuetype := r.PostFormValue("venuetype")
		Postcode := r.PostFormValue("Postcode")
		Opposition := r.PostFormValue("Opposition")
		TeamCaptain := r.PostFormValue("TeamCaptain")
		TeamVC := r.PostFormValue("TeamVC")
		comments := r.PostFormValue("comments")
		//log.Println(team, matchdate, venuetype, Postcode, Opposition, TeamCaptain, TeamVC, comments)
		tx, err := dbCon.Begin()
		if err != nil {
			log.Println("handleSaveMatch", "db con failure ", err)
		}
		stmt, err := tx.Prepare(`replace into match_sch ( team ,
								 date,venue_typ,postcode,vs,teamc,teamvc,comments ) values (?,?,?,?,?,?,?,?);
									);
								 `)
		if err != nil {
			log.Println("handleSaveMatch", "stmt prepare failure ", err)
		}
		defer stmt.Close()
		rs, err := stmt.Exec(team, matchdate, venuetype, Postcode, Opposition, TeamCaptain, TeamVC, comments)
		if err != nil {
			log.Println("handleSaveMatch", "stmt exec failure ", err)
		}
		rc, err := rs.RowsAffected()
		log.Println(rc, err)
		if err != nil {
			log.Println("handleSaveMatch", "stmt tx RowsAffected ", err)
		}
		err = tx.Commit()
		if err != nil {
			log.Println("handleSaveMatch", "stmt tx commit ", err)
		}
		// set header to 'application/json'
		w.Header().Set("Content-Type", "application/json")
		//setSecureHeader(w)
		// write the response
		err = json.NewEncoder(w).Encode(fmt.Sprintf("{\"userupd\" :  %d }", rc))
		if err != nil {
			log.Println("handleManagePlayer", "json.NewEncoder(w)", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func handleSaveXi(w http.ResponseWriter, r *http.Request) {
	userSessionValidation(w, r)
	//userName, selectedService, err := _, role, _ := userSessionValidation(w, r)
	//log.Println("handleSaveAvail", "method:", r.Method)
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Println("handleSaveAvail", "ERROR: Error occured r.ParseForm() ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		team := r.PostFormValue("team")
		teamdate := r.PostFormValue("teamdate")
		check11 := r.PostFormValue("check11")

		tx, err := dbCon.Begin()
		if err != nil {
			log.Println("handleSaveAvail", "db con failure ", err)
		}
		stmt, err := tx.Prepare(`replace into playing_xi ( team ,
								 date,xi ) values (?,?,?);
									);
								 `)
		if err != nil {
			log.Println("handleSaveAvail", "stmt prepare failure ", err)
		}
		defer stmt.Close()
		var rs sql.Result
		rs, err = stmt.Exec(team, teamdate, check11)
		if err != nil {
			log.Println("handleSaveAvail", "stmt exec failure ", err)
		}

		rc, err := rs.RowsAffected()
		if err != nil {
			log.Println("handleSaveAvail", "stmt tx RowsAffected ", err)
		}
		err = tx.Commit()
		if err != nil {
			log.Println("handleSaveAvail", "stmt tx commit ", err)
		}
		// set header to 'application/json'
		w.Header().Set("Content-Type", "application/json")
		//setSecureHeader(w)
		// write the response
		err = json.NewEncoder(w).Encode(fmt.Sprintf("{\"userupd\" :  %d }", rc))
		if err != nil {
			log.Println("handleManagePlayer", "json.NewEncoder(w)", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func handleSaveAvail(w http.ResponseWriter, r *http.Request) {
	userSessionValidation(w, r)
	//userName, selectedService, err := _, role, _ := userSessionValidation(w, r)
	//log.Println("handleSaveAvail", "method:", r.Method)
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Println("handleSaveAvail", "ERROR: Error occured r.ParseForm() ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		playername := r.PostFormValue("playername")
		avl := r.PostFormValue("avlYN")
		avlDate := r.PostFormValue("avlDate")
		comments := r.PostFormValue("comments")
		log.Println(playername, avlDate, avl, comments)
		dates := strings.Split(avlDate, ",")
		avlYN := avl
		if avl == "true" {
			avlYN = "Yes"
		} else {
			avlYN = "No"
		}
		tx, err := dbCon.Begin()
		if err != nil {
			log.Println("handleSaveAvail", "db con failure ", err)
		}
		stmt, err := tx.Prepare(`replace into player_avl ( name ,
								 date,avl_yn,comments ) values (?,?,?,?);
									);
								 `)
		if err != nil {
			log.Println("handleSaveAvail", "stmt prepare failure ", err)
		}
		defer stmt.Close()
		var rs sql.Result
		for _, date := range dates {
			rs, err = stmt.Exec(playername, date, avlYN, comments)
			if err != nil {
				log.Println("handleSaveAvail", "stmt exec failure ", err)
			}
		}
		rc, err := rs.RowsAffected()
		log.Println(rc, err)
		if err != nil {
			log.Println("handleSaveAvail", "stmt tx RowsAffected ", err)
		}
		err = tx.Commit()
		if err != nil {
			log.Println("handleSaveAvail", "stmt tx commit ", err)
		}
		// set header to 'application/json'
		w.Header().Set("Content-Type", "application/json")
		//setSecureHeader(w)
		// write the response
		err = json.NewEncoder(w).Encode(fmt.Sprintf("{\"userupd\" :  %d }", rc))
		if err != nil {
			log.Println("handleManagePlayer", "json.NewEncoder(w)", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func handleManagePlayer(w http.ResponseWriter, r *http.Request) {
	userSessionValidation(w, r)
	//userName, selectedService, err := _, role, _ := userSessionValidation(w, r)
	log.Println("handleManagePlayer", "method:", r.Method)
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Println("handleManagePlayer", "ERROR: Error occured r.ParseForm() ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		playername := r.PostFormValue("playername")
		playercontact := r.PostFormValue("playercontact")
		tx, err := dbCon.Begin()
		if err != nil {
			log.Println("handleManagePlayer", "db con failure ", err)
		}
		stmt, err := tx.Prepare(`replace into player_master ( name ,
								 contact ) values (?,?);
									);
								 `)
		if err != nil {
			log.Println("handleManagePlayer", "stmt prepare failure ", err)
		}
		rs, err := stmt.Exec(playername, playercontact)
		if err != nil {
			log.Println("handleManagePlayer", "stmt exec failure ", err)
		}
		rc, err := rs.RowsAffected()
		log.Println(rc, err)
		if err != nil {
			log.Println("handleManagePlayer", "stmt tx RowsAffected ", err)
		}
		err = tx.Commit()
		if err != nil {
			log.Println("handleManagePlayer", "stmt tx commit ", err)
		}
		// set header to 'application/json'
		w.Header().Set("Content-Type", "application/json")
		//setSecureHeader(w)
		// write the response
		err = json.NewEncoder(w).Encode(fmt.Sprintf("{\"userupd\" :  %d }", rc))
		if err != nil {
			log.Println("handleManagePlayer", "json.NewEncoder(w)", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
func handleRemovePlayer(w http.ResponseWriter, r *http.Request) {
	userSessionValidation(w, r)
	//userName, selectedService, err := _, role, _ := userSessionValidation(w, r)
	log.Println("handleRemovePlayer", "method:", r.Method)
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			log.Println("handleRemovePlayer", "ERROR: Error occured r.ParseForm() ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		playername := r.PostFormValue("playername")
		tx, err := dbCon.Begin()
		if err != nil {
			log.Println("handleRemovePlayer", "db con failure ", err)
		}
		stmt, err := tx.Prepare(`delete from player_master where name = ? ; `)
		if err != nil {
			log.Println("handleRemovePlayer", "stmt prepare failure ", err)
		}
		rs, err := stmt.Exec(playername)
		if err != nil {
			log.Println("handleRemovePlayer", "stmt exec failure ", err)
		}
		rc, err := rs.RowsAffected()
		log.Println(rc, err, playername)
		if err != nil {
			log.Println("handleRemovePlayer", "stmt tx RowsAffected ", err)
		}
		err = tx.Commit()
		if err != nil {
			log.Println("handleManagePlayer", "stmt tx commit ", err)
		}
		//voucherDet := r.PostFormValue("voucherdet")
		// set header to 'application/json'
		w.Header().Set("Content-Type", "application/json")
		//setSecureHeader(w)
		// write the response
		err = json.NewEncoder(w).Encode(fmt.Sprintf("{\"userupd\" :  %d }", rc))
		if err != nil {
			log.Println("handleRemovePlayer", "json.NewEncoder(w)", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

type PlayerAvail struct {
	Name    string
	Date    string
	Contact string
	Comment string
	AvlYn   string
	Day     string
	Role    string
}
type MatchSch struct {
	Team       string
	Date       string
	Opposition string
	VenueType  string
	Postcode   string
	TeamC      string
	TeamVC     string
	Comment    string
	TeamSel    []string
	Role       string
}
type PlayerAvailList struct {
	PlayerAvailList []PlayerAvail
	Role            string
}

type MatchSchList struct {
	MatchSchList []MatchSch
	Role         string
}

//errorMessages to for invalid login
type ErrorMessages struct {
	LoginError         string
	SessionTimeOutMins int
}

//errorPage represents shows an error message
type ErrorPage struct {
	ErrorMsg string
}

func add(a, b int) int {
	return a + b
}
