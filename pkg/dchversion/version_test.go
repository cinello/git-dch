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
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"gitlab.yuribugelli.it/debian/git-dch-go/pkg/git"
)

func TestMustParse(t *testing.T) {
	type args struct {
		v string
	}
	tests := []struct {
		name      string
		args      args
		want      Version
		wantPanic bool
	}{
		{
			name: `stable`,
			args: args{
				v: "1.0.0-1",
			},
			want: NewVersion(0, "1.0.0", "1"),
		},
		{
			name: `staging`,
			args: args{
				v: "1.0.0~stg-2",
			},
			want: NewVersion(0, "1.0.0~stg", "2"),
		},
		{
			name: `development`,
			args: args{
				v: "1.0.0.20180101-1",
			},
			want: NewVersion(0, "1.0.0.20180101", "1"),
		},
		{
			name: `snapshot`,
			args: args{
				v: "1.0.0~1.gbp123456",
			},
			want: NewVersion(0, "1.0.0~1.gbp123456", ""),
		},
		{
			name: `panic`,
			args: args{
				v: "abcd",
			},
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()

				if !tt.wantPanic && r != nil {
					t.Errorf("got unexpected panic: %s", r)
				}

				if tt.wantPanic {
					if r != nil {
						t.Logf("got expected panic: %s", r)
						return
					}
					t.Error("expected panic, got nothing")
				}
			}()

			got := MustParse(tt.args.v)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parse dchversion string := '%v' MustParse(v) = '%v', want '%v'", tt.args, got, tt.want)
			}
		})
	}
}

func TestBuild(t *testing.T) {
	root, _ := os.Getwd()
	wrongWd := filepath.FromSlash(root + "/pkg/dchversion")

	gr, _ := git.NewRepositoryFromCurrentDirectory()
	hash, _ := gr.LastCommitHash(6)

	type args struct {
		v Version
		t ReleaseType
	}
	tests := []struct {
		name      string
		args      args
		wd        string
		want      Version
		wantError bool
	}{
		{
			name: `nativeToStable`,
			args: args{v: NewVersion(0, "1.0.0", ""), t: Release},
			wd:   root,
			want: NewVersion(0, "1.0.0", "1"),
		},
		{
			name: `stableToStable`,
			args: args{v: NewVersion(2, "1.0.0", "1"), t: Release},
			wd:   root,
			want: NewVersion(2, "1.0.0", "1"),
		},
		{
			name: `stagingToStable`,
			args: args{v: NewVersion(4, "1.0.0~stg", "4"), t: Release},
			wd:   root,
			want: NewVersion(4, "1.0.0", "1"),
		},
		{
			name:      `developmentToStable`,
			args:      args{v: NewVersion(3, "1.0.0.20180101", "4"), t: Release},
			wd:        root,
			wantError: true,
		},
		{
			name: `snapshotToStable`,
			args: args{v: NewVersion(3, "1.0.0~1.gbp123456", ""), t: Release},
			wd:   root,
			want: NewVersion(3, "1.0.0", "1"),
		},
		{
			name: `nativeToStaging1`,
			args: args{v: NewVersion(0, "1.0.0", ""), t: Staging},
			wd:   root,
			want: NewVersion(0, "1.0.0~stg", "1"),
		},
		{
			name: `nativeToStaging2`,
			args: args{v: NewVersion(3, "1.0.0", ""), t: Staging},
			wd:   root,
			want: NewVersion(3, "1.0.0~stg", "1"),
		},
		{
			name:      `stableToStaging`,
			args:      args{v: NewVersion(0, "1.0.0", "1"), t: Staging},
			wd:        root,
			wantError: true,
		},

		{
			name: `stagingToStaging`,
			args: args{v: NewVersion(3, "1.0.0~stg", "1"), t: Staging},
			wd:   root,
			want: NewVersion(3, "1.0.0~stg", "1"),
		},
		{
			name:      `developmentToStaging`,
			args:      args{v: NewVersion(0, "1.0.0.20180101", "1"), t: Staging},
			wd:        root,
			wantError: true,
		},
		{
			name: `snapshotToStaging`,
			args: args{v: NewVersion(3, "1.0.0~1.gbp123456", ""), t: Staging},
			wd:   root,
			want: NewVersion(3, "1.0.0~stg", "1"),
		},
		{
			name: `nativeToDevelopment`,
			args: args{v: NewVersion(3, "1.0.0", ""), t: Development},
			wd:   root,
			want: NewVersion(3, "1.0.0."+developmentDate(), "1"),
		},
		{
			name: `stableToDevelopment1`,
			args: args{v: NewVersion(3, "1.0.0", "1"), t: Development},
			wd:   root,
			want: NewVersion(3, "1.0.0."+developmentDate(), "1"),
		},
		{
			name: `stableToDevelopment2`,
			args: args{v: NewVersion(3, "1.0.0", "34"), t: Development},
			wd:   root,
			want: NewVersion(3, "1.0.0."+developmentDate(), "1"),
		},
		{
			name:      `stagingToDevelopment`,
			args:      args{v: NewVersion(3, "1.0.0~stg", "22"), t: Development},
			wd:        root,
			wantError: true,
		},
		{
			name: `developmentToDevelopment1`,
			args: args{v: NewVersion(3, "1.0.0.20180101", "33"), t: Development},
			wd:   root,
			want: NewVersion(3, "1.0.0.20180101", "33"),
		},
		{
			name: `developmentToDevelopment2`,
			args: args{v: NewVersion(3, "1.0.0."+developmentDate(), "33"), t: Development},
			wd:   root,
			want: NewVersion(3, "1.0.0."+developmentDate(), "33"),
		},
		{
			name:      `snapshotToDevelopment`,
			args:      args{v: NewVersion(3, "1.0.0~1.gbp123456", ""), t: Development},
			wd:        root,
			wantError: true,
		},
		{
			name: `nativeToSnapshot`,
			args: args{v: NewVersion(3, "1.0.0", ""), t: Snapshot},
			wd:   root,
			want: NewVersion(3, "1.0.0~1.gbp"+hash, ""),
		},
		{
			name:      `stableToSnapshot`,
			args:      args{v: NewVersion(3, "1.0.0", "1"), t: Snapshot},
			wd:        root,
			wantError: true,
		},
		{
			name:      `stagingToSnapshot`,
			args:      args{v: NewVersion(3, "1.0.0~stg", "1"), t: Snapshot},
			wd:        root,
			wantError: true,
		},
		{
			name:      `developmentToSnapshot`,
			args:      args{v: NewVersion(3, "1.0.0.20180101", "1"), t: Snapshot},
			wd:        root,
			wantError: true,
		},
		{
			name: `snapshotToSnapshot`,
			args: args{v: NewVersion(3, "1.0.0~35.gbp123456", ""), t: Snapshot},
			wd:   root,
			want: NewVersion(3, "1.0.0~35.gbp"+hash, ""),
		},
		{
			name:      `errorHash`,
			args:      args{v: NewVersion(3, "1.0.0~35.gbp123456", ""), t: Snapshot},
			wd:        wrongWd,
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Chdir(tt.wd)

			got, err := tt.args.v.Build(tt.args.t)

			if !tt.wantError && err != nil {
				t.Errorf("cannot build dchversion: %s", err)
			}

			if tt.wantError {
				if err != nil {
					t.Logf("got expected error: %s", err)
					return
				}
				t.Error("expected an error, got nothing")
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("build new dchversion := '%v' MustParse(v) = '%v', want '%v'", tt.args, got, tt.want)
			}
		})
	}
	os.Chdir(root)
}

func TestExtractNative(t *testing.T) {
	type args struct {
		v Version
	}
	tests := []struct {
		name string
		args args
		want Version
	}{
		{
			name: `native`,
			args: args{
				v: NewVersion(0, "1.0.0", ""),
			},
			want: NewVersion(0, "1.0.0", ""),
		},
		{
			name: `stable`,
			args: args{
				v: NewVersion(2, "1.0.0", "1"),
			},
			want: NewVersion(0, "1.0.0", ""),
		},
		{
			name: `staging`,
			args: args{
				v: NewVersion(4, "1.0.0~stg", "4"),
			},
			want: NewVersion(0, "1.0.0", ""),
		},
		{
			name: `development`,
			args: args{
				v: NewVersion(3, "1.0.0.20180101", "4"),
			},
			want: NewVersion(0, "1.0.0", ""),
		},
		{
			name: `snapshot`,
			args: args{
				v: NewVersion(3, "1.0.0~1.gbp123456", ""),
			},
			want: NewVersion(0, "1.0.0", ""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.v.ExtractNative()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extract native dchversion := '%v' MustParse(v) = '%v', want '%v'", tt.args, got, tt.want)
			}
		})
	}
}

func TestCompareVersionStrings(t *testing.T) {
	type args struct {
		a string
		b string
	}
	tests := []struct {
		name      string
		args      args
		want      int
		wantError bool
	}{
		{
			name: `error1`,
			args: args{
				a: "abc",
				b: "1.0.0-1",
			},
			wantError: true,
		},
		{
			name: `error2`,
			args: args{
				a: "1.0.0-1",
				b: "def",
			},
			wantError: true,
		},
		{
			name: `equalStable`,
			args: args{
				a: "1.0.0-1",
				b: "1.0.0-1",
			},
			want: 0,
		},
		{
			name: `equalStaging`,
			args: args{
				a: "1.0.0~stg-1",
				b: "1.0.0~stg-1",
			},
			want: 0,
		},
		{
			name: `equalDevelopment`,
			args: args{
				a: "1.0.0.20180101-1",
				b: "1.0.0.20180101-1",
			},
			want: 0,
		},
		{
			name: `equalSnapshot`,
			args: args{
				a: "1.0.0~1.gbp123456",
				b: "1.0.0~1.gbp123456",
			},
			want: 0,
		},
		{
			name: `snapshotLessThanStaging`,
			args: args{
				a: "1.0.0~1.gbp123456",
				b: "1.0.0~stg-1",
			},
			want: -115,
		},
		{
			name: `stagingLessThanStable`,
			args: args{
				a: "1.0.0~stg-1",
				b: "1.0.0-1",
			},
			want: -1,
		},
		{
			name: `snapshotLessThanStable`,
			args: args{
				a: "1.0.0~1.gbp123456",
				b: "1.0.0-1",
			},
			want: -1,
		},
		{
			name: `stableLessThanDevelopment`,
			args: args{
				a: "1.0.0-1",
				b: "1.0.0.20180101-1",
			},
			want: -302,
		},
		{
			name: `stableLessThanStable`,
			args: args{
				a: "1.0.0-1",
				b: "1.0.0-2",
			},
			want: -1,
		},
		{
			name: `stagingLessThanStaging`,
			args: args{
				a: "1.0.0~stg-1",
				b: "1.0.0~stg-2",
			},
			want: -1,
		},
		{
			name: `developmentLessThanDevelopment`,
			args: args{
				a: "1.0.0.20180101-1",
				b: "1.0.0.20180101-2",
			},
			want: -1,
		},
		{
			name: `snapshotLessThanSnapshot`,
			args: args{
				a: "1.0.0~1.gbp123456",
				b: "1.0.0~2.gbp123456",
			},
			want: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompareStrings(tt.args.a, tt.args.b)

			if !tt.wantError && err != nil {
				t.Errorf("cannot compare versions: %s", err)
			}

			if tt.wantError {
				if err != nil {
					t.Logf("got expected error: %s", err)
					return
				}
				t.Error("expected an error, got nothing")
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("compare dchversion := '%v' CompareVersionStrings(a, b) = '%v', want '%v'", tt.args, got, tt.want)
			}
		})
	}
}

func TestIsSnapshot(t *testing.T) {
	type args struct {
		v Version
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: `stable`,
			args: args{
				v: MustParse("1.0.0-1"),
			},
			want: false,
		},
		{
			name: `staging`,
			args: args{
				v: MustParse("1.0.0~stg-2"),
			},
			want: false,
		},
		{
			name: `development`,
			args: args{
				v: MustParse("1.0.0.20180101-1"),
			},
			want: false,
		},
		{
			name: `snapshot`,
			args: args{
				v: MustParse("1.0.0~1.gbp123456"),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.v.IsSnapshot()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("test snapshot := '%v' IsSnapshot(v) = '%v', want '%v'", tt.args, got, tt.want)
			}
		})
	}
}

func TestIsStaging(t *testing.T) {
	type args struct {
		v Version
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: `stable`,
			args: args{
				v: MustParse("1.0.0-1"),
			},
			want: false,
		},
		{
			name: `staging`,
			args: args{
				v: MustParse("1.0.0~stg-2"),
			},
			want: true,
		},
		{
			name: `development`,
			args: args{
				v: MustParse("1.0.0.20180101-1"),
			},
			want: false,
		},
		{
			name: `snapshot`,
			args: args{
				v: MustParse("1.0.0~1.gbp123456"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.v.IsStaging()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("test staging := '%v' IsStaging(v) = '%v', want '%v'", tt.args, got, tt.want)
			}
		})
	}
}

func TestIsDevelopment(t *testing.T) {
	type args struct {
		v Version
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: `stable`,
			args: args{
				v: MustParse("1.0.0-1"),
			},
			want: false,
		},
		{
			name: `staging`,
			args: args{
				v: MustParse("1.0.0~stg-2"),
			},
			want: false,
		},
		{
			name: `development`,
			args: args{
				v: MustParse("1.0.0.20180101-1"),
			},
			want: true,
		},
		{
			name: `snapshot`,
			args: args{
				v: MustParse("1.0.0~1.gbp123456"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.v.IsDevelopment()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("test development := '%v' IsDevelopment(v) = '%v', want '%v'", tt.args, got, tt.want)
			}
		})
	}
}

func TestIsStable(t *testing.T) {
	type args struct {
		v Version
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: `stable`,
			args: args{
				v: MustParse("1.0.0-1"),
			},
			want: true,
		},
		{
			name: `staging`,
			args: args{
				v: MustParse("1.0.0~stg-2"),
			},
			want: false,
		},
		{
			name: `development`,
			args: args{
				v: MustParse("1.0.0.20180101-1"),
			},
			want: false,
		},
		{
			name: `snapshot`,
			args: args{
				v: MustParse("1.0.0~1.gbp123456"),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.v.IsStable()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("test stable := '%v' IsStable(v) = '%v', want '%v'", tt.args, got, tt.want)
			}
		})
	}
}

func TestIncrementRevision(t *testing.T) {
	gr, _ := git.NewRepositoryFromCurrentDirectory()
	hash, _ := gr.LastCommitHash(6)
	root, _ := os.Getwd()

	type args struct {
		v Version
	}
	tests := []struct {
		name      string
		args      args
		wd        string
		want      Version
		wantError bool
	}{
		{
			name: `stable`,
			args: args{
				v: MustParse("1.0.0-1"),
			},
			wd:   root,
			want: MustParse("1.0.0-2"),
		},
		{
			name: `staging`,
			args: args{
				v: MustParse("1.0.0~stg-1"),
			},
			wd:   root,
			want: MustParse("1.0.0~stg-2"),
		},
		{
			name: `development1`,
			args: args{
				v: MustParse("1.0.0.20180101-1"),
			},
			wd:   root,
			want: MustParse("1.0.0." + developmentDate() + "-1"),
		},
		{
			name: `development2`,
			args: args{
				v: MustParse("1.0.0." + developmentDate() + "-1"),
			},
			wd:   root,
			want: MustParse("1.0.0." + developmentDate() + "-2"),
		},
		{
			name: `snapshot`,
			args: args{
				v: MustParse("1.0.0~1.gbp123456"),
			},
			wd:   root,
			want: MustParse("1.0.0~2.gbp" + hash),
		},
		{
			name: `custom1`,
			args: args{
				v: MustParse("1.0.0-test1"),
			},
			wd:   root,
			want: MustParse("1.0.0-test2"),
		},
		{
			name: `errorStable1`,
			args: args{
				v: MustParse("1.0.0-test"),
			},
			wd:        root,
			wantError: true,
		},
		{
			name: `errorStable2`,
			args: args{
				v: MustParse("1.0.0-9999999999999999999"),
			},
			wd:        root,
			wantError: true,
		},
		{
			name: `errorSnapshot1`,
			args: args{
				v: MustParse("1.0.0~9999999999999999999.gbp123456"),
			},
			wd:        root,
			wantError: true,
		},
		{
			name: `errorSnapshot2`,
			args: args{
				v: MustParse("1.0.0~1xc.gbp123456"),
			},
			wd:        root,
			wantError: true,
		},
		{
			name: `errorSnapshot3`,
			args: args{
				v: MustParse("1.0.0~1.gbp123456"),
			},
			wd:        filepath.FromSlash(root + "/pkg/dchversion"),
			wantError: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.Chdir(tt.wd); err != nil {
				t.Errorf("cannot change working directory to %s: %s", tt.wd, err)
			}

			out, err := tt.args.v.IncrementRevision()

			if !tt.wantError && err != nil {
				t.Errorf("cannot increment revision: %s", err)
			}

			if tt.wantError {
				if err != nil {
					t.Logf("got expected error: %s", err)
					return
				}
				t.Error("expected an error, got nothing")
			}

			if !reflect.DeepEqual(out, tt.want) {
				t.Errorf("increment dchversion := '%v' IncrementRevision(v) = '%v', want '%v'", tt.args.v, out, tt.want)
			}
		})
	}
}

func TestParse(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name      string
		args      args
		want      Version
		wantError bool
	}{
		{
			name: `error`,
			args: args{
				s: "abc",
			},
			wantError: true,
		},
		{
			name: `native1`,
			args: args{
				s: "1.0.0",
			},
			want: NewVersion(0, "1.0.0", ""),
		},
		{
			name: `native2`,
			args: args{
				s: "4:1.0.0",
			},
			want: NewVersion(4, "1.0.0", ""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args.s)

			if !tt.wantError && err != nil {
				t.Errorf("cannot parse dchversion: %s", err)
			}

			if tt.wantError {
				if err != nil {
					t.Logf("got expected error: %s", err)
					return
				}
				t.Error("expected an error, got nothing")
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parse dchversion := '%v' Parse(s) = '%v', want '%v'", tt.args, got, tt.want)
			}
		})
	}
}
