// This file is part of git-dch-go.
//
// git-dch-go is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// git-dch-go is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with git-dch-go.  If not, see <http://www.gnu.org/licenses/>.

// git-dch, permits to generate debian changelog
// from git commits in a repository.
//
// Created: 2018-02-15

// http://honk.sigxcpu.org/projects/git-buildpackage/manual-html/man.gbp.dch.html
// https://manpages.debian.org/jessie/git-buildpackage/git-dch.1.en.html
//
// gbp dch [ --version ] [ --help ] [ --verbose ] [ --color= [auto|on|off] ] [ --color-scheme=COLOR_SCHEME ]
//   [ --debian-branch=branch_name ] [ --debian-tag=tag-format ] [ --upstream-tag=tag-format ] [ --ignore-branch ]
//   [ --snapshot | --release ] [ --auto | --since=commitish ] [ --new-version=version ]
//   [ --bpo | --nmu | --qa | --team ] [ --distribution=name ] [ --force-distribution ] [ --urgency=level ]
//   [ --[no-]full ] [ --[no-]meta ] [ --meta-closes=bug-close-tags ] [ --snapshot-number=expression ]
//   [ --id-length=number ] [ --git-log=git-log-options ] [ --[no-]git-author ] [ --[no-]multimaint ]
//   [ --[no-]multimaint-merge ] [ --spawn-editor=[always|snapshot|release] ] [ --commit-msg=msg-format ] [ --commit ]
//   [ --customizations= customization-file ] [path1 path2]

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gitlab.yuribugelli.it/debian/git-dch-go/pkg/changelog"
	"gitlab.yuribugelli.it/debian/git-dch-go/pkg/dchversion"
	"gitlab.yuribugelli.it/debian/git-dch-go/pkg/git"

	"github.com/jessevdk/go-flags"
)

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

var (
	version, commitHash string
	buildDate           = "20180101_000000_+0000_GMT"
	parsedDate          time.Time
)

var (
	options Options
	gr      git.Repository
)

func init() {
	var err error
	if parsedDate, err = time.Parse("20060102_150405_-0700_MST", buildDate); err != nil {
		panic(err)
	}
}

func main() {
	var err error

	if _, err = checkOptions(); err != nil {
		printError("ERROR: %s\n", err)
	}

	if gr, err = git.NewRepositoryFromCurrentDirectory(); err != nil {
		printError("ERROR: %s\n", err)
	}

	var author string
	if author, err = getAuthor(); err != nil {
		printError("ERROR: %s\n", err)
	}

	if err = updateChangelog(author); err != nil {
		printError("ERROR: %s\n", err)
	}
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
	fmt.Printf("Timestamp: %s\n", parsedDate)
	os.Exit(0)
}

func printError(format string, a ...interface{}) {
	if _, err := fmt.Fprintf(os.Stderr, format, a); err != nil {
		fmt.Printf(format, a)
	}
	os.Exit(1)
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
		printError("the given version (%s) has wrong format: %s\n", options.NewVersion, err)
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
		printError("ERROR: %s\n", err)
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
