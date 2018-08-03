package main

import (
	"fmt"
	"html/template"
	"net/http"
)

func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("web/html/index.html")
	if err != nil {
		panic(err.Error())
	}
	err = t.Execute(w, nil)
	if err != nil {
		fmt.Printf("Error loading INDEX.HTML: %s", err)
	}

}

func commands(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("web/html/commands.html")
	if err != nil {
		panic(err.Error())
	}
	err = t.Execute(w, nil)
	if err != nil {
		fmt.Printf("Error loading COMMANDS.HTML: %s", err)
	}
}

func ServerMain() {
	fmt.Println("Starting server component...")
	http.HandleFunc("/", index)
	//http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	http.Handle("/web/css", http.StripPrefix("/web/css/", http.FileServer(http.Dir("css"))))
	http.HandleFunc("/commands", commands)
	http.ListenAndServe(":8000", nil)
}
