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
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"gitlab.yuribugelli.it/debian/git-dch-go/pkg/dchversion"
	"gitlab.yuribugelli.it/debian/git-dch-go/pkg/git"

	"github.com/gandalfmagic/go-debian/changelog"
	"github.com/gandalfmagic/go-debian/version"
)

func TestNew(t *testing.T) {

	pwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	type args struct {
		file string
	}
	tests := []struct {
		name       string
		args       args
		want       int
		wantError  bool
		fromReader bool
	}{
		{name: `test.1`, args: args{file: "/test/file/changelog.01"}, want: 1},
		{name: `test.2`, args: args{file: "/test/file/changelog.02"}, want: 3, fromReader: true},
		{name: `test.3`, args: args{file: "/test/file/changelog.03"}, want: 1},
		{name: `test.4`, args: args{file: "/test/file/changelog.04"}, want: 1},
		{name: `empty`, args: args{file: "/test/file/changelog.empty"}, want: 0},
		{name: `error.1`, args: args{file: "/test/file/changelog.error.01"}, wantError: true},
		{name: `error.2`, args: args{file: "/test/file/changelog.error.02"}, wantError: true},
		{name: `error.3`, args: args{file: "/test/file/changelog.error.03"}, wantError: true},
		{name: `error.4`, args: args{file: "/test/file/changelog.error.03"}, fromReader: true, wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var items *File
			var err error

			if tt.fromReader {
				f, errFile := os.Open(pwd + filepath.FromSlash(tt.args.file))
				if errFile != nil {
					t.Errorf("NewItemListFromFile(file), cannot create reader from %s\n", tt.args.file)
				}
				defer f.Close()
				items, err = New(f)
			} else {
				items, err = NewFromFile(pwd + filepath.FromSlash(tt.args.file))
			}

			if !tt.wantError && err != nil {
				t.Errorf("cannot read changelog: %s", err)
				return
			}

			if tt.wantError {
				if err == nil {
					t.Error("expected an error, got nothing")
				}
				return
			}

			got := items.Len()
			if got != tt.want {
				t.Errorf("New(%v) = '%v', want '%v'", tt.args.file, got, tt.want)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {

	pwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	type args struct {
		file string
	}
	tests := []struct {
		name       string
		args       args
		want       bool
		fromReader bool
	}{
		{name: `test`, args: args{file: "/test/file/changelog.01"}, want: false},
		{name: `empty`, args: args{file: "/test/file/changelog.empty"}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			f, err := NewFromFile(pwd + filepath.FromSlash(tt.args.file))
			if err != nil {
				t.Errorf("cannot read changelog: %s", err)
			}

			got := f.IsEmpty()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("IsEmpty() = '%v', want '%v'", got, tt.want)
			}
		})
	}
}

func TestLen(t *testing.T) {

	pwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	type args struct {
		file string
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: `test`, args: args{file: "/test/file/changelog.01"}, want: 1},
		{name: `empty`, args: args{file: "/test/file/changelog.empty"}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			f, err := NewFromFile(pwd + filepath.FromSlash(tt.args.file))
			if err != nil {
				t.Errorf("cannot read changelog: %s", err)
			}

			got := f.Len()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Len() = '%v', want '%v'", got, tt.want)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	const (
		changelog01 = `test (0.0.3-1) unstable; urgency=medium

  * Initial release.

 -- Test Author <test.author@nomail.org>  Tue, 14 Mar 2017 17:34:52 +0000
`
	)

	pwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	type args struct {
		file string
	}
	tests := []struct {
		name      string
		args      args
		want      string
		wantError bool
	}{
		{name: `test`, args: args{file: "/test/file/changelog.01"}, want: changelog01},
		{name: `empty`, args: args{file: "/test/file/changelog.empty"}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			f, err := NewFromFile(pwd + filepath.FromSlash(tt.args.file))
			if err != nil {
				t.Errorf("cannot read changelog: %s", err)
			}

			got := bytes.NewBufferString("")
			f.Write(got)
			if !reflect.DeepEqual(got.String(), tt.want) {
				t.Errorf("Write() =\n'%v', want\n'%v'", got, tt.want)
			}
		})
	}
}

func TestComputeNewVersion(t *testing.T) {

	pwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	type args struct {
		file string
		v    dchversion.Version
	}
	tests := []struct {
		name         string
		args         args
		want         dchversion.Version
		wantError    bool
		wantSnapshot bool
	}{
		{
			name: `emptyRelease`,
			args: args{
				file: "/test/file/changelog.empty",
				v:    dchversion.NewVersion(0, "1.0.0", "1"),
			},
			want: dchversion.NewVersion(0, "1.0.0", "1"),
		},
		{
			name: `emptyNative`,
			args: args{
				file: "/test/file/changelog.empty",
				v:    dchversion.NewVersion(0, "1.0.0", ""),
			},
			want: dchversion.NewVersion(0, "1.0.0", "1"),
		},
		{
			name: `emptySnapshot`,
			args: args{
				file: "/test/file/changelog.empty",
				v:    dchversion.NewVersion(3, "0.0.3~1.gbp123456", ""),
			},
			want: dchversion.NewVersion(3, "0.0.3~1.gbp123456", ""),
		},
		{
			name: `emptyStaging`,
			args: args{
				file: "/test/file/changelog.empty",
				v:    dchversion.NewVersion(3, "0.0.3~stg", "1"),
			},
			want: dchversion.NewVersion(3, "0.0.3~stg", "1"),
		},
		{
			name: `emptyDevelopment`,
			args: args{
				file: "/test/file/changelog.empty",
				v:    dchversion.NewVersion(3, "0.0.3.20180101", "1"),
			},
			want: dchversion.NewVersion(3, "0.0.3.20180101", "1"),
		},
		{
			name: `test1`,
			args: args{
				file: "/test/file/changelog.01",
				v:    dchversion.NewVersion(0, "1.0.0", "1"),
			},
			want: dchversion.NewVersion(0, "1.0.0", "1"),
		},
		{
			name: `test1`,
			args: args{
				file: "/test/file/changelog.01",
				v:    dchversion.NewVersion(0, "0.0.3.20180101", "1"),
			},
			want: dchversion.NewVersion(0, "0.0.3.20180101", "1"),
		},
		{
			name: `incrementRevision1`,
			args: args{
				file: "/test/file/changelog.01",
				v:    dchversion.NewVersion(0, "0.0.3", "1"),
			},
			want: dchversion.NewVersion(0, "0.0.3", "2"),
		},
		{
			name: `incrementRevision2`,
			args: args{
				file: "/test/file/changelog.01",
				v:    dchversion.NewVersion(0, "0.0.3", ""),
			},
			want: dchversion.NewVersion(0, "0.0.3", "2"),
		},
		{
			name: `errorLessThan1`,
			args: args{
				file: "/test/file/changelog.01",
				v:    dchversion.NewVersion(0, "0.0.2", "4"),
			},
			want:      dchversion.Version{},
			wantError: true,
		},
		{
			name: `errorLessThan2`,
			args: args{
				file: "/test/file/changelog.01",
				v:    dchversion.NewVersion(0, "0.0.3~stg", "1"),
			},
			want:      dchversion.Version{},
			wantError: true,
		},
		{
			name: `errorLessThan3`,
			args: args{
				file: "/test/file/changelog.01",
				v:    dchversion.NewVersion(0, "0.0.3~1.gbp123456", ""),
			},
			want:      dchversion.Version{},
			wantError: true,
		},
		{
			name: `errorLessThan4`,
			args: args{
				file: "/test/file/changelog.01",
				v:    dchversion.NewVersion(0, "0.0.2", ""),
			},
			want:      dchversion.Version{},
			wantError: true,
		},
		{
			name: `testSnapshotIncrement01`,
			args: args{
				file: "/test/file/changelog.05",
				v:    dchversion.NewVersion(0, "0.0.4~1.gbp123456", ""),
			},
			want:         dchversion.Version{},
			wantSnapshot: true,
		},
		{
			name: `testSnapshotIncrement02`,
			args: args{
				file: "/test/file/changelog.05",
				v:    dchversion.NewVersion(0, "0.0.4~2.gbp123456", ""),
			},
			want:         dchversion.Version{},
			wantSnapshot: true,
		},
		{
			name: `testSnapshotIncrement03`,
			args: args{
				file: "/test/file/changelog.05",
				v:    dchversion.NewVersion(0, "0.0.4~0.gbp123456", ""),
			},
			want:         dchversion.Version{},
			wantSnapshot: true,
		},
		{
			name: `errorWrongSnapshot`,
			args: args{
				file: "/test/file/changelog.05",
				v:    dchversion.NewVersion(0, "0.0.4~2.gbp1234578", ""),
			},
			want:      dchversion.Version{},
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			f, err := NewFromFile(pwd + filepath.FromSlash(tt.args.file))
			if err != nil {
				t.Errorf("cannot read changelog: %s", err)
			}

			got, err := f.computeNewVersion(tt.args.v)

			if !tt.wantError && err != nil {
				t.Errorf("cannot compute new version: %s", err)
				return
			}

			if tt.wantError {
				if err == nil {
					t.Error("expected an error, got nothing")
				}
				return
			}

			if tt.wantSnapshot {
				if !got.IsSnapshot() {
					t.Errorf("expected a snapshot wersion, got something else: %v", got)
				}
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("compute new version := '%v' LogEntriesToTime(%v) = '%v', want '%v'", tt.args, tt.name, got, tt.want)
			}
		})
	}
}

func TestLastVersion(t *testing.T) {
	pwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	type args struct {
		file string
	}
	tests := []struct {
		name      string
		args      args
		want      dchversion.Version
		wantError bool
	}{
		{
			name: `emptyRelease`,
			args: args{
				file: "/test/file/changelog.empty",
			},
			want:      dchversion.Version{},
			wantError: true,
		},
		{
			name: `test01`,
			args: args{
				file: "/test/file/changelog.01",
			},
			want: dchversion.NewVersion(0, "0.0.3", "1"),
		},
		{
			name: `test02`,
			args: args{
				file: "/test/file/changelog.02",
			},
			want: dchversion.NewVersion(0, "0.0.5", "1"),
		},
		{
			name: `test03`,
			args: args{
				file: "/test/file/changelog.05",
			},
			want: dchversion.NewVersion(0, "0.0.4~1.gbp123456", ""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			f, err := NewFromFile(pwd + filepath.FromSlash(tt.args.file))
			if err != nil {
				t.Errorf("cannot read changelog: %s", err)
			}

			got, err := f.LastVersion()

			if !tt.wantError && err != nil {
				t.Errorf("cannot compute new version: %s", err)
				return
			}

			if tt.wantError {
				if err == nil {
					t.Error("expected an error, got nothing")
				}
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("gel last version := '%v' LastVersion(%v) = '%v', want '%v'", tt.args, tt.name, got, tt.want)
			}
		})
	}
}

func TestAddSimple(t *testing.T) {
	const (
		textSingleEntry = `test (0.0.3-1) unstable; urgency=medium

  * Initial release.

 -- Test Author <test.author@nomail.org>  Tue, 14 Mar 2017 17:34:52 +0000
`
	)

	wd, _ := os.Getwd()
	wd = wd[strings.LastIndex(wd, string(filepath.Separator))+1:]

	type args struct {
		r       io.Reader
		source  string
		v       dchversion.Version
		urgency string
		target  string
		changes string
		author  string
	}
	tests := []struct {
		name      string
		args      args
		want      changelog.ChangelogEntry
		wantError bool
	}{
		{
			name: `test`,
			args: args{
				r:       strings.NewReader(textSingleEntry),
				v:       dchversion.NewVersion(0, "1.0.0", "1"),
				target:  "unstable",
				changes: "test args",
				author:  "Test Author <test.author@nomail.org>",
			},
			want: changelog.ChangelogEntry{
				Version:   version.Version{Epoch: 0, Version: "1.0.0", Revision: "1"},
				Source:    "test",
				Target:    "unstable",
				Arguments: map[string]string{"urgency": "medium"},
				Changelog: "\ntest args\n\n",
				ChangedBy: "Test Author <test.author@nomail.org>",
				When:      time.Now(),
			},
		},
		{
			name: `incrementRevision`,
			args: args{
				r:       strings.NewReader(textSingleEntry),
				v:       dchversion.NewVersion(0, "0.0.3", "1"),
				target:  "unstable",
				changes: "test args",
				author:  "Test Author <test.author@nomail.org>",
			},
			want: changelog.ChangelogEntry{
				Version:   version.Version{Epoch: 0, Version: "0.0.3", Revision: "2"},
				Source:    "test",
				Target:    "unstable",
				Arguments: map[string]string{"urgency": "medium"},
				Changelog: "\ntest args\n\n",
				ChangedBy: "Test Author <test.author@nomail.org>",
				When:      time.Now(),
			},
		},
		{
			name: `empty1`,
			args: args{
				r:       strings.NewReader(""),
				v:       dchversion.NewVersion(0, "1.0.0", "1"),
				target:  "unstable",
				changes: "test args",
				author:  "Test Author <test.author@nomail.org>",
			},
			want: changelog.ChangelogEntry{
				Version:   version.Version{Epoch: 0, Version: "1.0.0", Revision: "1"},
				Source:    wd,
				Target:    "unstable",
				Arguments: map[string]string{"urgency": "medium"},
				Changelog: "\ntest args\n\n",
				ChangedBy: "Test Author <test.author@nomail.org>",
				When:      time.Now(),
			},
		},
		{
			name: `empty2`,
			args: args{
				source:  "source",
				r:       strings.NewReader(""),
				v:       dchversion.NewVersion(0, "1.0.0", "1"),
				urgency: "low",
				target:  "unstable",
				changes: "test args",
				author:  "Test Author <test.author@nomail.org>",
			},
			want: changelog.ChangelogEntry{
				Version:   version.Version{Epoch: 0, Version: "1.0.0", Revision: "1"},
				Source:    "source",
				Target:    "unstable",
				Arguments: map[string]string{"urgency": "low"},
				Changelog: "\ntest args\n\n",
				ChangedBy: "Test Author <test.author@nomail.org>",
				When:      time.Now(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := New(tt.args.r)
			if err != nil {
				t.Errorf("cannot read changelog: %s", err)
			}

			_, got, err := file.addSimple(tt.args.source, tt.args.v, tt.args.urgency, tt.args.target, tt.args.changes, tt.args.author)
			if err != nil {
				t.Errorf("cannot add args to changelog: %s", err)
			}
			got.When = tt.want.When
			if !reflect.DeepEqual(got.ChangelogEntry, tt.want) {
				t.Errorf("addSimple(%v) =\n'%v', want\n'%v'", tt, got, tt.want)
			}
		})
	}
}

func TestAddSnapshot(t *testing.T) {
	const (
		textSingleEntry = `test (0.0.3-1) unstable; urgency=medium

  * Initial release.

 -- Test Author <test.author@nomail.org>  Tue, 14 Mar 2017 17:34:52 +0000
`
	)

	wd, _ := os.Getwd()
	wd = wd[strings.LastIndex(wd, string(filepath.Separator))+1:]

	gr, _ := git.NewRepositoryFromCurrentDirectory()
	hash, _ := gr.LastCommitHash(6)

	type args struct {
		r      io.Reader
		source string
		v      dchversion.Version
		author string
	}
	tests := []struct {
		name      string
		args      args
		want      changelog.ChangelogEntry
		wantError bool
	}{
		{
			name: `fromStable`,
			args: args{
				r:      strings.NewReader(textSingleEntry),
				v:      dchversion.NewVersion(0, "1.0.0", "1"),
				author: "Test Author <test.author@nomail.org>",
			},
			wantError: true,
		},
		{
			name: `fromStaging`,
			args: args{
				r:      strings.NewReader(textSingleEntry),
				v:      dchversion.NewVersion(0, "1.0.0~stg", "1"),
				author: "Test Author <test.author@nomail.org>",
			},
			wantError: true,
		},
		{
			name: `fromDevelopment`,
			args: args{
				r:      strings.NewReader(textSingleEntry),
				v:      dchversion.NewVersion(0, "1.0.0.20180101", "1"),
				author: "Test Author <test.author@nomail.org>",
			},
			wantError: true,
		},
		{
			name: `fromNative`,
			args: args{
				r:      strings.NewReader(textSingleEntry),
				v:      dchversion.NewVersion(0, "1.0.0", ""),
				author: "Test Author <test.author@nomail.org>",
			},
			want: changelog.ChangelogEntry{
				Version:   version.Version{Epoch: 0, Version: "1.0.0~1.gbp" + hash, Revision: ""},
				Source:    "test",
				Target:    "UNRELEASED",
				Arguments: map[string]string{"urgency": "low"},
				ChangedBy: "Test Author <test.author@nomail.org>",
				When:      time.Now(),
			},
		},
		{
			name: `fromSnapshot`,
			args: args{
				r:      strings.NewReader(textSingleEntry),
				v:      dchversion.NewVersion(0, "1.0.0~1.gbp"+hash, ""),
				author: "Test Author <test.author@nomail.org>",
			},
			want: changelog.ChangelogEntry{
				Version:   version.Version{Epoch: 0, Version: "1.0.0~1.gbp" + hash, Revision: ""},
				Source:    "test",
				Target:    "UNRELEASED",
				Arguments: map[string]string{"urgency": "low"},
				ChangedBy: "Test Author <test.author@nomail.org>",
				When:      time.Now(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			file, err := New(tt.args.r)
			if err != nil {
				t.Errorf("cannot read changelog: %s", err)
			}
			log, _ := file.buildSnapshotLog("", false, false)

			_, got, err := file.AddSnapshot("", tt.args.source, tt.args.v, tt.args.author, false, false)

			if !tt.wantError && err != nil {
				t.Errorf("cannot add args to changelog: %s", err)
			}

			if tt.wantError {
				if err == nil {
					t.Error("expected an error, got nothing")
				}
				return
			}

			tt.want.Changelog = "\n" + log + "\n"
			got.When = tt.want.When

			if !reflect.DeepEqual(got.ChangelogEntry, tt.want) {
				t.Errorf("AddSnapshot(%v) = '%v', want '%v'", tt.args, got, tt.want)
			}
		})
	}
}

func TestComputeSourceName(t *testing.T) {
	const (
		textSingleEntry = `test (0.0.3-1) unstable; urgency=medium

  * Initial release.

 -- Test Author <test.author@nomail.org>  Tue, 14 Mar 2017 17:34:52 +0000
`
	)

	wd, _ := os.Getwd()
	wd = wd[strings.LastIndex(wd, string(filepath.Separator))+1:]

	type args struct {
		r      io.Reader
		source string
	}
	tests := []struct {
		name      string
		args      args
		want      string
		wantError bool
	}{
		{
			name: `empty`,
			args: args{
				r:      strings.NewReader(""),
				source: "",
			},
			want: wd,
		},
		{
			name: `empty2`,
			args: args{
				r:      strings.NewReader(textSingleEntry),
				source: "",
			},
			want: "test",
		},
		{
			name: `error`,
			args: args{
				r:      strings.NewReader(""),
				source: "wrong value",
			},
			wantError: true,
		},
		{
			name: `test`,
			args: args{
				r:      strings.NewReader(""),
				source: "name",
			},
			want: "name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := New(tt.args.r)
			if err != nil {
				t.Errorf("cannot read changelog: %s", err)
			}

			got, err := f.computeSourceName(tt.args.source)

			if !tt.wantError && err != nil {
				t.Errorf("cannot compute source name: %s", err)
			}

			if tt.wantError {
				if err == nil {
					t.Error("expected an error, got nothing")
				}
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("computeSourceName(%v) = '%v', want '%v'", tt.args, got, tt.want)
			}
		})
	}
}

func TestComputeTargetName(t *testing.T) {
	const (
		textSingleEntry = `test (0.0.3-1) unstable; urgency=medium

  * Initial release.

 -- Test Author <test.author@nomail.org>  Tue, 14 Mar 2017 17:34:52 +0000
`
	)

	type args struct {
		r      io.Reader
		target string
	}
	tests := []struct {
		name      string
		args      args
		want      string
		wantError bool
	}{
		{
			name: `empty`,
			args: args{
				r:      strings.NewReader(""),
				target: "",
			},
			wantError: true,
		},
		{
			name: `empty2`,
			args: args{
				r:      strings.NewReader(textSingleEntry),
				target: "",
			},
			want: "unstable",
		},
		{
			name: `error`,
			args: args{
				r:      strings.NewReader(""),
				target: "wrong value",
			},
			wantError: true,
		},
		{
			name: `test`,
			args: args{
				r:      strings.NewReader(""),
				target: "value",
			},
			want: "value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := New(tt.args.r)
			if err != nil {
				t.Errorf("cannot read changelog: %s", err)
			}

			got, err := f.computeTargetName(tt.args.target)

			if !tt.wantError && err != nil {
				t.Errorf("cannot compute source name: %s", err)
			}

			if tt.wantError {
				if err == nil {
					t.Error("expected an error, got nothing")
				}
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("computeTargetName(%v) = '%v', want '%v'", tt.args, got, tt.want)
			}
		})
	}
}

func TestComputeAuthor(t *testing.T) {
	const (
		textSingleEntry = `test (0.0.3-1) unstable; urgency=medium

  * Initial release.

 -- Test Author <test.author@nomail.org>  Tue, 14 Mar 2017 17:34:52 +0000
`
	)

	type args struct {
		r      io.Reader
		target string
	}
	tests := []struct {
		name      string
		args      args
		want      string
		wantError bool
	}{
		{
			name: `empty`,
			args: args{
				r:      strings.NewReader(""),
				target: "",
			},
			wantError: true,
		},
		{
			name: `empty2`,
			args: args{
				r:      strings.NewReader(textSingleEntry),
				target: "",
			},
			want: "Test Author <test.author@nomail.org>",
		},
		{
			name: `test`,
			args: args{
				r:      strings.NewReader(""),
				target: "test value",
			},
			want: "test value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := New(tt.args.r)
			if err != nil {
				t.Errorf("cannot read changelog: %s", err)
			}

			got, err := f.computeAuthor(tt.args.target)

			if !tt.wantError && err != nil {
				t.Errorf("cannot compute source name: %s", err)
			}

			if tt.wantError {
				if err == nil {
					t.Error("expected an error, got nothing")
				}
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("computeAuthor(%v) = '%v', want '%v'", tt.args, got, tt.want)
			}
		})
	}
}
