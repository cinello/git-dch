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

// git-dch, permits to generate debian changelog
// from git commits in a repository.
//
// Created: 2018-02-15

// http://honk.sigxcpu.org/projects/git-buildpackage/manual-html/man.gbp.dch.html
// https://manpages.debian.org/jessie/git-buildpackage/git-dch.1.en.html
//
// gbp dch [ --version ] [ --help ] [ --verbose ] [ --color= [auto|on|off] ] [ --color-scheme=COLOR_SCHEME ]
//   [ --debian-branch=branch_name ] [ --debian-tag=tag-format ] [ --upstream-tag=tag-format ] [ --ignore-branch ]
//   [ --snapshot | --release ] [ --auto | --since=commitish ] [ --new-version=version ]
//   [ --bpo | --nmu | --qa | --team ] [ --distribution=name ] [ --force-distribution ] [ --urgency=level ]
//   [ --[no-]full ] [ --[no-]meta ] [ --meta-closes=bug-close-tags ] [ --snapshot-number=expression ]
//   [ --id-length=number ] [ --git-log=git-log-options ] [ --[no-]git-author ] [ --[no-]multimaint ]
//   [ --[no-]multimaint-merge ] [ --spawn-editor=[always|snapshot|release] ] [ --commit-msg=msg-format ] [ --commit ]
//   [ --customizations= customization-file ] [path1 path2]

package main

import (
	"log"
	"os"

	"gitlab.yuribugelli.it/debian/git-dch-go/internal/app/git-dch"
)

func main() {
	if err := git_dch.RunApplication(); err != nil {
		log.SetOutput(os.Stderr)
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
	os.Exit(0)
}
