package main

import (
	"flag"
	"fmt"
	"strconv"

	. "github.com/goarne/msort/core"
)

//Function builds the appconfiguration before the main part of the app.
func init() {

	flag.StringVar(&CmdPrm.ConfigFile, "configfile", "", "Configurationfile for the application.")

	flag.BoolVar(&CmdPrm.ShallArchive, "archive", false, "Flag tells the application to copy the archiving.")
	flag.BoolVar(&CmdPrm.Verbose, "verbose", false, "Prints debug information.")
	flag.BoolVar(&CmdPrm.Overwrite, "overwrite", false, "Checks if existing files shall be overwritten.")

	flag.StringVar(&CmdPrm.FilePattern, "pattern", "*", "Regex-pattern for filenames to be copied.")
	flag.StringVar(&CmdPrm.Source, "source", "./", "Namme of sourcefolder, when other than current working folder.")

	flag.StringVar(&CmdPrm.Target, "target", "./", "Name of target folder, when other than current working folder.")
	flag.StringVar(&CmdPrm.TargetPattern, "targetpattern", "yyyy/mm/dd", "Structure of targetfolder.")

	flag.Parse()
	if CmdPrm.ConfigFile != "" {
		CmdPrm.ReadConfig()
	}
}

//Executes the program in two goroutines. One fine and one archiving.
func main() {
	StartAsync()

	filesToArchive := make(chan *ArchiveFile)

	go FindFiles(filesToArchive)

	go ArchiveFiles(filesToArchive)

	FinishAsync()

	fmt.Println("Params:", CmdPrm)
	fmt.Println("Found ", strconv.Itoa(FileCount), " file(s).")
}
