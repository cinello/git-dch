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
	"strings"
	"time"

	"github.com/cinello/git-dch/pkg/dchversion"

	"github.com/cinello/go-debian/changelog"
	"github.com/cinello/go-debian/version"
)

// Item encapsulate a debian changelog entry
type Item struct {
	changelog.ChangelogEntry
}

// NewItem creates a new Item variable, this function accept the following parameters:
//
//   source: is the package name
//   ver: is the version for the new release identified by this entry
//   target: is the distribution this package is released for
//   urgency: the urgency level of the package
//   changelog: the text describing the changes in this release
//   author: the author of thi release
//
// The function returns a new Item variable. If the variable cannot be created because a
// parameter is empty or wrong, an error is returned
func NewItem(source string, ver dchversion.Version, target, urgency, changelog, author string) (cl Item, err error) {

	if err = cl.SetSource(source); err != nil {
		return
	}

	newVersion := version.Version{Epoch: ver.Epoch(), Version: ver.Version(), Revision: ver.Revision()}
	if err = cl.SetVersion(newVersion); err != nil {
		return
	}

	if err = cl.SetTarget(target); err != nil {
		return
	}

	cl.Arguments = make(map[string]string)
	if err = cl.SetUrgency(urgency); err != nil {
		return
	}

	cl.SetChangelog(changelog)

	if err = cl.SetAuthor(author); err != nil {
		return
	}

	cl.When = time.Now()

	return
}

// isStringValid is used to validate text strings passed as parameters
// in a changelog entry. These strings must not be empty, and must not
// contain spaces.
func isStringValid(value string) bool {

	if value == "" {
		return false
	}

	if strings.Contains(value, " ") {
		return false
	}

	return true
}

// SetSource method change the Source field of an Item variable
func (e *Item) SetSource(value string) error {

	if !isStringValid(value) {
		return fmt.Errorf("changelog Source %s is not valid", value)
	}

	e.Source = value

	return nil
}

// isVersionValid is used to validate a version passed to a changelog
// entry. A valid version must not be an empty string in the Version
// field. It can contain any value in Epoch an Revision fields.
func isVersionValid(value version.Version) bool {

	if value.Version == "" {
		return false
	}

	return true
}

// SetVersion method change the Version field of an Item variable
func (e *Item) SetVersion(value version.Version) error {

	if !isVersionValid(value) {
		return fmt.Errorf("changelog Version %s is not valid", value.String())
	}

	e.Version = value

	return nil
}

// SetTarget method change the Target field of an Item variable
func (e *Item) SetTarget(value string) error {

	if !isStringValid(value) {
		return fmt.Errorf("changelog Target %s is not valid", value)
	}

	e.Target = value

	return nil
}

// SetUrgency method change the Urgency field of an Item variable
func (e *Item) SetUrgency(value string) error {

	if value == "" {
		e.Arguments["urgency"] = "medium"
		return nil
	}

	if !isStringValid(value) {
		return fmt.Errorf("changelog Urgency %s is not valid", value)
	}

	e.Arguments["urgency"] = value

	return nil
}

// SetChangelog method change the Changelog field of an Item variable
// If the changelog is empty, its automatically converted to a minimal
// valid text (  * )
func (e *Item) SetChangelog(value string) {

	if value == "" {
		value = "  *"
	}
	// Add newlines before and after changelog entries
	for !strings.HasPrefix(value, "\n") {
		value = "\n" + value
	}
	for !strings.HasSuffix(value, "\n\n") {
		value += "\n"
	}

	e.Changelog = value
}

// isAuthorValid is used to validate the author of a changelog entry.
// A valid author must not be an empty string.
func isAuthorValid(value string) bool {

	if value == "" {
		return false
	}

	return true
}

// SetAuthor method change the Author field of an Item variable
func (e *Item) SetAuthor(value string) error {

	if !isAuthorValid(value) {
		return fmt.Errorf("changelog author cannot be empty")
	}

	e.ChangedBy = value

	return nil
}

// SetWhen method change the When field of an Item variable
func (e *Item) SetWhen(value time.Time) {

	e.When = value
}

// WhenToString method convert a time.Time value to a valid
// text representation in RFC1123Z format
func (e *Item) WhenToString() string {

	return WhenToString(e.When)
}

// WhenToString function convert a time.Time value to a valid
// text representation in RFC1123Z format
func WhenToString(t time.Time) string {

	return t.Format(time.RFC1123Z)
}

// SetArgument method create or change an Argument in an Item
// variable. Urgency is a standard argument, other not standard
// arguments can be created by this method
func (e *Item) SetArgument(key, value string) error {

	if !isStringValid(key) {
		return fmt.Errorf("changelog argument %s key is not valid", key)
	}

	if !isStringValid(value) {
		return fmt.Errorf("changelog argument value %s for key %s is not valid", value, key)
	}

	e.Arguments[key] = value

	return nil
}

// areArgumentsValid is used to validate all the arguments of an Item variable.
func areArgumentsValid(args map[string]string) bool {

	for _, a := range args {
		if !isStringValid(a) {
			return false
		}
	}

	return true
}

// isValid validate all the fields of an ite variable..
func (e *Item) isValid() bool {

	if !isStringValid(e.Source) {
		return false
	}
	if !isVersionValid(e.Version) {
		return false
	}
	if !isStringValid(e.Target) {
		return false
	}
	if !areArgumentsValid(e.Arguments) {
		return false
	}
	if !isAuthorValid(e.ChangedBy) {
		return false
	}

	return true
}

// String converts an Item variable to a valid string
// representation ready to be saved in a changelog file
func (e *Item) String() (out string) {

	if !e.isValid() {
		return ""
	}

	out = fmt.Sprintf("%s (%s) %s;", e.Source, e.Version, e.Target)
	for k, v := range e.Arguments {
		out = out + " " + k + "=" + v
	}
	out = out + "\n" + e.Changelog
	out = out + " -- " + e.ChangedBy + "  " + e.WhenToString()
	return out + "\n"
}

// Items is a slice of Item
// Is used to represent the entire changelog file
type Items []Item

// NewItemList returns a slice of Items, reading the contents from a Reader interface
func NewItemList(reader io.Reader) (Items, error) {

	entries, err := changelog.Parse(reader)
	if err != nil {
		return nil, err
	}

	e := Items{}
	for _, entry := range entries {
		e = append(e, Item{entry})
	}

	return e, nil
}

// NewItemList returns a slice of Items, reading the contents a file at the given path
func NewItemListFromFile(path string) (Items, error) {

	entries, err := changelog.ParseFile(path)
	if err != nil {
		return nil, err
	}

	e := Items{}
	for _, entry := range entries {
		e = append(e, Item{entry})
	}

	return e, nil
}

// Source return the Source field of the (chronologically) last Item in the slice
func (e *Items) Source() string {
	if e != nil && len(*e) > 0 {
		return (*e)[0].Source
	}

	return ""
}

// Source return the Version field of the (chronologically) last Item in the slice
func (e *Items) Version() version.Version {
	if e != nil && len(*e) > 0 {
		return (*e)[0].Version
	}

	return version.Version{Epoch: 0, Version: "0.0.0", Revision: "1"}
}

// Source return the Target field of the (chronologically) last Item in the slice
func (e *Items) Target() string {
	if e != nil && len(*e) > 0 {
		return (*e)[0].Target
	}

	return ""
}

// Source return the Changelog field of the (chronologically) last Item in the slice
func (e *Items) Changelog() string {
	if e != nil && len(*e) > 0 {
		return (*e)[0].Changelog
	}

	return ""
}

// Source return the Author field of the (chronologically) last Item in the slice
func (e *Items) Author() string {
	if e != nil && len(*e) > 0 {
		return (*e)[0].ChangedBy
	}

	return ""
}

// Source return the When field of the (chronologically) last Item in the slice
func (e *Items) When() time.Time {
	if e != nil && len(*e) > 0 {
		return (*e)[0].When
	}

	return time.Time{}
}

// String returns a valid text changelog, ready to be saved to a file
func (e *Items) String() string {

	var contents string
	for _, entry := range *e {
		if contents != "" {
			contents += "\n"
		}
		c := entry.String()
		contents = contents + c
	}
	return contents
}
