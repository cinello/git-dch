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

import "gopkg.in/src-d/go-git.v4/plumbing/object"

type sortCommitsByDate []*object.Commit

func (s sortCommitsByDate) Len() int {
	return len(s)
}
func (s sortCommitsByDate) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s sortCommitsByDate) Less(i, j int) bool {
	return s[i].Author.When.After(s[j].Author.When)
}
