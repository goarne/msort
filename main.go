package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/xiam/exif"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	//Contains global configuratins for the app.
	cmdPrm    = CmdLineParams{}
	fileCount = 0
	wg        = sync.WaitGroup{}

	//The regex extracts datevalues from an image file in the format yyyy:mm:dd
	dateRegex, _ = regexp.Compile("(?P<YYYY>\\d{4}):(?P<MM>\\d{2}):(?P<DD>\\d{2})")
	//The regex parses the stringvalue yyyy, mm and dd in the format yyyy/mm/dd.  The search is case insensitive. Each field is optional.
	targetPatternRegex, _ = regexp.Compile("(?P<YYYY>((?i)YYYY)?)/?(?P<MM>((?i)MM)?)/?(?P<DD>((?i)DD)?)")
)

//Function builds the appconfiguration before the main part of the app.
func init() {
	flag.StringVar(&cmdPrm.ConfigFile, "configfile", "", "Configurationfile for the application.")

	flag.BoolVar(&cmdPrm.ShallArchive, "archive", false, "Flag tells the application to copy the archiving.")
	flag.BoolVar(&cmdPrm.Verbose, "verbose", false, "Prints debug information.")
	flag.BoolVar(&cmdPrm.Overwrite, "overwrite", false, "Checks if existing files shall be overwritten.")

	flag.StringVar(&cmdPrm.FilePattern, "pattern", "*", "Regex-pattern for filenames to be copied.")
	flag.StringVar(&cmdPrm.Source, "source", "./", "Namme of sourcefolder, when other than current working folder.")

	flag.StringVar(&cmdPrm.Target, "target", "./", "Name of target folder, when other than current working folder.")
	flag.StringVar(&cmdPrm.TargetPattern, "targetpattern", "yyyy/mm/dd", "Structure of targetfolder.")

	flag.Parse()
	if cmdPrm.ConfigFile != "" {
		cmdPrm.ReadConfig()
	}
}

//Executes the program in two goroutines. One fine and one archiving.
func main() {
	wg = sync.WaitGroup{}
	wg.Add(1)

	filesToArchive := make(chan *ArchiveFile)

	go findFiles(filesToArchive)

	go archiveFiles(filesToArchive)

	wg.Wait()

	fmt.Println("Params:", cmdPrm)
	fmt.Println("Found ", strconv.Itoa(fileCount), " file(s).")
}

//The function sorts a list based on modified date and returns the sorted array.
func findFiles(fileSink chan *ArchiveFile) {
	defer close(fileSink)
	dir := filepath.Dir(cmdPrm.Source)

	filepath.Walk(dir, func(sourcePath string, f os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err.Error())
			return nil
		}
		if f.IsDir() == true {
			return nil
		}

		if f.Name() == cmdPrm.ConfigFile {
			return nil //Dont want to archive configfile.
		}

		if fileMatch, err := filepath.Match(cmdPrm.FilePattern, f.Name()); fileMatch == false {
			return err
		}

		foundFile := &ArchiveFile{f, sourcePath, ""}
		debug("Found file: ", f.Name())

		fileSink <- foundFile
		fileCount++
		return nil
	})
}

//The function copy files received in the channel from source to target.
//Targetpath is calculated from cameradate.
func archiveFiles(fileSource chan *ArchiveFile) {
	defer wg.Done()

	for file := range fileSource {

		file.parseTargetPath()

		if cmdPrm.ShallArchive == false { //Want to check this late because of debug messages in parseTargetPath
			continue
		}

		if file.fileExists() {
			if cmdPrm.Overwrite == false { //Dont want to copy unnescessary
				continue
			}
			debug("File ", file.targetPath, file.Name(), "exists.")

			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Do you want to replace file ", file.Name(), "\n[y] yes / [n] no: ")
			reply, _ := reader.ReadString('\n')

			if strings.Contains(strings.ToLower(reply), "n") {
				continue
			}

		}

		debug("Archiving:", file.targetPath+file.Name())
		file.copyFile()
	}
}

//Structure representing a file which will be archive.
type ArchiveFile struct {
	os.FileInfo
	sourcePath string
	targetPath string
}

func (s *ArchiveFile) fileExists() bool {
	_, err := os.Stat(s.targetPath + s.Name())
	return err == nil
}

func (s *ArchiveFile) parseTargetPath() {

	year, month, day, err := s.extractPictureTakenDatepart()

	if year == "" || err != nil {
		debug("Unable to parse dateattribute \"Picture taken\" from file:", s.Name(), ". Using \"Modified date\".")
		year, month, day = s.extractFileModifiedDate()
	}

	matches := parseDates(targetPatternRegex, cmdPrm.TargetPattern)

	var pathBuff bytes.Buffer
	pathBuff.WriteString(cmdPrm.Target)

	if matches["YYYY"] != "" {
		pathBuff.WriteString(year + "/")
	}

	if matches["MM"] != "" {
		pathBuff.WriteString(month + "/")
	}

	if matches["DD"] != "" {
		pathBuff.WriteString(day + "/")
	}

	s.targetPath = pathBuff.String()
}

func (s *ArchiveFile) extractPictureTakenDatepart() (year string, month string, day string, err error) {
	data, err := exif.Read(s.sourcePath)

	if err != nil {
		return "", "", "", err
	}

	val := data.Tags["Date and Time (Original)"]
	debug("Checking date:Date and Time (Original)")
	matches := parseDates(dateRegex, val)

	if matches["YYYY"] == "" {
		val = data.Tags["Date and Time (Digitized)"]
		debug("Checking date: Date and Time (Digitized)")
		matches = parseDates(dateRegex, val)
	}

	return matches["YYYY"], matches["MM"], matches["DD"], nil
}
func (s *ArchiveFile) extractFileModifiedDate() (year string, month string, day string) {
	year = strconv.Itoa(s.ModTime().Year())
	month = s.paddWithLeadingZero(int(s.ModTime().Month()))
	day = s.paddWithLeadingZero(s.ModTime().Day())
	return year, month, day
}

func (s *ArchiveFile) paddWithLeadingZero(num int) string {
	if num < 10 {
		return "0" + strconv.Itoa(num)
	}

	return strconv.Itoa(num)
}

func (s *ArchiveFile) copyFile() {

	if _, err := os.Stat(s.targetPath); err != nil {
		os.MkdirAll(s.targetPath, 0777)
	}

	// open files r and w
	r, err := os.Open(s.sourcePath)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	targetFile := s.targetPath + s.Name()
	w, err := os.Create(targetFile)

	if err != nil {
		panic(err)
	}
	defer w.Close()

	// do the actual work
	n, err := io.Copy(w, r)
	if err != nil {
		panic(err)
	}

	os.Chtimes(targetFile, s.ModTime(), s.ModTime())

	debug("Copied ", strconv.FormatInt(n, 10), " bytes for ", s.Name())
}

//Structure stores configurationparameters provided either through commandline or parameterfile.
type CmdLineParams struct {
	ConfigFile    string
	ShallArchive  bool
	Verbose       bool
	FilePattern   string
	Source        string
	Target        string
	Overwrite     bool
	TargetPattern string
}

func (c *CmdLineParams) ReadConfig() {
	fileContent, e := ioutil.ReadFile(c.ConfigFile)

	if e != nil {
		fmt.Println("Could not load configfile.", e)
		return
	}

	json.Unmarshal(fileContent, &c)

}
func (c CmdLineParams) String() string {
	return "[configfile=" + c.ConfigFile +
		", source=" + c.Source +
		", target=" + c.Target +
		", pattern=" + c.FilePattern +
		", targetpattern=" + c.TargetPattern +
		", verbose=" + strconv.FormatBool(cmdPrm.Verbose) +
		", archive=" + strconv.FormatBool(c.ShallArchive) +
		", overwrite=" + strconv.FormatBool(c.Overwrite) +
		"]"
}

func parseDates(regx *regexp.Regexp, val string) map[string]string {
	matches := make(map[string]string)
	match := regx.FindStringSubmatch(val)

	if match == nil {
		return matches
	}

	for i, name := range regx.SubexpNames() {

		if name == "" {
			continue
		}

		matches[name] = match[i]
	}
	return matches
}

func debug(msg ...string) {
	if cmdPrm.Verbose {
		fmt.Println(msg)
	}
}
