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
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"gitlab.yuribugelli.it/debian/git-dch-go/pkg/dchversion"

	"gitlab.yuribugelli.it/debian/go-debian/changelog"
	"gitlab.yuribugelli.it/debian/go-debian/version"
)

func TestEntryToString(t *testing.T) {

	now := time.Now()

	type args struct {
		source    string
		v         dchversion.Version
		target    string
		urgency   string
		c         string
		author    string
		when      time.Time
		customArg [2]string
	}
	tests := []struct {
		name         string
		args         args
		want         []string
		wantError    bool
		wantArgError bool
	}{
		{
			name: "entryToString-ok",
			args: args{
				source:  "package-name",
				v:       dchversion.NewVersion(0, "1.0.0", "1"),
				target:  "unstable",
				urgency: "low",
				c:       "Text",
				author:  "My Name <my.name@mail.tld>",
				when:    now,
			},
			want: []string{
				"package-name (1.0.0-1) unstable; urgency=low\n" +
					"\n" +
					"Text\n" +
					"\n" +
					" -- My Name <my.name@mail.tld>  " + WhenToString(now) + "\n",
			},
			wantError: false,
		},
		{
			name: "entryToString-custom-argument-ok",
			args: args{
				source:    "package-name",
				v:         dchversion.NewVersion(0, "1.0.0", "1"),
				target:    "unstable",
				urgency:   "low",
				c:         "Text",
				author:    "My Name <my.name@mail.tld>",
				when:      now,
				customArg: [2]string{"custom", "value"},
			},
			want: []string{
				"package-name (1.0.0-1) unstable; urgency=low custom=value\n" +
					"\n" +
					"Text\n" +
					"\n" +
					" -- My Name <my.name@mail.tld>  " + WhenToString(now) + "\n",
				"package-name (1.0.0-1) unstable; custom=value urgency=low\n" +
					"\n" +
					"Text\n" +
					"\n" +
					" -- My Name <my.name@mail.tld>  " + WhenToString(now) + "\n",
			},
		},
		{
			name: "entryToString-no-source-error",
			args: args{
				source:  "",
				v:       dchversion.NewVersion(0, "1.0.0", "1"),
				target:  "unstable",
				urgency: "low",
				c:       "Text",
				author:  "My Name <my.name@mail.tld>",
				when:    now,
			},
			want:      []string{""},
			wantError: true,
		},
		{
			name: "entryToString-wrong-version-error",
			args: args{
				source:  "package-name",
				v:       dchversion.NewVersion(0, "", "1"),
				target:  "unstable",
				urgency: "low",
				c:       "Text",
				author:  "My Name <my.name@mail.tld>",
				when:    now,
			},
			want:      []string{""},
			wantError: true,
		},
		{
			name: "entryToString-no-target-error",
			args: args{
				source:  "package-name",
				v:       dchversion.NewVersion(0, "1.0.0", "1"),
				target:  "",
				urgency: "low",
				c:       "Text",
				author:  "My Name <my.name@mail.tld>",
				when:    now,
			},
			want:      []string{""},
			wantError: true,
		},
		{
			name: "entryToString-empty-urgency",
			args: args{
				source:  "package-name",
				v:       dchversion.NewVersion(0, "1.0.0", "1"),
				target:  "unstable",
				urgency: "",
				c:       "Text",
				author:  "My Name <my.name@mail.tld>",
				when:    now,
			},
			want: []string{
				"package-name (1.0.0-1) unstable; urgency=medium\n" +
					"\n" +
					"Text\n" +
					"\n" +
					" -- My Name <my.name@mail.tld>  " + WhenToString(now) + "\n",
			},
		},
		{
			name: "entryToString-no-urgency-error",
			args: args{
				source:  "package-name",
				v:       dchversion.NewVersion(0, "1.0.0", "1"),
				target:  "unstable",
				urgency: "wrong value",
				c:       "Text",
				author:  "My Name <my.name@mail.tld>",
				when:    now,
			},
			want:      []string{""},
			wantError: true,
		},
		{
			name: "entryToString-no-changelog",
			args: args{
				source:  "package-name",
				v:       dchversion.NewVersion(0, "1.0.0", "1"),
				target:  "unstable",
				urgency: "low",
				c:       "",
				author:  "My Name <my.name@mail.tld>",
				when:    now,
			},
			want: []string{
				"package-name (1.0.0-1) unstable; urgency=low\n" +
					"\n" +
					"  *\n" +
					"\n" +
					" -- My Name <my.name@mail.tld>  " + WhenToString(now) + "\n",
			},
		},
		{
			name: "entryToString-no-author-error",
			args: args{
				source:  "package-name",
				v:       dchversion.NewVersion(0, "1.0.0", "1"),
				target:  "unstable",
				urgency: "low",
				c:       "Text",
				author:  "",
				when:    now,
			},
			want:      []string{""},
			wantError: true,
		},
		{
			name: "entryToString-custom-argument-error",
			args: args{
				source:    "package-name",
				v:         dchversion.NewVersion(0, "1.0.0", "1"),
				target:    "unstable",
				urgency:   "low",
				c:         "Text",
				author:    "My Name <my.name@mail.tld>",
				when:      now,
				customArg: [2]string{"custom", "wrong value"},
			},
			want: []string{
				"package-name (1.0.0-1) unstable; urgency=low\n" +
					"\n" +
					"Text\n" +
					"\n" +
					" -- My Name <my.name@mail.tld>  " + WhenToString(now) + "\n",
			},
			wantArgError: true,
		},
		{
			name: "entryToString-error1",
			args: args{
				source:    "package-name",
				v:         dchversion.NewVersion(0, "1.0.0", "1"),
				target:    "unstable",
				urgency:   "low",
				c:         "Text",
				author:    "My Name <my.name@mail.tld>",
				customArg: [2]string{"custom value", "value"},
				when:      now,
			},
			want: []string{
				"package-name (1.0.0-1) unstable; urgency=low\n" +
					"\n" +
					"Text\n" +
					"\n" +
					" -- My Name <my.name@mail.tld>  " + WhenToString(now) + "\n",
			},
			wantArgError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			item, err := NewItem(
				test.args.source, test.args.v, test.args.target, test.args.urgency, test.args.c, test.args.author)

			if test.wantError && err == nil {
				t.Errorf("entryToString(args), expecting an error, got none")
				return
			}

			if !test.wantError && err != nil {
				t.Errorf("entryToString(args), not expecting an error, got one: %s", err)
				return
			}

			item.SetWhen(now)

			var argErr error
			if test.args.customArg[0] != "" || test.args.customArg[1] != "" {
				argErr = item.SetArgument(test.args.customArg[0], test.args.customArg[1])
			}

			if test.wantArgError && argErr == nil {
				t.Errorf("entryToString(args), expecting an error, got none")
				return
			}

			if !test.wantArgError && argErr != nil {
				t.Errorf("entryToString(args), not expecting an error, got one: %s", argErr)
				return
			}

			got := item.String()

			var found bool
			for _, want := range test.want {
				if got == want {
					found = true
				}
			}
			if !found {
				t.Errorf("entryToString(args), wanted:\n%s\ngot:\n%s\n", test.want, got)
			}
		})
	}
}

func TestGetEntries(t *testing.T) {

	now := time.Now()

	type args struct {
		source    string
		v         dchversion.Version
		target    string
		urgency   string
		c         string
		author    string
		when      time.Time
		customArg string
	}
	tests := []struct {
		name      string
		args      []args
		want      string
		wantError bool
	}{
		{
			name:      "getEntries-empty",
			args:      []args{},
			want:      "",
			wantError: false,
		},
		{
			name: "getEntries-one-item-changelog-empty",
			args: []args{
				{
					source:  "package-name",
					v:       dchversion.NewVersion(0, "1.0.0", "1"),
					target:  "unstable",
					urgency: "low",
					c:       "",
					author:  "My Name <my.name@mail.tld>",
					when:    now,
				},
			},
			want: "package-name (1.0.0-1) unstable; urgency=low\n" +
				"\n" +
				"  *\n" +
				"\n" +
				" -- My Name <my.name@mail.tld>  " + WhenToString(now) + "\n",
			wantError: false,
		},
		{
			name: "getEntries-one-item",
			args: []args{
				{
					source:  "package-name",
					v:       dchversion.NewVersion(0, "1.0.0", "1"),
					target:  "unstable",
					urgency: "low",
					c:       "Text",
					author:  "My Name <my.name@mail.tld>",
					when:    now,
				},
			},
			want: "package-name (1.0.0-1) unstable; urgency=low\n" +
				"\n" +
				"Text\n" +
				"\n" +
				" -- My Name <my.name@mail.tld>  " + WhenToString(now) + "\n",
			wantError: false,
		},
		{
			name: "getEntries-two-items",
			args: []args{
				{
					source:  "package-name",
					v:       dchversion.NewVersion(0, "1.1.0", "1"),
					target:  "unstable",
					urgency: "low",
					c:       "Text 2",
					author:  "My Name <my.name@mail.tld>",
					when:    now,
				},
				{
					source:  "package-name",
					v:       dchversion.NewVersion(0, "1.0.0", "1"),
					target:  "unstable",
					urgency: "low",
					c:       "Text",
					author:  "My Name <my.name@mail.tld>",
					when:    now,
				},
			},
			want: "package-name (1.1.0-1) unstable; urgency=low\n" +
				"\n" +
				"Text 2\n" +
				"\n" +
				" -- My Name <my.name@mail.tld>  " + WhenToString(now) + "\n" +
				"\n" +
				"package-name (1.0.0-1) unstable; urgency=low\n" +
				"\n" +
				"Text\n" +
				"\n" +
				" -- My Name <my.name@mail.tld>  " + WhenToString(now) + "\n",
			wantError: false,
		},
		{
			name: "getEntries-error",
			args: []args{
				{
					source:  "",
					v:       dchversion.NewVersion(0, "1.1.0", "1"),
					target:  "unstable",
					urgency: "low",
					c:       "Text 2",
					author:  "My Name <my.name@mail.tld>",
					when:    now,
				},
			},
			want:      "",
			wantError: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			var items Items

			for _, i := range test.args {
				item, err := NewItem(i.source, i.v, i.target, i.urgency, i.c, i.author)

				if test.wantError && err == nil {
					t.Errorf("getEntries(args), expecting an error, got none")
					return
				}

				if !test.wantError && err != nil {
					t.Errorf("getEntries(args), not expecting an error, got one: %s", err)
					return
				}

				items = append(items, item)
			}

			got := items.String()

			if got != test.want {
				t.Errorf("getEntries(args), wanted:\n%s\ngot:\n%s\n", test.want, got)
			}
		})
	}
}

func TestNewItemList(t *testing.T) {

	pwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	tests := []struct {
		name       string
		file       string
		fromReader bool
		want       changelog.ChangelogEntry
		when       string
		wantError  bool
	}{
		{
			name: "entryToString-ok-05",
			file: "/test/changelog.05",
			want: changelog.ChangelogEntry{
				Source:    "git-dch",
				Version:   version.Version{Epoch: 0, Version: "0.1.0", Revision: ""},
				Target:    "unstable",
				Arguments: map[string]string{"urgency": "medium"},
				Changelog: "  *\n",
				ChangedBy: "My Name <my.name@mail.tld>",
			},
			when: "Mon, 29 Jan 2018 10:46:01 +0100",
		},
		{
			name: "entryToString-ok-01",
			file: "/test/changelog.01",
			want: changelog.ChangelogEntry{
				Source:    "git-dch",
				Version:   version.Version{Epoch: 0, Version: "0.1.0", Revision: ""},
				Target:    "unstable",
				Arguments: map[string]string{"urgency": "medium"},
				Changelog: "  test\n\n",
				ChangedBy: "My Name <my.name@mail.tld>",
			},
			when: "Mon, 29 Jan 2018 10:46:01 +0100",
		},
		{
			name: "entryToString-ok-02",
			file: "/test/changelog.02",
			want: changelog.ChangelogEntry{
				Source:    "git-dch",
				Version:   version.Version{Epoch: 0, Version: "1.0.0", Revision: "1"},
				Target:    "jessie",
				Arguments: map[string]string{"urgency": "medium"},
				Changelog: "\n  test\n\n",
				ChangedBy: "My Name <my.name@mail.tld>",
			},
			when: "Fri, 02 Mar 2018 14:20:44 +0100",
		},
		{
			name: "entryToString-ok-03",
			file: "/test/changelog.03",
			want: changelog.ChangelogEntry{
				Source:    "git-dch",
				Version:   version.Version{Epoch: 3, Version: "3.1.0", Revision: "1"},
				Target:    "jessie",
				Arguments: map[string]string{"urgency": "medium"},
				Changelog: "\n  test\n\n",
				ChangedBy: "My Name <my.name@mail.tld>",
			},
			when: "Fri, 02 Mar 2018 14:20:44 +0100",
		},
		{
			name: "entryToString-ok-04",
			file: "/test/changelog.04",
			want: changelog.ChangelogEntry{
				Source:    "git-dch",
				Version:   version.Version{Epoch: 0, Version: "0.1.0", Revision: ""},
				Target:    "unstable",
				Arguments: map[string]string{"urgency": "medium"},
				Changelog: "  *\n",
				ChangedBy: "My Name <my.name@mail.tld>",
			},
			when: "Mon, 29 Jan 2018 10:46:01 +0100",
		},
		{
			name:       "entryToString-ok-04-reader",
			file:       "/test/changelog.04",
			fromReader: true,
			want: changelog.ChangelogEntry{
				Source:    "git-dch",
				Version:   version.Version{Epoch: 0, Version: "0.1.0", Revision: ""},
				Target:    "unstable",
				Arguments: map[string]string{"urgency": "medium"},
				Changelog: "  *\n",
				ChangedBy: "My Name <my.name@mail.tld>",
			},
			when: "Mon, 29 Jan 2018 10:46:01 +0100",
		},
		{
			name: "entryToString-empty",
			file: "/test/changelog.empty",
			want: changelog.ChangelogEntry{},
		},
		{
			name:      "entryToString-error-no-file",
			file:      "/test/changelog.wrong.01",
			want:      changelog.ChangelogEntry{},
			wantError: true,
		},
		{
			name:      "entryToString-error-01",
			file:      "/test/changelog.error.01",
			want:      changelog.ChangelogEntry{},
			wantError: true,
		},
		{
			name:      "entryToString-error-02",
			file:      "/test/changelog.error.02",
			want:      changelog.ChangelogEntry{},
			wantError: true,
		},
		{
			name:       "entryToString-error-02-reader",
			file:       "/test/changelog.error.02",
			fromReader: true,
			want:       changelog.ChangelogEntry{},
			wantError:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			var items Items
			var err error

			if test.fromReader {
				f, errFile := os.Open(pwd + filepath.FromSlash(test.file))
				if errFile != nil {
					t.Errorf("NewItemListFromFile(file), cannot create reader from %s\n", test.file)
				}
				defer f.Close()
				items, err = NewItemList(f)
			} else {
				items, err = NewItemListFromFile(pwd + filepath.FromSlash(test.file))
			}

			if test.wantError && err == nil {
				t.Errorf("NewItemListFromFile(file), expecting an error, got none")
				return
			}

			if !test.wantError && err != nil {
				t.Errorf("NewItemListFromFile(file), not expecting an error, got one: %s", err)
				return
			}

			if !test.wantError && test.when != "" {
				test.want.When, _ = time.Parse(time.RFC1123Z, test.when)
			}

			if len(items) > 0 {
				if !reflect.DeepEqual(items[0].ChangelogEntry, test.want) {
					t.Errorf("NewItemListFromFile(file), wanted:\n%s\ngot:\n%s\n", test.want, items[0].ChangelogEntry)
				}
			}
		})
	}
}

func TestSource(t *testing.T) {

	pwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "entryToString-ok-01",
			file: "/test/changelog.01",
			want: "git-dch",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			items, err := NewItemListFromFile(pwd + filepath.FromSlash(test.file))
			if err != nil {
				t.Errorf("Source(), cannot read changelog from %s\n", test.file)
			}

			got := items.Source()
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Source(file), wanted:\n%s\ngot:\n%s\n", test.want, got)
			}
		})
	}
}

func TestVersion(t *testing.T) {

	pwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	tests := []struct {
		name string
		file string
		want version.Version
	}{
		{
			name: "entryToString-ok-01",
			file: "/test/changelog.01",
			want: version.Version{Epoch: 0, Version: "0.1.0", Revision: ""},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			items, err := NewItemListFromFile(pwd + filepath.FromSlash(test.file))
			if err != nil {
				t.Errorf("Version(), cannot read changelog from %s\n", test.file)
			}

			got := items.Version()
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Version(), wanted:\n%s\ngot:\n%s\n", test.want, got)
			}
		})
	}
}

func TestTarget(t *testing.T) {

	pwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "entryToString-ok-01",
			file: "/test/changelog.01",
			want: "unstable",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			items, err := NewItemListFromFile(pwd + filepath.FromSlash(test.file))
			if err != nil {
				t.Errorf("Target(), cannot read changelog from %s\n", test.file)
			}

			got := items.Target()
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Target(), wanted:\n%s\ngot:\n%s\n", test.want, got)
			}
		})
	}
}

func TestChangelog(t *testing.T) {

	pwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "entryToString-ok-01",
			file: "/test/changelog.01",
			want: "  test\n\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			items, err := NewItemListFromFile(pwd + filepath.FromSlash(test.file))
			if err != nil {
				t.Errorf("Changelog(), cannot read changelog from %s\n", test.file)
			}

			got := items.Changelog()
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Changelog(), wanted:\n%s\ngot:\n%s\n", test.want, got)
			}
		})
	}
}

func TestAuthor(t *testing.T) {

	pwd, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	tests := []struct {
		name string
		file string
		want string
	}{
		{
			name: "entryToString-ok-01",
			file: "/test/changelog.01",
			want: "My Name <my.name@mail.tld>",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			items, err := NewItemListFromFile(pwd + filepath.FromSlash(test.file))
			if err != nil {
				t.Errorf("Author(), cannot read changelog from %s\n", test.file)
			}

			got := items.Author()
			if !reflect.DeepEqual(got, test.want) {
				t.Errorf("Author(), wanted:\n%s\ngot:\n%s\n", test.want, got)
			}
		})
	}
}
