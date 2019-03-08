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
	"sort"
	"time"

	"github.com/cinello/go-debian/version"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type searchCommitsData struct {
	commits         []*object.Commit
	head            *plumbing.Reference
	conditionBefore func(c *object.Commit) bool
	conditionAfter  func(c *object.Commit) bool
	ignoreMerges    bool
}

func searchCommitsToCondition(data searchCommitsData) (commits []*object.Commit) {

	var headFound bool
	for _, c := range data.commits {
		if data.conditionBefore != nil && data.conditionBefore(c) {
			break
		}

		if c.Hash == data.head.Hash() {
			headFound = true
		}

		// Ignore merge commits
		if data.ignoreMerges && len(c.ParentHashes) > 1 {
			if data.conditionAfter != nil && data.conditionAfter(c) {
				break
			}
			continue
		}

		if headFound {
			commits = append(commits, c)
		}

		if data.conditionAfter != nil && data.conditionAfter(c) {
			break
		}
	}

	return
}

func (gr *Repository) sortedCommits() (commits []*object.Commit, head *plumbing.Reference, err error) {

	var (
		i object.CommitIter
	)

	// Obtain commits iterator
	if i, err = gr.repository.Log(&git.LogOptions{}); err != nil {
		return
	}
	defer i.Close()

	// Obtain head reference
	if head, err = gr.repository.Head(); err != nil {
		return
	}

	// Iterate through the commits to create the unsorted list
	i.ForEach(func(commit *object.Commit) error {
		commits = append(commits, commit)
		return nil
	})

	// Sort the list
	sort.Sort(sortCommitsByDate(commits))

	return
}

func (gr *Repository) CommitsToTime(t time.Time, ignoreMerges bool) (commits []*object.Commit, err error) {

	var (
		data searchCommitsData
	)

	data.commits, data.head, err = gr.sortedCommits()
	data.conditionBefore = func(c *object.Commit) bool {
		return c.Author.When.Before(t)
	}
	data.ignoreMerges = ignoreMerges
	if err == nil {
		commits = searchCommitsToCondition(data)
	}

	return
}

func (gr *Repository) CommitsToCommit(commit string, ignoreMerges bool) (commits []*object.Commit, err error) {

	// Find the commit hash for the passed reference
	var commitFromReference string
	commitFromReference = gr.CommitAtTag(version.Version{Version: commit})
	if commitFromReference == "" {
		commitFromReference = gr.CommitAtTagObject(version.Version{Version: commit})
	}
	if commitFromReference == "" {
		commitFromReference = gr.CommitAtReference(commit)
	}
	// If a commit was not found the passed value was a commit hash itself or an empty value
	if commitFromReference != "" {
		commit = commitFromReference
	}

	var (
		data searchCommitsData
	)

	data.commits, data.head, err = gr.sortedCommits()
	data.conditionAfter = func(c *object.Commit) bool {
		return c.Hash.String() == commit
	}
	data.ignoreMerges = ignoreMerges

	if err == nil {
		commits = searchCommitsToCondition(data)
	}

	return
}
