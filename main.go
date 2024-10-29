package main

import (
	"errors"
	"fmt"
	"github.com/adrg/xdg"
	"github.com/mwinters0/hnjobs/cmd"
	"github.com/mwinters0/hnjobs/config"
	"github.com/mwinters0/hnjobs/db"
	"github.com/mwinters0/hnjobs/theme"
	"log"
	"os"
	"strings"
)

var reset = "\033[0m"
var fgBlue = "\033[38;5;6;48;5;0m"
var fgYellow = "\033[38;5;11;48;5;0m"

func main() {
	configPath, err := config.GetPath()
	if err != nil {
		log.Fatal(err)
	}
	// TODO cleanup: move these to their respective homes, make all of firstRun less ugly
	themePath, err := xdg.ConfigFile("hnjobs/") //no filename because theme loading code needs to determine it
	if err != nil {
		log.Fatal(err)
	}
	dbPath, err := xdg.DataFile("hnjobs/hnjobs.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	if checkFirstRun(configPath, themePath+"/theme-default.json", dbPath) {
		printPath := configPath
		if strings.ContainsAny(printPath, " \t\n") {
			printPath = "\"" + printPath + "\""
		}
		fmt.Printf("\n"+
			fgYellow+"Edit the config file to set up your scoring rules and then run me again:\n"+
			fgBlue+"%s"+reset+"\n", printPath)
		return
	}

	err = config.Reload()
	if err != nil {
		log.Fatal("Error loading the config file: " + err.Error())
	}
	err = theme.LoadTheme(config.GetConfig().Display.Theme, themePath)
	if err != nil {
		log.Fatal("Error loading the theme file: " + err.Error())
	}
	err = db.OpenDB(dbPath)
	if err != nil {
		log.Fatal("Error opening the database: " + err.Error())
	}

	cmd.Execute()
}

func checkFirstRun(configPath string, themePath string, dbPath string) bool {
	_, configError := os.Stat(configPath)
	if configError != nil && !errors.Is(configError, os.ErrNotExist) {
		log.Fatal(fmt.Sprintf("Unable to stat config file: %s", configError))
	}
	_, themeError := os.Stat(themePath)
	if themeError != nil && !errors.Is(themeError, os.ErrNotExist) {
		log.Fatal(fmt.Sprintf("unable to stat theme file: %s", themeError))
	}
	_, dbError := os.Stat(dbPath)
	if dbError != nil && !errors.Is(dbError, os.ErrNotExist) {
		log.Fatal(fmt.Sprintf("unable to stat db file: %s", dbError))
	}

	if errors.Is(configError, os.ErrNotExist) ||
		errors.Is(themeError, os.ErrNotExist) ||
		errors.Is(dbError, os.ErrNotExist) {
		fmt.Printf("First run detected!\n")
	} else {
		return false
	}

	// Config
	if errors.Is(configError, os.ErrNotExist) {
		fmt.Printf("  - Creating config file \""+fgBlue+"%s"+reset+"\" ...", configPath)
		err := os.WriteFile(configPath, config.DefaultConfigFileContents(), 0644)
		if err != nil {
			log.Fatal(fmt.Sprintf("Error creating config file: %v", err))
		}
		fmt.Printf(" Done.\n")
	}

	// Theme
	if errors.Is(themeError, os.ErrNotExist) {
		fmt.Printf("  - Creating theme \""+fgBlue+"%s"+reset+"\" ...", themePath)
		err := os.WriteFile(themePath, theme.DefaultThemeFileContents(), 0644)
		if err != nil {
			log.Fatal(fmt.Sprintf("Error creating theme file: %v", err))
		}
		fmt.Printf(" Done.\n")
	}

	// DB
	if errors.Is(dbError, os.ErrNotExist) {
		fmt.Printf("  - Creating database \""+fgBlue+"%s"+reset+"\" ...", dbPath)
		file, err := os.Create(dbPath)
		if err != nil {
			log.Fatal(fmt.Sprintf("Error creating database file: %v", err))
		}
		err = file.Chmod(0644)
		if err != nil {
			log.Fatal(fmt.Sprintf("Error setting database mode: %v", err))
		}
		err = db.NewDB(dbPath)
		if err != nil {
			log.Fatal(fmt.Sprintf("Error populating new DB: %v", err))
		}
		fmt.Printf(" Done.\n")
	}

	return true
}
