package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/goarne/msort/core"
)

//Function builds the appconfiguration before the main part of the app.
func init() {

	flag.StringVar(&core.CmdPrm.ConfigFile, "configfile", "", "Configurationfile for the application.")

	flag.BoolVar(&core.CmdPrm.ShallArchive, "archive", false, "Flag tells the application to copy the archiving.")
	flag.BoolVar(&core.CmdPrm.Verbose, "verbose", false, "Prints debug information.")
	flag.BoolVar(&core.CmdPrm.Overwrite, "overwrite", false, "Checks if existing files shall be overwritten.")

	flag.StringVar(&core.CmdPrm.FilePattern, "pattern", "*", "Regex-pattern for filenames to be copied.")
	flag.StringVar(&core.CmdPrm.Source, "source", "./", "Namme of sourcefolder, when other than current working folder.")

	flag.StringVar(&core.CmdPrm.Target, "target", "./", "Name of target folder, when other than current working folder.")
	flag.StringVar(&core.CmdPrm.TargetPattern, "targetpattern", "yyyy/mm/dd", "Structure of targetfolder.")

	flag.Parse()
	if core.CmdPrm.ConfigFile != "" {
		core.CmdPrm.ReadConfig()
	}
}

//Executes the program in two goroutines. One fine and one archiving.
func main() {
	core.StartAsync(2)

	filesToArchive := make(chan *core.ArchiveFile)

	go core.FindFiles(filesToArchive)

	go core.ArchiveFiles(filesToArchive)

	core.WaitAsync()

	fmt.Println("Params:", core.CmdPrm)
	fmt.Println("Found ", strconv.Itoa(core.FileCount), " file(s).")
}
