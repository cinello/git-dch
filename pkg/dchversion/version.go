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

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/cinello/git-dch/pkg/git"

	"github.com/cinello/go-debian/version"
)

var (
	regExStaging     = regexp.MustCompile(`^(?:\d+:)?\d+\.\d+\.\d+~stg-\d+$`)
	regExDevelopment = regexp.MustCompile(`^(?:\d+:)?\d+\.\d+\.\d+.\d{8}-\d+$`)
	regExSnapshot    = regexp.MustCompile(`^(?:\d+:)?\d+\.\d+\.\d+~\d+\.gbp[0-9a-f]{6}$`)

	regExSplitRevision           = regexp.MustCompile(`(.*?)?(\d+)`)
	regExSplitStagingVersion     = regexp.MustCompile(`^(.*)~stg$`)
	regExSplitDevelopmentVersion = regexp.MustCompile(`^(.*)\.(\d{8})$`)
	regExSplitSnapshotVersion    = regexp.MustCompile(`(.*?)~(\d+)\.gbp([0-9a-f]{6})`)
)

func developmentDate() string {
	return time.Now().Format("20060102")
}

type Version struct {
	v version.Version
}

func NewVersion(epoch uint, v string, revision string) Version {
	return Version{v: version.Version{Epoch: epoch, Version: v, Revision: revision}}
}

func NewVersionFromDebian(v version.Version) Version {
	return Version{v: v}
}

func Parse(input string) (Version, error) {

	value, err := version.Parse(input)
	return Version{v: value}, err
}

func MustParse(s string) Version {
	v, err := version.Parse(s)
	if err != nil {
		panic(`dchversion: Parse(` + s + `): ` + err.Error())
	}

	return Version{v: v}
}

func (v Version) Epoch() uint {
	return v.v.Epoch
}

func (v *Version) SetEpoch(epoch uint) {
	v.v.Epoch = epoch
}

func (v Version) Version() string {
	return v.v.Version
}

func (v *Version) SetVersion(version string) {
	v.v.Version = version
}

func (v Version) Revision() string {
	return v.v.Revision
}

func (v *Version) SetRevision(revision string) {
	v.v.Revision = revision
}

func (v Version) IsNative() bool {
	return v.v.IsNative()
}

func (v *Version) UnmarshalControl(data string) error {
	return v.v.UnmarshalControl(data)
}

func (v Version) MarshalControl() (string, error) {
	return v.v.MarshalControl()
}

func (v Version) String() string {
	return v.v.String()
}

func (v Version) IsSnapshot() bool {
	return regExSnapshot.MatchString(v.v.String())
}

func (v Version) IsStaging() bool {
	return regExStaging.MatchString(v.v.String())
}

func (v Version) IsDevelopment() bool {
	return regExDevelopment.MatchString(v.v.String())
}

func (v Version) IsStable() bool {
	return !v.IsStaging() && !v.IsDevelopment() && !v.IsSnapshot()
}

func (v Version) Type() ReleaseType {
	if v.IsSnapshot() {
		return Snapshot
	}
	if v.IsDevelopment() {
		return Development
	}
	if v.IsStaging() {
		return Staging
	}

	return Release
}

func (v Version) ExtractNative() (out Version) {
	out = NewVersion(v.v.Epoch, v.v.Version, v.v.Revision)
	out.v.Epoch = 0

	switch {
	case out.IsStaging():
		old := regExSplitStagingVersion.FindAllStringSubmatch(out.v.Version, -1)
		out.v.Version = old[0][1]
		out.v.Revision = ""
	case out.IsDevelopment():
		old := regExSplitDevelopmentVersion.FindAllStringSubmatch(out.v.Version, -1)
		out.v.Version = old[0][1]
		out.v.Revision = ""
	case out.IsSnapshot():
		old := regExSplitSnapshotVersion.FindAllStringSubmatch(out.v.Version, -1)
		out.v.Version = old[0][1]
		out.v.Revision = ""
	default:
		out.v.Revision = ""
	}
	return
}

// Build return a new dchversion of type t, given the native dchversion v
func (v Version) Build(t ReleaseType) (out Version, err error) {
	out = v

	switch t {
	case Release:
		switch {
		case out.IsSnapshot():
			fallthrough
		case out.IsStaging():
			e := out.v.Epoch
			out = out.ExtractNative()
			out.v.Epoch = e
			out.v.Revision = "1"
		case out.IsNative():
			out.v.Revision = "1"
		case out.IsStable():
		default:
			err = fmt.Errorf("cannot build stable dchversion from %s", out.String())
			return
		}
	case Staging:
		switch {
		case out.IsSnapshot():
			e := out.v.Epoch
			out = out.ExtractNative()
			out.v.Version += "~stg"
			out.v.Epoch = e
			out.v.Revision = "1"
		case out.IsStaging():
		case out.IsNative():
			out.v.Version += "~stg"
			out.v.Revision = "1"
		default:
			err = fmt.Errorf("cannot build staging dchversion from %s", out.String())
			return
		}
	case Development:
		switch {
		case out.IsDevelopment():
		case out.IsStable():
			out.v.Version += "." + developmentDate()
			out.v.Revision = "1"
		default:
			err = fmt.Errorf("cannot build development dchversion from %s", out.String())
			return
		}
	case Snapshot:
		var gr git.Repository
		if gr, err = git.NewRepositoryFromCurrentDirectory(); err != nil {
			return
		}
		var hash string
		if hash, err = gr.LastCommitHash(6); err != nil {
			return
		}
		switch {
		case out.IsSnapshot():
			old := regExSplitSnapshotVersion.FindAllStringSubmatch(out.v.Version, -1)
			out.v.Version = old[0][1] + "~" + old[0][2] + ".gbp" + hash
			out.v.Revision = ""
		case out.IsNative():
			out.v.Version += "~1.gbp" + hash
			out.v.Revision = ""
		default:
			err = fmt.Errorf("cannot build snapshot dchversion from %s", out.String())
			return
		}
	}
	return
}

func (v Version) IncrementRevision() (newVersion Version, err error) {

	// Epoch is left untouched
	newVersion.v.Epoch = v.v.Epoch

	var revision int64
	if v.Type() == Snapshot {

		// Split dchversion from snapshot revision
		values := regExSplitSnapshotVersion.FindAllStringSubmatch(v.v.Version, -1)
		revision, err = strconv.ParseInt(values[0][2], 10, 64)
		if err != nil {
			err = fmt.Errorf("cannot get the revision from the value %s: %s", v.v.String(), err)
			return
		}

		// Get the last commit hash from local git repository
		var gr git.Repository
		if gr, err = git.NewRepositoryFromCurrentDirectory(); err != nil {
			return
		}
		var hash string
		if hash, err = gr.LastCommitHash(6); err != nil {
			return
		}
		if err != nil {
			err = fmt.Errorf("cannot get the hash from last git commit %s", err)
			return
		}

		// Build the new dchversion
		newVersion.v.Version = values[0][1] + "~" + strconv.FormatInt(revision+1, 10) + ".gbp" + hash

		// Revision is left untouched
		newVersion.v.Revision = v.v.Revision
	} else {
		// Split revision number from any text before
		values := regExSplitRevision.FindAllStringSubmatch(v.v.Revision, -1)
		if len(values) != 1 || len(values[0]) != 3 {
			err = fmt.Errorf("cannot find valid revision number in %s", v.v.String())
			return
		}
		revision, err = strconv.ParseInt(values[0][2], 10, 64)
		if err != nil {
			err = fmt.Errorf("cannot get the revision from the value %s: %s", v.v.String(), err)
			return
		}

		newVersion.v.Version = v.v.Version
		if v.Type() == Development {
			date := time.Now().Format("20060102")
			oldDate := regExSplitDevelopmentVersion.FindAllStringSubmatch(v.v.Version, -1)
			if oldDate[0][2] == date {
				revision++
			} else {
				newVersion.v.Version = oldDate[0][1] + "." + date
				revision = 1
			}
		} else {
			revision++
		}

		// Build new revision value
		newVersion.v.Revision = values[0][1] + strconv.FormatInt(revision, 10)
	}

	return
}

func Compare(a Version, b Version) int {
	return version.Compare(a.v, b.v)
}

// CompareVersions compares the two provided Debian versions. It returns
// 0 if a and b are equal, a value < 0 if a is smaller than b and a
// value > 0 if a is greater than b.
func CompareStrings(a string, b string) (int, error) {
	var err error
	va, err := version.Parse(a)
	if err != nil {
		return 0, fmt.Errorf("value %s is not a valid dchversion", a)
	}
	vb, err := version.Parse(b)
	if err != nil {
		return 0, fmt.Errorf("value %s is not a valid dchversion", b)
	}

	return version.Compare(va, vb), nil
}

func GetSnapshotRelease(v Version) (r int, err error) {
	s := regExSplitSnapshotVersion.FindAllStringSubmatch(v.v.Version, -1)
	if len(s) != 1 && len(s[0]) < 4 {
		return r, fmt.Errorf("the dchversion %s is not a valid snapshot", v.v.String())
	}
	return strconv.Atoi(s[0][2])
}

func SetSnapshotRelease(old Version, r int) (v Version, err error) {
	s := regExSplitSnapshotVersion.FindAllStringSubmatch(old.v.Version, -1)
	if len(s) != 1 && len(s[0]) < 4 {
		return v, fmt.Errorf("the dchversion %s is not a valid snapshot", v.v.String())
	}

	v.v.Version = s[0][1] + "~" + strconv.Itoa(r) + ".gbp" + s[0][3]
	return
}

func CompareSnapshots(a Version, b Version) (r int, err error) {
	sOld := regExSplitSnapshotVersion.FindAllStringSubmatch(a.v.Version, -1)
	if len(sOld) != 1 && len(sOld[0]) < 4 {
		return r, fmt.Errorf("the dchversion %s is not a valid snapshot", a.v.String())
	}
	sNew := regExSplitSnapshotVersion.FindAllStringSubmatch(b.v.Version, -1)
	if len(sNew) != 1 && len(sNew[0]) < 4 {
		return r, fmt.Errorf("the dchversion %s is not a valid snapshot", b.v.String())
	}

	a.v.Version = sOld[0][1] + "~" + sOld[0][2] + ".gbp"
	b.v.Version = sNew[0][1] + "~" + sNew[0][2] + ".gbp"

	return version.Compare(a.v, b.v), nil
}
