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
	"reflect"
	"testing"
)

func TestReleaseTypeSourceBranch(t *testing.T) {
	type args struct {
		t ReleaseType
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: `release`,
			args: args{
				t: Release,
			},
			want: "release",
		},
		{
			name: `staging`,
			args: args{
				t: Staging,
			},
			want: "staging",
		},
		{
			name: `development`,
			args: args{
				t: Development,
			},
			want: "develop",
		},
		{
			name: `snapshot`,
			args: args{
				t: Snapshot,
			},
			want: "develop",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			releaseType := tt.args.t
			got := releaseType.SourceBranch()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("get allowed source branch := '%v' SourceBranch(v) = '%v', want '%v'", tt.args, got, tt.want)
			}
		})
	}
}

func TestReleaseTypeFromBranch(t *testing.T) {
	type args struct {
		b string
	}
	tests := []struct {
		name string
		args args
		want ReleaseType
	}{
		{
			name: `release`,
			args: args{
				b: "release",
			},
			want: Release,
		},
		{
			name: `staging`,
			args: args{
				b: "staging",
			},
			want: Staging,
		},
		{
			name: `development`,
			args: args{
				b: "develop",
			},
			want: Development,
		},
		{
			name: `master`,
			args: args{
				b: "master",
			},
			want: Release,
		},
		{
			name: `feature`,
			args: args{
				b: "feature/add_feature",
			},
			want: Development,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ReleaseTypeFromBranch(tt.args.b)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("get release from branch := '%v' ReleaseTypeFromBranch(v) = '%v', want '%v'", tt.args, got, tt.want)
			}
		})
	}
}
