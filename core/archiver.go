package core

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/goarne/logging"
	"github.com/xiam/exif"
)

var (
	//Contains global configuratins for the app.
	CmdPrm = CmdLineParams{}

	//Counter for files processed.
	FileCount int

	//Handles the syncronization of goroutines
	WaitGrp = sync.WaitGroup{}

	//The regex extracts datevalues from an image file in the format yyyy:mm:dd
	dateRegex, _ = regexp.Compile("(?P<YYYY>\\d{4}):(?P<MM>\\d{2}):(?P<DD>\\d{2})")

	//The regex parses the stringvalue yyyy, mm and dd in the format yyyy/mm/dd.  The search is case insensitive. Each field is optional.
	targetPatternRegex, _ = regexp.Compile("(?P<YYYY>((?i)YYYY)?)/?(?P<MM>((?i)MM)?)/?(?P<DD>((?i)DD)?)")
)

//The function sorts a list based on modified date and returns the sorted array.
func FindFiles(fileSink chan *ArchiveFile) {
	FileCount = 0
	defer close(fileSink)
	dir := filepath.Dir(CmdPrm.Source)

	filepath.Walk(dir, func(sourcePath string, f os.FileInfo, err error) error {
		if err != nil {
			logging.Error.Println(err.Error())
			return nil
		}
		if f.IsDir() == true {
			return nil
		}

		if f.Name() == CmdPrm.ConfigFile {
			return nil //Dont want to archive configfile.
		}

		if fileMatch, err := filepath.Match(CmdPrm.FilePattern, f.Name()); fileMatch == false {
			return err
		}

		foundFile := &ArchiveFile{f, sourcePath, ""}
		debug("Found file: ", f.Name())

		fileSink <- foundFile
		FileCount++
		return nil
	})
}

//The function copy files received in the channel from source to target.
//Targetpath is calculated from cameradate.
func ArchiveFiles(fileSource chan *ArchiveFile) {
	defer WaitGrp.Done()

	for file := range fileSource {

		file.parseTargetPath()

		if CmdPrm.ShallArchive == false { //Want to check this late because of debug messages in parseTargetPath
			continue
		}

		if file.fileExists() {
			if CmdPrm.Overwrite == false { //Dont want to copy unnescessary
				continue
			}
			debug("File ", file.TargetPath, file.Name(), "exists.")

			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Do you want to replace file ", file.Name(), "\n[y] yes / [n] no: ")
			reply, _ := reader.ReadString('\n')

			if strings.Contains(strings.ToLower(reply), "n") {
				continue
			}

		}

		debug("Archiving:", file.TargetPath+file.Name())
		file.copyFile()
	}
}

//Structure representing a file which will be archive.
type ArchiveFile struct {
	os.FileInfo
	SourcePath string
	TargetPath string
}

func (s *ArchiveFile) fileExists() bool {
	_, err := os.Stat(s.TargetPath + s.Name())
	return err == nil
}

func (s *ArchiveFile) parseTargetPath() {

	year, month, day, err := s.extractPictureTakenDatepart()

	if year == "" || err != nil {
		debug("Unable to parse dateattribute \"Picture taken\" from file:", s.Name(), ". Using \"Modified date\".")
		year, month, day = s.extractFileModifiedDate()
	}

	matches := parseDates(targetPatternRegex, CmdPrm.TargetPattern)

	var pathBuff bytes.Buffer
	pathBuff.WriteString(CmdPrm.Target)

	if matches["YYYY"] != "" {
		pathBuff.WriteString(year + "/")
	}

	if matches["MM"] != "" {
		pathBuff.WriteString(month + "/")
	}

	if matches["DD"] != "" {
		pathBuff.WriteString(day + "/")
	}

	s.TargetPath = pathBuff.String()
}

func (s *ArchiveFile) extractPictureTakenDatepart() (year string, month string, day string, err error) {
	data, err := exif.Read(s.SourcePath)

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

	if _, err := os.Stat(s.TargetPath); err != nil {
		os.MkdirAll(s.TargetPath, 0777)
	}

	// open files r and w
	r, err := os.Open(s.SourcePath)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	targetFile := s.TargetPath + s.Name()
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
		logging.Error.Println("Could not load configfile.", e)
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
		", verbose=" + strconv.FormatBool(CmdPrm.Verbose) +
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
	if CmdPrm.Verbose {
		logging.Trace.Println(msg)
	}
}