// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package webtheme

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseThemeMetaInfoToMap(t *testing.T) {
	assert.Equal(t, parseThemeMetaInfoToMap(`
		:root {
			--theme-display-name: unused;
		  --theme-display-name: "Dark (Red/Green Colorblind-Friendly)";
		  --is-dark-theme: true;
		  --color-diff-added-word-bg: #388bfd66;
		  --color-diff-added-row-bg: #388bfd26;
		}
	`), map[string]string{
		"--theme-display-name": "Dark (Red/Green Colorblind-Friendly)",
		"--is-dark-theme":      "true",
	})

	assert.Equal(t, parseThemeMetaInfoToMap(`
		:root {
			--theme-display-name: unused;
			--is-dark-theme: "true";
		}
		:root {
			--theme-display-name: "unused2";
			--is-dark-theme: "false";
		}
		:root {
			--theme-display-name: Light;
		}
	`), map[string]string{
		"--theme-display-name": "Light",
		"--is-dark-theme":      "false",
	})
}
