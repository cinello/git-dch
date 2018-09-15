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
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

func TestLogEntriesToTime(t *testing.T) {
	var (
		location, _ = time.LoadLocation("CET")
		root, _     = os.Getwd()
		wrongWd     = filepath.FromSlash(root + "/pkg/git")

		t20180216221220 = time.Date(2018, 2, 16, 22, 12, 20, 0, location)
		t20000101000000 = time.Date(2000, 1, 1, 0, 0, 0, 0, location)

		cde07e6 = "de07e6b6692d81e6162ae84c58e77b1586309c37"
		cb0dade = "b0dade3fb660b1b078ca0d441b769fb900b3b3f9"
	)

	type args struct {
		t time.Time
	}
	tests := []struct {
		name      string
		args      args
		wd        string
		want      string
		wantError bool
	}{
		{name: `log`, args: args{t: t20180216221220}, wd: root, want: cde07e6},
		{name: `error1`, args: args{t: t20180216221220}, wd: wrongWd, wantError: true},
		{name: `outOfTime`, args: args{t: t20000101000000}, wd: root, want: cb0dade},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Chdir(tt.wd)

			gr, err := NewRepositoryFromCurrentDirectory()

			var gitLog []*object.Commit
			if err == nil {
				gitLog, err = gr.CommitsToTime(tt.args.t, false)
			}

			if !tt.wantError && err != nil {
				t.Errorf("cannot obtain git log: %s", err)
			}

			if tt.wantError {
				if err != nil {
					t.Logf("got expected error: %s", err)
					return
				}
				t.Error("expected an error, got nothing")
			}

			got := gitLog[len(gitLog)-1].Hash.String()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("git log to time := '%v' CommitsToTime(v) = '%v', want '%v'", tt.args, got, tt.want)
			}
		})
	}
	os.Chdir(root)
}

func TestLogEntriesToCommit(t *testing.T) {
	const (
		cde07e6 = "de07e6b6692d81e6162ae84c58e77b1586309c37"
		c123456 = "1234567890123456789012345678901234567890"
		cb0dade = "b0dade3fb660b1b078ca0d441b769fb900b3b3f9"
		cd4ac82 = "d4ac82a99006737d79508a4e753a8b21bfa4f91d"
	)
	var (
		root, _ = os.Getwd()
		wrongWd = filepath.FromSlash(root + "/pkg/git")
	)

	type args struct {
		c string
	}
	tests := []struct {
		name      string
		args      args
		wd        string
		want      string
		wantError bool
	}{
		{name: `log`, args: args{c: cde07e6}, wd: root, want: cde07e6},
		{name: `error1`, args: args{c: cde07e6}, wd: wrongWd, wantError: true},
		{name: `log`, args: args{c: c123456}, wd: root, want: cb0dade},
		{name: `0.0.1`, args: args{c: "0.0.1"}, wd: root, want: cd4ac82},
		{name: `refs/tags/0.0.1`, args: args{c: "refs/tags/0.0.1"}, wd: root, want: cd4ac82},
		{name: `refs/heads/master`, args: args{c: "refs/heads/master"}, wd: root, want: cb0dade},
		{name: `master`, args: args{c: "master"}, wd: root, want: cb0dade},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Chdir(tt.wd)

			gr, err := NewRepositoryFromCurrentDirectory()
			var gitLog []*object.Commit
			if err == nil {
				gitLog, err = gr.CommitsToCommit(tt.args.c, false)
			}

			if !tt.wantError && err != nil {
				t.Errorf("cannot obtain git log: %s", err)
			}

			if tt.wantError {
				if err != nil {
					t.Logf("got expected error: %s", err)
					return
				}
				t.Error("expected an error, got nothing")
			}

			got := gitLog[len(gitLog)-1].Hash.String()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("git log to commit := '%v' CommitsToCommit(v) = '%v', want '%v'", tt.args, got, tt.want)
			}
		})
	}
	os.Chdir(root)
}

func TestLogEntriesToCommitText(t *testing.T) {

}

func TestLogEntryText(t *testing.T) {
	gr, _ := NewRepositoryFromCurrentDirectory()

	var (
		location  = time.Now().Location()
		gitLog, _ = gr.CommitsToTime(time.Date(2018, 2, 17, 0, 0, 0, 0, location), false)
		commit    = gitLog[len(gitLog)-1].Hash.String()
	)

	type args struct {
		c          string
		withAuthor bool
		withHash   bool
		withLine   bool
		full       bool
	}
	tests := []struct {
		name      string
		args      args
		want      string
		wantLines int
		wantError bool
	}{
		{
			name:      `log`,
			args:      args{c: commit, withAuthor: false, withHash: false, withLine: false},
			want:      "Add Build and ExtractNative helpers to version package\n",
			wantLines: 1,
		},
		{
			name:      `logWithAuthor`,
			args:      args{c: commit, withAuthor: true, withHash: false, withLine: false},
			want:      "Add Build and ExtractNative helpers to version package (Yuri Bugelli)\n",
			wantLines: 1,
		},
		{
			name:      `logWithHash`,
			args:      args{c: commit, withAuthor: false, withHash: true, withLine: false},
			want:      "[b900d2d] Add Build and ExtractNative helpers to version package\n",
			wantLines: 1,
		},
		{
			name:      `logWithLine`,
			args:      args{c: commit, withAuthor: false, withHash: false, withLine: true},
			want:      "  * Add Build and ExtractNative helpers to version package\n",
			wantLines: 1,
		},
		{
			name:      `logWithAuthorAndHash`,
			args:      args{c: commit, withAuthor: true, withHash: true, withLine: false},
			want:      "[b900d2d] Add Build and ExtractNative helpers to version package (Yuri Bugelli)\n",
			wantLines: 1,
		},
		{
			name:      `logWithAuthorAndLine`,
			args:      args{c: commit, withAuthor: true, withHash: false, withLine: true},
			want:      "  * Add Build and ExtractNative helpers to version package (Yuri Bugelli)\n",
			wantLines: 1,
		},
		{
			name:      `logWithHashAndLine`,
			args:      args{c: commit, withAuthor: false, withHash: true, withLine: true},
			want:      "  * [b900d2d] Add Build and ExtractNative helpers to version package\n",
			wantLines: 1,
		},
		{
			name:      `logWithAuthorAndHashAndLine`,
			args:      args{c: commit, withAuthor: true, withHash: true, withLine: true},
			want:      "  * [b900d2d] Add Build and ExtractNative helpers to version package (Yuri Bugelli)\n",
			wantLines: 1,
		},
		{
			name: `logWithAuthorAndHashAndLine`,
			args: args{c: commit, withAuthor: true, withHash: true, withLine: true, full: true},
			want: `  * [b900d2d] Add Build and ExtractNative helpers to version package (Yuri Bugelli)

    func Build(v version.Version, t ReleaseType)
      return a new version of type t, given the native version v

    func ExtractNative(v version.Version)
      return the native part of the passed version

`,
			wantLines: 8,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := gr.LogToCommit(tt.args.c, tt.args.withAuthor, tt.args.withHash, tt.args.withLine, tt.args.full, false)

			if !tt.wantError && err != nil {
				t.Errorf("cannot obtain git log: %s", err)
			}

			if tt.wantError {
				if err != nil {
					t.Logf("got expected error: %s", err)
					return
				}
				t.Error("expected an error, got nothing")
			}
			got = strings.Join(strings.Split(got, "\n")[len(strings.Split(got, "\n"))-(tt.wantLines+1):], "\n")

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("log text to commit := '%v' LogToCommit(v) = '%v', want '%v'", tt.args, got, tt.want)
			}
		})
	}
}
