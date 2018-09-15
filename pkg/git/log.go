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
	"strings"
	"time"

	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func buildLogEntryText(c *object.Commit, withAuthor, withHash, withStar, full bool) (out string) {
	lines := strings.Split(c.Message, "\n")

	firstLine := true
	for _, line := range lines {

		if !firstLine {
			out += "\n"
		}

		// Star
		if withStar && firstLine {
			out += "  * "
		}
		if withStar && !firstLine && line != "" {
			out += "    "
		}

		// Text
		if withHash && firstLine {
			out += fmt.Sprintf("[%s] %s", c.Hash.String()[0:7], line)
		} else {
			out += line
		}

		// Author
		if withAuthor && firstLine {
			out += " (" + c.Author.Name + ")"
		}

		if firstLine {
			firstLine = false
		}

		if !full {
			break
		}
	}
	out += "\n"

	return
}

func (gr *Repository) Log(withAuthor, withHash, withStar, full, ignoreMerges bool) (out string, err error) {

	var (
		list []*object.Commit
	)

	if list, err = gr.CommitsToCommit("", ignoreMerges); err != nil {
		return
	}

	for _, c := range list {
		out += buildLogEntryText(c, withAuthor, withHash, withStar, full)
	}

	return
}

func (gr *Repository) LogToTime(t time.Time, withAuthor, withHash, withStar, full, ignoreMerges bool) (out string, err error) {

	var (
		list []*object.Commit
	)

	if list, err = gr.CommitsToTime(t, ignoreMerges); err != nil {
		return
	}

	for _, c := range list {
		out += buildLogEntryText(c, withAuthor, withHash, withStar, full)
	}

	return
}

func (gr *Repository) LogToCommit(commit string, withAuthor, withHash, withStar, full, ignoreMerges bool) (out string, err error) {

	var (
		list []*object.Commit
	)

	if list, err = gr.CommitsToCommit(commit, ignoreMerges); err != nil {
		return
	}

	for _, c := range list {
		out += buildLogEntryText(c, withAuthor, withHash, withStar, full)
	}

	return
}
