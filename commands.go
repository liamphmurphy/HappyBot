package main

// MakeCommand assigns the response and permissions for each command
func MakeCommand(response, permission string) *CustomCommand {
	return &CustomCommand{
		CommandResponse:   response,
		CommandPermission: permission,
	}
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
