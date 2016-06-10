# About MSORT

MSsort is a commandline application which organizes mediafiles into folders based on "picture taken" date or "created date".

It support commandline arguments aswell as storing parameters in a json file. 

## Usage
```
msort -configfile=./msort.config.json

with jsonfile:

{
	"shallarchive":true,
	"target":"archived/",
	"targetpattern":"yyyy/mm/dd/",
	"overwrite":true
}

```
or
```
msort -source=./  -target=./ -pattern=* -targetpattern=yyyy/mm/dd -verbose=false -archive=false -overwrite=false

source: path to root of sourcearchive which to start traversal
target: path to root of destination to build targetpath
pattern: regex for filepatternmatting
targetpattern: destinationfolder structure
verbose: more debug information
archive: do the actual copy of the file
overwrite: if true, the application can overwrites existing destination file if user accepts.
```







