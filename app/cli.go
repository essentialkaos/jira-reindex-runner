package app

// ////////////////////////////////////////////////////////////////////////////////// //
//                                                                                    //
//                         Copyright (c) 2021 ESSENTIAL KAOS                          //
//      Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>     //
//                                                                                    //
// ////////////////////////////////////////////////////////////////////////////////// //

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	"pkg.re/essentialkaos/ek.v12/fmtc"
	"pkg.re/essentialkaos/ek.v12/fsutil"
	"pkg.re/essentialkaos/ek.v12/knf"
	"pkg.re/essentialkaos/ek.v12/log"
	"pkg.re/essentialkaos/ek.v12/options"
	"pkg.re/essentialkaos/ek.v12/usage"
	"pkg.re/essentialkaos/ek.v12/usage/completion/bash"
	"pkg.re/essentialkaos/ek.v12/usage/completion/fish"
	"pkg.re/essentialkaos/ek.v12/usage/completion/zsh"
	"pkg.re/essentialkaos/ek.v12/usage/man"
	"pkg.re/essentialkaos/ek.v12/usage/update"

	knfv "pkg.re/essentialkaos/ek.v12/knf/validators"
	knff "pkg.re/essentialkaos/ek.v12/knf/validators/fs"
	knfn "pkg.re/essentialkaos/ek.v12/knf/validators/network"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Basic application info
const (
	APP  = "JiraReindexRunner"
	VER  = "0.0.1"
	DESC = "Application for periodical running Jira re-index process"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// Options
const (
	OPT_CONFIG   = "c:config"
	OPT_NO_COLOR = "nc:no-color"
	OPT_HELP     = "h:help"
	OPT_VER      = "v:version"

	OPT_COMPLETION   = "completion"
	OPT_GENERATE_MAN = "generate-man"
)

// Configuration file properties
const (
	MAIN_ENABLED      = "main:enabled"
	JIRA_URL          = "jira:url"
	JIRA_USERNAME     = "jira:username"
	JIRA_PASSWORD     = "jira:password"
	JIRA_REINDEX_TYPE = "jira:reindex-type"
	LOG_DIR           = "log:dir"
	LOG_FILE          = "log:file"
	LOG_PERMS         = "log:perms"
	LOG_LEVEL         = "log:level"
)

// ////////////////////////////////////////////////////////////////////////////////// //

// optMap contains information about all supported options
var optMap = options.Map{
	OPT_CONFIG:   {Value: "/etc/jira-reindex-runner.knf"},
	OPT_NO_COLOR: {Type: options.BOOL},
	OPT_HELP:     {Type: options.BOOL, Alias: "u:usage"},
	OPT_VER:      {Type: options.BOOL, Alias: "ver"},

	OPT_COMPLETION:   {},
	OPT_GENERATE_MAN: {Type: options.BOOL},
}

// useRawOutput is raw output flag (for cli command)
var useRawOutput = false

// ////////////////////////////////////////////////////////////////////////////////// //

// Init is main function
func Init() {
	runtime.GOMAXPROCS(1)

	_, errs := options.Parse(optMap)

	if len(errs) != 0 {
		for _, err := range errs {
			printError(err.Error())
		}

		os.Exit(1)
	}

	preConfigureUI()

	if options.Has(OPT_COMPLETION) {
		os.Exit(genCompletion())
	}

	if options.Has(OPT_GENERATE_MAN) {
		os.Exit(genMan())
	}

	configureUI()

	if options.GetB(OPT_VER) {
		os.Exit(showAbout())
	}

	if options.GetB(OPT_HELP) {
		os.Exit(showUsage())
	}

	loadConfig()
	validateConfig()
	setupLogger()

	if knf.GetB(MAIN_ENABLED, false) {
		os.Exit(0)
	}

	log.Aux(strings.Repeat("-", 80))
	log.Aux("%s %s starting…", APP, VER)

	process()
}

// preConfigureUI preconfigures UI based on information about user terminal
func preConfigureUI() {
	term := os.Getenv("TERM")

	fmtc.DisableColors = true

	if term != "" {
		switch {
		case strings.Contains(term, "xterm"),
			strings.Contains(term, "color"),
			term == "screen":
			fmtc.DisableColors = false
		}
	}

	if !fsutil.IsCharacterDevice("/dev/stdout") && os.Getenv("FAKETTY") == "" {
		fmtc.DisableColors = true
		useRawOutput = true
	}

	if os.Getenv("NO_COLOR") != "" {
		fmtc.DisableColors = true
	}
}

// configureUI configures user interface
func configureUI() {
	if options.GetB(OPT_NO_COLOR) {
		fmtc.DisableColors = true
	}
}

// loadConfig reads and parses configuration file
func loadConfig() {
	err := knf.Global(options.GetS(OPT_CONFIG))

	if err != nil {
		printErrorAndExit(err.Error())
	}
}

// validateConfig validates configuration file values
func validateConfig() {
	errs := knf.Validate([]*knf.Validator{
		{JIRA_URL, knfv.Empty, nil},
		{JIRA_USERNAME, knfv.Empty, nil},
		{JIRA_PASSWORD, knfv.Empty, nil},

		{JIRA_URL, knfn.URL, nil},

		{JIRA_REINDEX_TYPE, knfv.NotContains, []string{
			"", "FOREGROUND", "BACKGROUND", "BACKGROUND_PREFERRED",
		}},

		{LOG_DIR, knff.Perms, "DW"},
		{LOG_DIR, knff.Perms, "DX"},

		{LOG_LEVEL, knfv.NotContains, []string{
			"debug", "info", "warn", "error", "crit",
		}},
	})

	if len(errs) != 0 {
		printError("Error while configuration file validation:")

		for _, err := range errs {
			printError("  %v", err)
		}

		os.Exit(1)
	}
}

// setupLogger confugures logger subsystems
func setupLogger() {
	err := log.Set(knf.GetS(LOG_FILE), knf.GetM(LOG_PERMS, 644))

	if err != nil {
		printErrorAndExit(err.Error())
	}

	err = log.MinLevel(knf.GetS(LOG_LEVEL))

	if err != nil {
		printErrorAndExit(err.Error())
	}
}

// process starts processing
func process() {
	os.Exit(runReindex())
}

// printError prints error message to console
func printError(f string, a ...interface{}) {
	fmtc.Fprintf(os.Stderr, "{r}"+f+"{!}\n", a...)
}

// printError prints warning message to console
func printWarn(f string, a ...interface{}) {
	fmtc.Fprintf(os.Stderr, "{y}"+f+"{!}\n", a...)
}

// printErrorAndExit print error mesage and exit with exit code 1
func printErrorAndExit(f string, a ...interface{}) {
	printError(f, a...)
	os.Exit(1)
}

// ////////////////////////////////////////////////////////////////////////////////// //

// showUsage prints usage info
func showUsage() int {
	genUsage().Render()
	return 0
}

// showAbout prints info about version
func showAbout() int {
	genAbout().Render()
	return 0
}

// genCompletion generates completion for different shells
func genCompletion() int {
	info := genUsage()

	switch options.GetS(OPT_COMPLETION) {
	case "bash":
		fmt.Printf(bash.Generate(info, "jira-reindex-runner"))
	case "fish":
		fmt.Printf(fish.Generate(info, "jira-reindex-runner"))
	case "zsh":
		fmt.Printf(zsh.Generate(info, optMap, "jira-reindex-runner"))
	default:
		return 1
	}

	return 0
}

// genMan generates man page
func genMan() int {
	fmt.Println(
		man.Generate(
			genUsage(),
			genAbout(),
		),
	)

	return 0
}

// genUsage generates usage info
func genUsage() *usage.Info {
	info := usage.NewInfo()

	info.AddOption(OPT_CONFIG, "Path to configuration file", "config")
	info.AddOption(OPT_NO_COLOR, "Disable colors in output")
	info.AddOption(OPT_HELP, "Show this help message")
	info.AddOption(OPT_VER, "Show version")

	return info
}

// genAbout generates info about version
func genAbout() *usage.About {
	return &usage.About{
		App:           APP,
		Version:       VER,
		Desc:          DESC,
		Year:          2009,
		Owner:         "ESSENTIAL KAOS",
		License:       "Apache License, Version 2.0 <https://www.apache.org/licenses/LICENSE-2.0>",
		UpdateChecker: usage.UpdateChecker{"essentialkaos/jira-reindex-runner", update.GitHubChecker},
	}
}

// ////////////////////////////////////////////////////////////////////////////////// //
