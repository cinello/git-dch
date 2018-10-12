package git_dch

import (
	"fmt"
	"os"
	"path/filepath"

	"gitlab.yuribugelli.it/debian/git-dch-go/pkg/changelog"
	"gitlab.yuribugelli.it/debian/git-dch-go/pkg/dchversion"
	"gitlab.yuribugelli.it/debian/git-dch-go/pkg/git"

	"github.com/jessevdk/go-flags"
)

//go:generate go run ../../../internal/autogen -p git_dch

const (
	standardChangelogFile = "./debian/changelog"
)

var (
	dTesting = []string{
		"testing",
	}
	dUnstable = []string{
		"unstable",
	}
	dStable = []string{
		// Ubuntu
		"warty",
		"hoary",
		"breezy",
		"dapper",
		"edgy",
		"feisty",
		"gutsy",
		"hardy",
		"intrepid",
		"jaunty",
		"karmic",
		"lucid",
		"maverick",
		"natty",
		"oneiric",
		"precise",
		"quantal",
		"rairing",
		"saucy",
		"trusty",
		"utopic",
		"vivid",
		"willy",
		"xenial",
		"yakkety",
		"zesty",
		"artful",
		"bionic",
		// Debian
		"stable",
		"hamm",
		"slink",
		"potato",
		"woody",
		"sarge",
		"etch",
		"lenny",
		"squeeze",
		"wheezy",
		"jessie",
		"stretch",
		"buster",
	}
)

var (
	options Options
	gr      git.Repository
)

// Options is a struct containing all the accepted command line options
type Options struct {
	Auto              bool   `short:"a" long:"auto" description:"autocomplete changelog from last snapshot or tag"`
	Distribution      string `long:"distribution" description:"Set distribution" default:"unstable" value-name:"DISTRIBUTION"`
	ForceBranch       string `long:"force-branch" description:"Force the branch name to use while generating the changelog" default:"" value-name:"branch"`
	ForceDistribution bool   `long:"force-distribution" description:"Force the provided distribution to be used, even if it doesn't match the list of known distributions"`
	GitAuthor         bool   `long:"git-author" description:"Use name and email from git-config for changelog trailer, default is 'False'"`
	IgnoreMerges      bool   `long:"ignore-merges" description:"Ignore the merge commits in git history"`
	NewVersion        string `short:"N" long:"new-version" description:"use this as base for the new version number" default:"" value-name:"NEW_VERSION"`
	PurgeUnstable     bool   `long:"purge-unstable" description:"Purge from changelog file the old release unstable releases"`
	PurgeTesting      bool   `long:"purge-testing" description:"Purge from changelog file the old release testing releases"`
	Release           bool   `short:"R" long:"release" description:"mark as release"`
	Since             string `long:"since" description:"commit to start from (e.g. HEAD^^^, debian/0.4.3)" default:"" value-name:"SINCE"`
	Snapshot          bool   `short:"S" long:"snapshot" description:"mark as snapshot build"`
	Urgency           string `long:"urgency" description:"Set urgency level" default:"medium" choice:"low" choice:"medium" choice:"high" choice:"emergency" choice:"critical" value-name:"URGENCY"`
	Version           bool   `short:"v" long:"version" description:"show program's version number and exit"`

	Args struct {
		Filename string
	} `positional-args:"yes"`
}

func RunApplication() (err error) {
	if _, err = checkOptions(); err != nil {
		return err
	}

	if gr, err = git.NewRepositoryFromCurrentDirectory(); err != nil {
		return err
	}

	var author string
	if author, err = getAuthor(); err != nil {
		return err
	}

	if err = updateChangelog(author); err != nil {
		return err
	}

	return err
}

func isDistributionValidForBranch(distribution, branch string) bool {
	var list []string
	switch branch {
	case "develop":
		list = dUnstable
	case "staging":
		list = dTesting
	case "master":
		fallthrough
	case "release":
		list = dStable
	default:
		list = dUnstable
	}

	for _, d := range list {
		if d == distribution {
			return true
		}
	}
	return false
}

func printVersion() {
	fmt.Printf("Version:   %s\n", version)
	fmt.Printf("Git hash:  %s\n", commitHash)
	fmt.Printf("Timestamp: %s\n", buildDate)
	os.Exit(0)
}

func checkOptions() (args []string, err error) {
	if args, err = flags.ParseArgs(&options, os.Args[1:]); err != nil {
		return args, fmt.Errorf("cannot parse arguments on command line")
	}

	if options.Version {
		printVersion()
	}

	if options.Auto && options.Since != "" {
		return args, fmt.Errorf("options 'auto' and 'since' cannot be used together")
	}

	if options.Snapshot && options.Release {
		return args, fmt.Errorf("options 'release'  and 'snapshot' cannot be used together")
	}

	if len(args) > 0 {
		return args, fmt.Errorf("too many arguments on the command line")
	}

	return
}

func checkBranch(parsedVersion dchversion.Version, activeBranch string) error {
	var err error

	// we get the type of release from the new version to check if
	// the value is compatible with the active branch
	releaseForVersion := parsedVersion.Type()
	if !parsedVersion.IsNative() && releaseForVersion.SourceBranch() != activeBranch {
		return fmt.Errorf("cannot use version %s with branch %s", options.NewVersion, activeBranch)
	}

	if !isDistributionValidForBranch(options.Distribution, activeBranch) && !options.ForceDistribution {
		return fmt.Errorf("the distribution %s is not valid for branch %s\n"+
			"Use --force-distribution to use it anyway", options.Distribution, activeBranch)
	}

	return err
}

func getAuthor() (author string, err error) {
	var name, email string
	if name, err = gr.ConfigValue("user", "name"); err != nil {
		return author, fmt.Errorf("cannot get user name from LOCAL git configuration: %s", err)
	}
	if email, err = gr.ConfigValue("user", "email"); err != nil {
		return author, fmt.Errorf("cannot get user email from LOCAL git configuration: %s", err)
	}
	if name == "" {
		return author, fmt.Errorf("value of user.name in LOCAL git configuration is empty")
	}
	if email == "" {
		return author, fmt.Errorf("value of user.email in LOCAL git configuration is empty")
	}
	author = name + " <" + email + ">"

	return
}

func getVersion(f *changelog.File) (parsedVersion dchversion.Version, err error) {

	if options.NewVersion == "" {
		var v dchversion.Version
		if v, err = f.LastVersion(); err != nil {
			return
		}
		options.NewVersion = v.String()
	}

	// We parse the given new version number
	if parsedVersion, err = dchversion.Parse(options.NewVersion); err != nil {
		return
	}

	var activeBranch string
	// We get active branch
	if activeBranch, err = gr.ActiveBranch(); err != nil && options.ForceBranch == "" {
		return parsedVersion, fmt.Errorf("cannot get active branch from git: %s\n"+
			"Use the --force-branch parameter to fix this error", err)
	}

	// If ForceBranch option is set, we override the guessed value
	if options.ForceBranch != "" {
		activeBranch = options.ForceBranch
	}

	if !options.Snapshot {
		// We build a valid version number for the active branch
		releaseForBranch := dchversion.ReleaseTypeFromBranch(activeBranch)
		if parsedVersion, err = parsedVersion.Build(releaseForBranch); err != nil {
			return parsedVersion, fmt.Errorf("cannot build a valid version for branch %s: %s", activeBranch, err)
		}
	}

	// Check it the used branch, evaluated version  and distribution are compatible
	if err = checkBranch(parsedVersion, activeBranch); err != nil {
		return
	}

	return
}

func updateChangelog(author string) (err error) {

	filename := standardChangelogFile
	if options.Args.Filename != "" {
		filename = options.Args.Filename
	}
	// We open the debian changelog file
	var f *changelog.File
	if f, err = changelog.NewFromFile(filepath.FromSlash(filename)); err != nil {
		return fmt.Errorf("cannot open changelog file %s: %s", filename, err)
	}

	var parsedVersion dchversion.Version
	if parsedVersion, err = getVersion(f); err != nil {
		return
	}

	var v dchversion.Version
	switch {
	case options.Snapshot:
		v, _, err = f.AddSnapshot(options.Since, "", parsedVersion, author, options.Auto, options.IgnoreMerges)
	case options.Release:
		v, _, err = f.AddRelease(options.Since, "", parsedVersion, options.Urgency, options.Distribution, author,
			options.Auto, options.IgnoreMerges, options.PurgeTesting, options.PurgeUnstable)
	default:
		v, _, err = f.Add(options.Since, "", parsedVersion, options.Urgency, options.Distribution, author,
			options.Auto, options.IgnoreMerges, options.PurgeTesting, options.PurgeUnstable)
	}
	if err != nil {
		return
	}

	if _, err = f.WriteToFile(filepath.FromSlash(filename)); err != nil {
		return
	}

	fmt.Printf("New version: %s", v.String())

	return
}
