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

	"gopkg.in/src-d/go-git.v4/config"
)

func (gr *Repository) ConfigValue(section, key string) (value string, err error) {

	var (
		c *config.Config
	)

	if c, err = gr.repository.Config(); err != nil {
		return "", fmt.Errorf(textCannotGetConfigurationValue, err)
	}

	return c.Raw.Section(section).Option(key), nil
}
