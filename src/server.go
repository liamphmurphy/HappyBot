package main

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
)

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Loading index.html...")
	t, err := template.ParseFiles("html/index.html")
	if err != nil {
		panic(err.Error())
	}
	err = t.Execute(w, nil)
	if err != nil {
		fmt.Printf("Error loading INDEX.HTML: %s", err)
	}

}

func commands(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("html/commands.html")
	if err != nil {
		panic(err.Error())
	}
	err = t.Execute(w, nil)
	if err != nil {
		fmt.Printf("Error loading COMMANDS.HTML: %s", err)
	}
}

func addcomhandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles("html/index.html")
		t.Execute(w, nil)
	} else {
		r.ParseForm()

		t, _ := template.ParseFiles("html/index.html")
		t.Execute(w, nil)

	}
	AddCommand(r.Form)
}

func addtimedcomhandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		t, _ := template.ParseFiles("html/index.html")
		t.Execute(w, nil)
	} else {
		r.ParseForm()

		t, _ := template.ParseFiles("html/index.html")
		t.Execute(w, nil)

	}
	AddTimedCommand(r.Form)
}

func badwordhandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Method:", r.Method+"\n")
	if r.Method == "GET" {
		t, _ := template.ParseFiles("html/index.html")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
		t, _ := template.ParseFiles("html/index.html")
		t.Execute(w, nil)
	}
	AddBadWord(r.Form)
}

func addgoofhandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Method:", r.Method+"\n")
	if r.Method == "GET" {
		t, _ := template.ParseFiles("html/index.html")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
		t, _ := template.ParseFiles("html/index.html")
		t.Execute(w, nil)
	}
	AddGoofServer(r.Form)
}

func delcomhandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Method:", r.Method+"\n")
	if r.Method == "GET" {
		t, _ := template.ParseFiles("html/index.html")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
		t, _ := template.ParseFiles("html/index.html")
		t.Execute(w, nil)
	}
	DelCom(r.Form)
}

func delbadwordhandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Method:", r.Method+"\n")
	if r.Method == "GET" {
		t, _ := template.ParseFiles("html/index.html")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
		t, _ := template.ParseFiles("html/index.html")
		t.Execute(w, nil)
	}
	DelBadWord(r.Form)
}

func deltimedcomhandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Method:", r.Method+"\n")
	if r.Method == "GET" {
		t, _ := template.ParseFiles("html/index.html")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
		t, _ := template.ParseFiles("html/index.html")
		t.Execute(w, nil)
	}
	DelTimedCom(r.Form)
}

func delgoofhandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Method:", r.Method+"\n")
	if r.Method == "GET" {
		t, _ := template.ParseFiles("html/index.html")
		t.Execute(w, nil)
	} else {
		r.ParseForm()
		t, _ := template.ParseFiles("html/index.html")
		t.Execute(w, nil)
	}
	DelGoof(r.Form)
}

func AddBadWord(form url.Values) BadWord {
	db := InitializeDB()

	badWordName := strings.Join(form["bwname"], " ")
	insert, err := db.Prepare("INSERT INTO badwords (Badword) VALUES (?)")
	if err != nil {
		panic(err.Error())
	}

	insert.Exec(badWordName)
	return LoadBadWords()
}

func AddCommand(form url.Values) map[string]*CustomCommand {
	db := InitializeDB()
	commandName := strings.Join(form["cname"], " ")
	commandResponse := strings.Join(form["cresp"], " ")
	commandPermission := strings.Join(form["cperm"], " ")

	insert, err := db.Prepare("INSERT INTO commands (CommandName, CommandResponse, CommandPermission) VALUES (?,?,?)")
	if err != nil {
		panic(err.Error())
	}

	insert.Exec(commandName, commandResponse, commandPermission)
	return LoadCommands()
}

func AddTimedCommand(form url.Values) map[string]*CustomTimedCommand {
	db := InitializeDB()
	timedComMsg := strings.Join(form["tcmsg"], " ")
	timedComInterval := strings.Join(form["tcint"], " ")
	timedComName := strings.Join(form["tcname"], " ")

	insert, err := db.Prepare("INSERT INTO timedcommands (TimedResponse, Timer, TimedName) VALUES (?,?,?)")
	if err != nil {
		panic(err.Error())
	}

	insert.Exec(timedComMsg, timedComInterval, timedComName)
	return LoadTimedCommands()
}

func AddGoofServer(form url.Values) Goof {
	db := InitializeDB()
	newGoof := strings.Join(form["goof"], " ")

	insert, err := db.Prepare("INSERT INTO goofs (GoofName) VALUES (?)")
	if err != nil {
		panic(err.Error())
	}

	insert.Exec(newGoof)
	return LoadGoofs()
}

func DelCom(form url.Values) map[string]*CustomCommand {
	db := InitializeDB()
	commandName := strings.Join(form["cname"], " ")

	insert, err := db.Prepare("DELETE FROM commands WHERE CommandName = ?")
	if err != nil {
		panic(err.Error())
	}

	insert.Exec(commandName)
	return LoadCommands()
}

func DelBadWord(form url.Values) BadWord {
	db := InitializeDB()
	badWordName := strings.Join(form["bwname"], " ")

	insert, err := db.Prepare("DELETE FROM badwords WHERE Badword = ?")
	if err != nil {
		panic(err.Error())
	}

	insert.Exec(badWordName)
	return LoadBadWords()
}

func DelTimedCom(form url.Values) map[string]*CustomTimedCommand {
	db := InitializeDB()
	timedComName := strings.Join(form["tcname"], " ")

	insert, err := db.Prepare("DELETE FROM timedcommands WHERE TimedName = ?")
	if err != nil {
		panic(err.Error())
	}

	insert.Exec(timedComName)
	return LoadTimedCommands()
}

func DelGoof(form url.Values) Goof {
	db := InitializeDB()
	goofName := strings.Join(form["goof"], " ")

	insert, err := db.Prepare("DELETE FROM goofs WHERE GoofName = ?")
	if err != nil {
		panic(err.Error())
	}

	insert.Exec(goofName)
	return LoadGoofs()
}

func ServerMain() {
	fmt.Println("Starting server component...")
	http.HandleFunc("/", index)

	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	http.HandleFunc("/addcomhandler", addcomhandler)
	http.HandleFunc("/badwordhandler", badwordhandler)
	http.HandleFunc("/addtimedcomhandler", addtimedcomhandler)
	http.HandleFunc("/addgoofhandler", addgoofhandler)
	http.HandleFunc("/delcomhandler", delcomhandler)
	http.HandleFunc("/delbadwordhandler", delbadwordhandler)
	http.HandleFunc("/deltimedcomhandler", deltimedcomhandler)
	http.HandleFunc("/delgoofhandler", delgoofhandler)

	http.ListenAndServe(":8000", nil)
}
