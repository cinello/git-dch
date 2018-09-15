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
	"os"

	"gopkg.in/src-d/go-git.v4"
)

type Repository struct {
	repository *git.Repository
}

func NewRepository(path string) (Repository, error) {

	var err error
	var gr *git.Repository

	if gr, err = git.PlainOpen(path); err != nil {
		return Repository{}, err
	}
	return Repository{repository: gr}, nil
}

func NewRepositoryFromCurrentDirectory() (Repository, error) {

	var err error
	var path string

	path, err = os.Getwd()
	if err != nil {
		return Repository{}, fmt.Errorf(textCannotOpenWorkDir, err)
	}

	return NewRepository(path)
}
