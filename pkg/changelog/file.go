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

package changelog

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gitlab.yuribugelli.it/debian/git-dch-go/pkg/dchversion"
	"gitlab.yuribugelli.it/debian/git-dch-go/pkg/git"
)

var (
	regExSnapshot = regexp.MustCompile(`\s{2}\*\* SNAPSHOT build @([a-f0-9]{40}) \*\*`)
)

// File struct contains all the entries of a changelog
type File struct {
	el Items
}

// New function create a new File struct reading the contents from a Reader interface
func New(reader io.Reader) (*File, error) {

	entries, err := NewItemList(reader)
	if err != nil {
		return nil, err
	}

	return &File{el: entries}, nil
}

// NewFromFile function create a new File struct reading the contents from a file at the given path
func NewFromFile(path string) (*File, error) {

	entries, err := NewItemListFromFile(path)
	if err != nil {
		return nil, err
	}

	return &File{el: entries}, nil
}

func (f *File) computeNewVersion(v dchversion.Version) (newVersion dchversion.Version, err error) {

	newVersion = v

	if f.IsEmpty() {
		// If the changelog file has no entries, ensure the first version has
		// the correct value format
		switch {
		case newVersion.IsSnapshot():
			newVersion.SetRevision("")
		case newVersion.IsStaging():
			newVersion.SetRevision("1")
		case newVersion.IsDevelopment():
			newVersion.SetRevision("1")
		case newVersion.IsNative():
			newVersion.SetRevision("1")
			return
		}
		return
	}

	old := dchversion.NewVersionFromDebian(f.el.Version())

	// newVersion is native, extract native from old version and compare
	// if native new = old, then increment old
	if !v.IsSnapshot() && v.Revision() == "" {
		oldNative := old.ExtractNative()
		newNative := newVersion.ExtractNative()
		if dchversion.Compare(newNative, oldNative) < 0 {
			err = fmt.Errorf("the new version %s is lesser than the old version %s", v.String(), old.String())
		}
		if dchversion.Compare(newNative, oldNative) == 0 {
			newVersion, err = old.IncrementRevision()
		}
		return
	}

	// If both newVersion and oldVersion are snapshot, this must be handled
	// as a special case, using dedicated functions
	if v.IsSnapshot() && old.IsSnapshot() {
		var r int
		if r, err = dchversion.GetSnapshotRelease(old); err != nil {
			return
		}
		if v, err = dchversion.SetSnapshotRelease(v, r); err != nil {
			return
		}

		var c int
		if c, err = dchversion.CompareSnapshots(v, old); err != nil {
			return newVersion, err
		}
		if c == 0 {
			newVersion, err = v.IncrementRevision()
		}
		return
	}

	v.SetRevision(old.Revision())

	// if newVersion is not native, then it's the target version
	if dchversion.Compare(v, old) < 0 {
		err = fmt.Errorf("the new version %s is lesser than the old version %s", v.String(), old.String())
	}
	if dchversion.Compare(v, old) == 0 {
		newVersion, err = v.IncrementRevision()
	}
	return
}

func (f *File) computeSourceName(source string) (out string, err error) {

	out = source
	if out == "" {
		// if no source name was given in function parameters
		if !f.IsEmpty() {
			// if a previous changelog entry exists, we use the source name from it
			out = f.el.Source()
		} else {
			// if changelog is empty, we use current working directory name
			var path string
			if path, err = os.Getwd(); err != nil {
				return out, err
			}
			out = path[strings.LastIndex(path, string(filepath.Separator))+1:]
		}
	}

	if strings.Contains(out, " ") {
		return out, fmt.Errorf("source contains spaces")
	}

	if out == "" {
		return out, fmt.Errorf("target is an empty string")
	}

	return out, err
}

func (f *File) computeTargetName(target string) (out string, err error) {

	out = target
	if out == "" && !f.IsEmpty() {
		out = f.el.Target()
	}

	if strings.Contains(out, " ") {
		return out, fmt.Errorf("target contains spaces")
	}

	if out == "" {
		return out, fmt.Errorf("target is an empty string")
	}

	return
}

func (f *File) computeAuthor(author string) (out string, err error) {

	out = author
	if out == "" && !f.IsEmpty() {
		out = f.el.Author()
	}

	if out == "" {
		return out, fmt.Errorf("target is an empty string")
	}

	return
}

// IsEmpty returns true if the File ChangelogEntries contains no data
func (f *File) IsEmpty() bool {

	return len(f.el) == 0
}

// Len returns the number o entries File ChangelogEntries slice
func (f *File) Len() int {

	return len(f.el)
}

// LastVersion function return the version number of the last args in the changelog
func (f *File) LastVersion() (v dchversion.Version, err error) {

	if f.IsEmpty() {
		err = fmt.Errorf("the changelog file is empty, cannot get last release version")
		return
	}

	v = dchversion.NewVersionFromDebian(f.el.Version())
	return
}

// Write function write all the contents of the File slice to a Writer interface
func (f *File) Write(writer io.Writer) (int, error) {

	text := f.el.String()
	return writer.Write([]byte(text))
}

// WriteToFile function save all the changes of the File slice into a text file at the given path
func (f *File) WriteToFile(path string) (n int, err error) {

	var file *os.File

	if file, err = os.Create(path); err != nil {
		return
	}

	defer func() {
		if closeErr := file.Close(); err == nil {
			err = closeErr
		}
	}()

	n, err = file.WriteString(f.el.String())

	return
}

func (f *File) addSimple(
	source string,
	ver dchversion.Version,
	urgency, target, clog, author string,
) (v dchversion.Version, entry Item, err error) {

	var s string
	if s, err = f.computeSourceName(source); err != nil {
		return
	}

	if v, err = f.computeNewVersion(ver); err != nil {
		return
	}

	var t string
	if t, err = f.computeTargetName(target); err != nil {
		return
	}

	var a string
	if a, err = f.computeAuthor(author); err != nil {
		return
	}

	entry, err = NewItem(s, v, t, urgency, clog, a)

	f.el = append(Items{entry}, f.el...)

	return v, entry, nil
}

func (f *File) getLog(since string, auto bool, ignoreMerges bool) (out string, err error) {

	var gr git.Repository
	if gr, err = git.NewRepositoryFromCurrentDirectory(); err != nil {
		return
	}

	if since != "" {
		return gr.LogToCommit(since, false, false, true, false, ignoreMerges)
	}

	if !f.IsEmpty() && auto {
		// 1) The start commit is read from the snapshot banner (see below for details)
		v := dchversion.NewVersionFromDebian(f.el.Version())
		if v.Type() == dchversion.Snapshot {
			values := regExSnapshot.FindAllStringSubmatch(f.el.Changelog(), -1)
			if len(values) != 1 || len(values[0]) != 2 {
				err = fmt.Errorf("cannot find valit commit hash in the last snapshot args")
				return
			}
			return gr.LogToCommit(values[0][1], false, false, true, false, ignoreMerges)
		}

		// 2) If the topmost version of the debian/changelog is already tagged. Use the commit the tag points to as start commit.
		var commit string
		if commit = gr.CommitAtTagObject(f.el.Version()); commit == "" {
			commit = gr.CommitAtTag(f.el.Version())
		}
		if commit != "" {
			return gr.LogToCommit(commit, false, false, true, false, ignoreMerges)
		}

		// 3) the last git commit after the last changelog release is used as start commit.
		return gr.LogToTime(f.el.When(), false, false, true, false, ignoreMerges)
	}

	// 4) get all the entries
	return gr.Log(false, false, true, false, ignoreMerges)
}

func (f *File) buildReleaseLog(since string, ver dchversion.Version, auto, ignoreMerges bool) (out string, err error) {

	ver.SetEpoch(0)
	ver.SetRevision("")
	out = fmt.Sprintf("  ** Release version %s\n\n", ver.String())

	var log string
	if log, err = f.getLog(since, auto, ignoreMerges); err != nil {
		return
	}
	out += log

	return
}

func (f *File) buildSnapshotLog(since string, auto, ignoreMerges bool) (out string, err error) {

	var gr git.Repository
	if gr, err = git.NewRepositoryFromCurrentDirectory(); err != nil {
		return
	}
	var hash string
	if hash, err = gr.LastCommitHash(-1); err != nil {
		return
	}

	out = fmt.Sprintf("  ** SNAPSHOT build @%s **\n\n", hash)

	var log string
	if log, err = f.getLog(since, auto, ignoreMerges); err != nil {
		return
	}
	out += log

	return
}

func (f *File) purgeReleases(testing, unstable bool) {
	for i := 0; i < len(f.el); i++ {
		v := dchversion.NewVersionFromDebian(f.el.Version())
		isTesting := v.IsStaging() && testing
		isUnstable := v.IsDevelopment() && unstable
		if isTesting || isUnstable {
			f.el = append(f.el[:i], f.el[i+1:]...)
			i-- // Since we just deleted f.entries[i], we must redo that index
		}
	}
}

func (f *File) purgeSnapshotReleases() {

	for i := 0; i < len(f.el); i++ {
		v := dchversion.NewVersionFromDebian(f.el.Version())
		if v.IsSnapshot() {
			f.el = append(f.el[:i], f.el[i+1:]...)
			i-- // Since we just deleted f.entries[i], we must redo that index
		}
	}
}

// AddSnapshot function create a new snapshot changelog entry in the File ChangelogEntries slice
// This function accept the following parameters:
// since string:      the reference to the commit where the log must start from
// source string:     the name of the package (if empty is guessed from the previous entries)
// ver Version:       the version number for the new release
// author string:     the author name/email for the new package (if empty is guessed from the previous entries)
// ignoreMerges bool: if true, all the merge commits are omitted from the changelog
func (f *File) AddSnapshot(
	since, source string,
	ver dchversion.Version,
	author string,
	auto, ignoreMerges bool,
) (v dchversion.Version, entry Item, err error) {

	switch {
	case ver.IsStaging():
		fallthrough
	case ver.IsDevelopment():
		err = fmt.Errorf("cannot use value %s as snapshot version", ver.String())
		return
	default:
		if ver, err = ver.Build(dchversion.Snapshot); err != nil {
			err = fmt.Errorf("cannot create a snapshot version from value %s: %s", ver.String(), err)
			return
		}
	}

	var clog string
	if clog, err = f.buildSnapshotLog(since, auto, ignoreMerges); err != nil {
		return
	}

	return f.addSimple(source, ver, "low", "UNRELEASED", clog, author)
}

// AddRelease function create a new release changelog entry in the File ChangelogEntries slice
// The function always purge all the snapshots entries in the slice before add the new one
// This function accept the following parameters:
// since string:       the reference to the commit where the log must start from
// source string:      the name of the package (if empty is guessed from the previous entries)
// ver Version:        the version number for the new release
// urgency string:     the urgency identifier for the new release (is empty is set to medium)
// target string:      the distribution identifier for the new release (if empty is guessed from the previous entries)
// author string:      the author name/email for the new package (if empty is guessed from the previous entries)
// ignoreMerges bool:  if true, all the merge commits are omitted from the changelog
// purgeTesting bool:  if true, all the testing entries are purged from the changelog before the new one is added
// purgeUnstable bool: if true, all the unstable entries are purged from the changelog before the new one is added
func (f *File) AddRelease(
	since, source string,
	ver dchversion.Version,
	urgency, target, author string,
	auto, ignoreMerges, purgeTesting, purgeUnstable bool,
) (v dchversion.Version, entry Item, err error) {

	if ver.IsNative() {
		releaseType := dchversion.Release
		if target == "unstable" {
			releaseType = dchversion.Staging
		}

		if ver, err = ver.Build(releaseType); err != nil {
			err = fmt.Errorf("cannot create a %s version from value %s: %s", releaseType.SourceBranch(), ver.String(), err)
			return
		}
	}

	f.purgeReleases(purgeTesting, purgeUnstable)
	f.purgeSnapshotReleases()

	var clog string
	if clog, err = f.buildReleaseLog(since, ver, auto, ignoreMerges); err != nil {
		return
	}

	return f.addSimple(source, ver, urgency, target, clog, author)
}

// Add function create a new generic changelog entry in the File ChangelogEntries slice
// This function accept the following parameters:
// since string:       the reference to the commit where the log must start from
// source string:      the name of the package (if empty is guessed from the previous entries)
// ver Version:        the version number for the new release
// urgency string:     the urgency identifier for the new release (is empty is set to medium)
// target string:      the distribution identifier for the new release (if empty is guessed from the previous entries)
// author string:      the author name/email for the new package (if empty is guessed from the previous entries)
// ignoreMerges bool:  if true, all the merge commits are omitted from the changelog
// purgeTesting bool:  if true, all the testing entries are purged from the changelog before the new one is added
// purgeUnstable bool: if true, all the unstable entries are purged from the changelog before the new one is added
func (f *File) Add(
	since, source string,
	ver dchversion.Version,
	urgency, target, author string,
	auto, ignoreMerges, purgeTesting, purgeUnstable bool,
) (v dchversion.Version, c Item, err error) {

	f.purgeReleases(purgeTesting, purgeUnstable)

	var clog string
	if clog, err = f.getLog(since, auto, ignoreMerges); err != nil {
		return
	}

	return f.addSimple(source, ver, urgency, target, clog, author)
}
