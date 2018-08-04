package main

import (
	"fmt"
	"strings"
)

// MakeCommand assigns the response and permissions for each command
func MakeCommand(response, permission string) *CustomCommand {
	return &CustomCommand{
		CommandResponse:   response,
		CommandPermission: permission,
	}
}

func CommandOperations(chatmessage string) map[string]*CustomCommand {
	// Create a slice of the elements in a users message
	comSplit := strings.Split(chatmessage, " ")

	// Get the key and new value for sake of database
	comKey := comSplit[1]
	comNewValue := strings.Join(comSplit[2:], " ")

	fmt.Println(comSplit[2:])
	database := InitializeDB()
	if strings.Contains(chatmessage, "!editcom") {
		rows, err := database.Prepare("UPDATE commands SET CommandResponse = ? WHERE CommandName = ?")
		if err != nil {
			fmt.Println(err)
		}
		rows.Exec(comNewValue, comKey)
	}

	if strings.Contains(chatmessage, "!addcom") {
		rows, err := database.Prepare("INSERT INTO commands (CommandName, CommandResponse) VALUES(?,?)")
		if err != nil {
			fmt.Println(err)
		}
		rows.Exec(comKey, comNewValue)
	}

	if strings.Contains(chatmessage, "!setperm") {
		rows, err := database.Prepare("UPDATE commands SET CommandPermission = ? WHERE CommandName = ?")
		if err != nil {
			fmt.Println(err)
		}
		rows.Exec(comNewValue, comKey)
	}

	return LoadCommands()

}

// LoadCommands takes all commands from sqlite3 db and puts them in a map through CustomCommand struct
func LoadCommands() map[string]*CustomCommand {
	database := InitializeDB()

	rows, _ := database.Query("SELECT CommandName, CommandResponse, CommandPermission from commands")

	com := make(map[string]*CustomCommand)
	for rows.Next() {
		var CommandName, CommandResponse, CommandPermission string
		rows.Scan(&CommandName, &CommandResponse, &CommandPermission)
		com[CommandName] = MakeCommand(CommandResponse, CommandPermission)
	}
	return com
}
