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

package dchversion

type ReleaseType int

const (
	Release ReleaseType = iota
	Staging
	Development
	Snapshot
)

func (t ReleaseType) SourceBranch() string {
	switch t {
	case Release:
		return "release"
	case Staging:
		return "staging"
	}
	return "develop"
}

func ReleaseTypeFromBranch(release string) ReleaseType {
	switch release {
	case "master":
		return Release
	case "release":
		return Release
	case "staging":
		return Staging
	}
	return Development
}
