package log

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"

	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
)

// scream so user fixes it
const warnImbalancedKey = "FIX_IMBALANCED_PAIRS"
const warnImbalancedPairs = warnImbalancedKey + " => "

func badKeyAtIndex(i int) string {
	return "BAD_KEY_AT_INDEX_" + strconv.Itoa(i)
}

// DefaultLogLog is the default log for this package.
var DefaultLog Logger

// internalLog is the logger used by logxi itself
var InternalLog Logger

type loggerMap struct {
	sync.Mutex
	loggers map[string]Logger
}

var loggers = &loggerMap{
	loggers: map[string]Logger{},
}

func (lm *loggerMap) set(name string, logger Logger) {
	lm.loggers[name] = logger
}

// logxiEnabledMap maps log name patterns to levels
var logxiNameLevelMap map[string]int

// logxiFormat is the formatter kind to create
var logxiFormat string
var home string

var isTerminal bool
var defaultContextLines = 2
var defaultFormat string
var defaultLevel int
var defaultLogxiEnv string
var defaultMaxCol = 80
var defaultPretty = true
var defaultScheme string
var defaultTimeFormat string
var disableColors bool
var disableCallstack bool
var timeFormat string
var colorableStdout = colorable.NewColorableStdout()
var isPretty = true
var isWindows = runtime.GOOS == "windows"
var wd string
var pkgMutex sync.Mutex

// logxi reserved keys
const levelKey = "_l"
const messageKey = "_m"
const nameKey = "_n"
const timeKey = "_t"
const callstackKey = "_c"

var logxiKeys = []string{levelKey, messageKey, nameKey, timeKey, callstackKey}

func setDefaults(isTerminal bool) {
	var err error
	contextLines = defaultContextLines
	wd, err = os.Getwd()
	if err != nil {
		InternalLog.Error("Could not get working directory")
	}

	if isTerminal {
		defaultLogxiEnv = "*=WRN"
		defaultFormat = FormatHappy
		defaultLevel = LevelWarn
		defaultTimeFormat = "15:04:05.000000"
	} else {
		defaultLogxiEnv = "*=ERR"
		defaultFormat = FormatJSON
		defaultLevel = LevelError
		defaultTimeFormat = "2006-01-02T15:04:05-0700"
		disableColors = true
	}
	if isWindows {
		home = os.Getenv("HOMEPATH")
		// DefaultScheme is a color scheme optimized for dark background
		// but works well with light backgrounds
		defaultScheme = "key=cyan,value,misc=blue,source=magenta,DBG,WRN=yellow,INF=green,ERR=red"
	} else {
		home = os.Getenv("HOME")
		term := os.Getenv("TERM")
		if term == "xterm-256color" {
			defaultScheme = "key=cyan+h,value,misc=blue,source=88,DBG,WRN=yellow,INF=green+h,ERR=red+h"
		} else {
			defaultScheme = "key=cyan+h,value,misc=blue,source=magenta,DBG,WRN=yellow,INF=green,ERR=red+h"
		}
	}
}

func isReservedKey(k interface{}) (bool, error) {
	// check if reserved
	if key, ok := k.(string); ok {
		for _, key2 := range logxiKeys {
			if key == key2 {
				return true, nil
			}
		}
	} else {
		return false, fmt.Errorf("Key is not a string")
	}
	return false, nil
}

func init() {
	isTerminal = isatty.IsTerminal(os.Stdout.Fd())
	setDefaults(isTerminal)

	RegisterFormatFactory(FormatHappy, formatFactory)
	RegisterFormatFactory(FormatText, formatFactory)
	RegisterFormatFactory(FormatJSON, formatFactory)
	ProcessEnv(readFromEnviron())

	// the internal log must be plain and always work
	InternalLog = NewLogger(os.Stdout, "__logxi")
	InternalLog.SetLevel(LevelError)
	InternalLog.SetFormatter(NewTextFormatter("__logxi"))
	//InternalLog = New("__logxi")

	DefaultLog = New("~")
}
