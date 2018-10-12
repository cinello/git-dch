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
	"testing"

	"gitlab.yuribugelli.it/debian/go-debian/version"
)

func TestLastCommitHash(t *testing.T) {
	var (
		root, _ = os.Getwd()
		wrongWd = filepath.FromSlash(root + "/pkg/git")
	)

	type args struct {
		l int
	}
	tests := []struct {
		name      string
		args      args
		wd        string
		want      int
		wantError bool
	}{
		{name: `full`, args: args{l: -1}, wd: root, want: 40},
		{name: `six`, args: args{l: 6}, wd: root, want: 6},
		{name: `error1`, args: args{l: 6}, wd: wrongWd, wantError: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Chdir(tt.wd)

			gr, err := NewRepositoryFromCurrentDirectory()
			var hash string
			if err == nil {
				hash, err = gr.LastCommitHash(tt.args.l)
			}

			if !tt.wantError && err != nil {
				t.Errorf("cannot obtain commit hash: %s", err)
			}

			if tt.wantError {
				if err != nil {
					t.Logf("got expected error: %s", err)
					return
				}
				t.Error("expected an error, got nothing")
			}

			got := len(hash)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("last commit hash := '%v' LogEntriesToTime(v) = '%v' (%v), want '%v'", tt.args, got, hash, tt.want)
			}
		})
	}
	os.Chdir(root)
}

func TestCommitAtTag(t *testing.T) {
	type args struct {
		tag string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: `empty`, args: args{""}, want: ""},
		{name: `wrong`, args: args{"wrong/0.0.0"}, want: ""},
		{name: `0.0.0`, args: args{"0.0.0"}, want: "b0dade3fb660b1b078ca0d441b769fb900b3b3f9"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gr, _ := NewRepositoryFromCurrentDirectory()
			got := gr.CommitAtTag(version.Version{Epoch: 0, Version: tt.args.tag, Revision: ""})

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("get commit at tag := '%v' CommitAtTag(v) = '%v', want '%v'", tt.args, got, tt.want)
			}
		})
	}
}

func TestCommitAtReference(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: `empty`, args: args{""}, want: ""},
		{name: `wrong`, args: args{"wrong/0.0.0"}, want: ""},
		{name: `0.0.0`, args: args{"0.0.0"}, want: "b0dade3fb660b1b078ca0d441b769fb900b3b3f9"},
		{name: `0.0.1`, args: args{"0.0.1"}, want: "d4ac82a99006737d79508a4e753a8b21bfa4f91d"},
		{name: `master`, args: args{"master"}, want: "b0dade3fb660b1b078ca0d441b769fb900b3b3f9"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gr, _ := NewRepositoryFromCurrentDirectory()
			got := gr.CommitAtReference(tt.args.name)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("get commit at reference := '%v' CommitAtReference(v) = '%v', want '%v'", tt.args, got, tt.want)
			}
		})
	}
}

func TestCommitAtTagObject(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: `empty`, args: args{""}, want: ""},
		{name: `wrong`, args: args{"wrong/0.0.0"}, want: ""},
		{name: `0.0.0`, args: args{"0.0.0"}, want: ""},
		{name: `0.0.1`, args: args{"0.0.1"}, want: ""},
		{name: `annotated`, args: args{"annotated"}, want: "de07e6b6692d81e6162ae84c58e77b1586309c37"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gr, _ := NewRepositoryFromCurrentDirectory()
			got := gr.CommitAtTagObject(version.Version{Epoch: 0, Version: tt.args.name, Revision: ""})

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("get commit at tag object := '%v' CommitAtTagObject(v) = '%v', want '%v'", tt.args, got, tt.want)
			}
		})
	}
}
