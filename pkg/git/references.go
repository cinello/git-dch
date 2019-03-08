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

package git

import (
	"fmt"
	"io"
	"strings"

	"github.com/cinello/go-debian/version"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
)

func (gr *Repository) LastCommitHash(l int) (hash string, err error) {

	var (
		i object.CommitIter
		c *object.Commit
	)

	// Obtain the commits iterator
	if i, err = gr.repository.Log(&git.LogOptions{}); err != nil {
		return
	}
	defer i.Close()

	// Get the last commit commit from the iterator
	if c, err = i.Next(); err != nil {
		return
	}

	// Get the hash of the last commit
	hash = c.Hash.String()
	if l > len(hash) || l < 0 {
		l = len(hash)
	}

	// Return the first l characters of the commit hash
	return hash[0:l], err
}

func (gr *Repository) CommitAtTag(v version.Version) string {

	// function to search a tag from a list
	f := func(tags ...string) string {

		var (
			err error
			i   storer.ReferenceIter
		)

		if i, err = gr.repository.Tags(); err != nil {
			return ""
		}
		defer i.Close()

		for {
			tag, err := i.Next()
			if err == io.EOF {
				break
			}

			for i := range tags {
				if tags[i] == tag.Name().Short() {
					return tag.Hash().String()
				}
			}
		}

		return ""
	}

	// search tag using both full debian version and upstream version
	return f(v.String(), v.Version)
}

func (gr *Repository) CommitAtTagObject(v version.Version) (commit string) {

	// function to search a tag from a list
	f := func(tags ...string) string {

		var (
			err error
			i   *object.TagIter
		)

		if i, err = gr.repository.TagObjects(); err != nil {
			return ""
		}
		defer i.Close()

		for {
			tag, err := i.Next()
			if err == io.EOF {
				break
			}

			for i := range tags {
				if tags[i] == tag.Name && tag.TargetType == plumbing.CommitObject && tag.TargetType.Valid() {
					return tag.Target.String()
				}
			}
		}

		return ""
	}

	// search tag using both full debian version and upstream version
	return f(v.String(), v.Version)
}

func (gr *Repository) CommitAtReference(name string) string {

	var (
		err error
		i   storer.ReferenceIter
	)

	if i, err = gr.repository.References(); err != nil {
		return ""
	}
	defer i.Close()

	for {
		r, err := i.Next()
		if err == io.EOF {
			break
		}

		if r.Name().Short() == name || r.Name().String() == name {
			if strings.HasPrefix(r.String(), "ref:") {
				return gr.CommitAtReference(r.Target().String())
			}
			return r.Hash().String()
		}
	}

	return ""
}

func (gr *Repository) ActiveBranch() (branch string, err error) {

	var (
		branches storer.ReferenceIter
	)

	if branches, err = gr.repository.Branches(); err != nil {
		return branch, fmt.Errorf(textCannotGetBranches, err)
	}
	defer branches.Close()

	var head *plumbing.Reference
	if head, err = gr.repository.Head(); err != nil {
		return branch, fmt.Errorf(textCannotGetHead, err)
	}

	for {
		var i *plumbing.Reference
		if i, err = branches.Next(); err == io.EOF {
			err = fmt.Errorf(textCommitIsNotValidBranch)
			break
		}

		if head.Name().String() == i.Name().String() {
			return i.Name().Short(), nil
		}
	}

	return
}
