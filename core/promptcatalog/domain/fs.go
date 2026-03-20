package domain

import "os"

type fileInfo = os.FileInfo

var osStat = os.Stat
var isNotExist = os.IsNotExist
var sameFile = os.SameFile
