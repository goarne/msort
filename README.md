# About MSORT

This is a commandline application for organizing mediafiles into folders based on "picture taken" date or "created date".
Sample target structure

###source folderstructure
```
<workingdirectory>/somepictures/img_1.jpg
<workingdirectory>/somepictures/img_2.jpg
<workingdirectory>/somepictures/img_4.jpg
<workingdirectory>/somepictures/img_5.jpg
```
after running msort

###target folderstructure
```
<workingdirectory>/2010/01/02/img_1.jpg
<workingdirectory>/2010/01/02/img_2.jpg
<workingdirectory>/2010/02/03/img_3.jpg
<workingdirectory>/2010/02/04/img_4.jpg
<workingdirectory>/2015/06/04/img_5.jpg
```

The program support commandline arguments aswell as storing parameters in a json file. 

## Usage with parameters from commandline
```
msort -source=./  -target=./ -pattern=* -targetpattern=yyyy/mm/dd -verbose=false -archive=false -overwrite=false
```
###Parameters supported
```
source: path to root of sourcearchive which to start traversal
target: path to root of destination to build targetpath
pattern: regex for filepatternmatting
targetpattern: destinationfolder structure ex. yyyy/mm/dd or yyyy/mm or yyyy
verbose: more debug information
archive: do the actual copy of the file
overwrite: the application can overwrites existing destination file if user accepts.
```
## Usage with parameters from configurationfille
```
msort -configfile=./msort.config.json
```
###Parameters in file
```
with jsonfile:

{
	"shallarchive":true,
	"target":"archived/",
	"targetpattern":"yyyy/mm/dd/",
	"overwrite":true
}
```








