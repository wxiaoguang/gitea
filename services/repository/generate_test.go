// Copyright 2019 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package repository

import (
	"os"
	"path/filepath"
	"testing"

	repo_model "code.gitea.io/gitea/models/repo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGiteaTemplate(t *testing.T) {
	var giteaTemplate = []byte(`
# Header

# All .go files
**.go

# All text files in /text/
text/*.txt

# All files in modules folders
**/modules/*
`)

	gt := newGiteaTemplateFileMatcher("", giteaTemplate)
	assert.Len(t, gt.globs, 3)

	tt := []struct {
		Path  string
		Match bool
	}{
		{Path: "main.go", Match: true},
		{Path: "sub/sub/foo.go", Match: true},

		{Path: "a.txt", Match: false},
		{Path: "text/a.txt", Match: true},
		{Path: "sub/text/a.txt", Match: false},
		{Path: "text/a.json", Match: false},

		{Path: "a/b/c/modules/README.md", Match: true},
		{Path: "a/b/c/modules/d/README.md", Match: false},
	}

	for _, tc := range tt {
		assert.Equal(t, tc.Match, gt.Match(tc.Path), "path: %s", tc.Path)
	}
}

func TestFilePathSanitize(t *testing.T) {
	assert.Equal(t, "test_CON", filePathSanitize("test_CON"))
	assert.Equal(t, "test CON", filePathSanitize("test CON "))
	assert.Equal(t, "__/traverse/__", filePathSanitize(".. /traverse/ .."))
	assert.Equal(t, "./__/a/_git/b_", filePathSanitize("./../a/.git/ b: "))
	assert.Equal(t, "_", filePathSanitize("CoN"))
	assert.Equal(t, "_", filePathSanitize("LpT1"))
	assert.Equal(t, "_", filePathSanitize("CoM1"))
	assert.Equal(t, "_", filePathSanitize("\u0000"))
	assert.Equal(t, "目标", filePathSanitize("目标"))
	assert.Equal(t, "目标", filePathSanitize("目标"))
}

func TestProcessGiteaTemplateFile(t *testing.T) {
	tmpDir := filepath.Join(t.TempDir(), "gitea-template-test")

	assertFileContent := func(path, expected string) {
		data, err := os.ReadFile(filepath.Join(tmpDir, path))
		if expected == "" {
			assert.ErrorIs(t, err, os.ErrNotExist)
			return
		}
		require.NoError(t, err)
		assert.Equal(t, expected, string(data), "file content mismatch for %s", path)
	}

	assertSymLink := func(path, expected string) {
		link, err := os.Readlink(filepath.Join(tmpDir, path))
		if expected == "" {
			assert.ErrorIs(t, err, os.ErrNotExist)
			return
		}
		require.NoError(t, err)
		assert.Equal(t, expected, link, "symlink target mismatch for %s", path)
	}

	require.NoError(t, os.MkdirAll(tmpDir+"/.gitea", 0o755))
	require.NoError(t, os.WriteFile(tmpDir+"/.gitea/template", []byte("**"), 0o644))
	require.NoError(t, os.WriteFile(tmpDir+"/normal", []byte("normal content"), 0o644))
	require.NoError(t, os.WriteFile(tmpDir+"/template", []byte("template from ${TEMPLATE_NAME}"), 0o644))

	require.NoError(t, os.WriteFile(tmpDir+"/link-target", []byte("link content"), 0o644))
	require.NoError(t, os.Symlink(tmpDir+"/link-target", tmpDir+"/link"))

	require.NoError(t, os.WriteFile(tmpDir+"/subst-${REPO_NAME}", []byte("dummy subst repo name"), 0o644))
	require.NoError(t, os.WriteFile(tmpDir+"/subst-${TEMPLATE_NAME}-ok", []byte("dummy subst template name ok"), 0o644))     // will succeed
	require.NoError(t, os.WriteFile(tmpDir+"/subst-${TEMPLATE_NAME}-link", []byte("dummy subst template name link"), 0o644)) // will fail
	require.NoError(t, os.Symlink(tmpDir+"/link-target", tmpDir+"/subst-template-repo-name-link"))

	assertSubstTemplateName := func(ok, link string) {
		assertFileContent("subst-${TEMPLATE_NAME}-ok", ok)
		assertFileContent("subst-${TEMPLATE_NAME}-link", link)
	}
	assertSubstTemplateName("dummy subst template name ok", "dummy subst template name link")

	templateRepo := &repo_model.Repository{Name: "template-repo-name"}
	generatedRepo := &repo_model.Repository{Name: "/../.gIt/name"}
	fileMatcher, _ := readGiteaTemplateFile(tmpDir)
	err := processGiteaTemplateFile(t.Context(), tmpDir, templateRepo, generatedRepo, fileMatcher)
	require.NoError(t, err)

	assertFileContent("no-such", "")
	assertFileContent("normal", "normal content")
	assertFileContent("template", "template from template-repo-name")
	assertSymLink("link", tmpDir+"/link-target")
	assertFileContent("link-target", "link content")
	assertFileContent("subst-${REPO_NAME}", "")
	assertFileContent("subst-/__/_gIt/name", "dummy subst repo name")

	// the paths with templates should have been removed
	assertSubstTemplateName("", "")
	assertFileContent("subst-template-repo-name-ok", "dummy subst template name ok") // to a regular file, succeed
	assertSymLink("subst-template-repo-name-link", tmpDir+"/link-target")            // to a link, fail and the target is unchanged
}

func TestTransformers(t *testing.T) {
	cases := []struct {
		name     string
		expected string
	}{
		{"SNAKE", "abc_def_xyz"},
		{"KEBAB", "abc-def-xyz"},
		{"CAMEL", "abcDefXyz"},
		{"PASCAL", "AbcDefXyz"},
		{"LOWER", "abc_def-xyz"},
		{"UPPER", "ABC_DEF-XYZ"},
		{"TITLE", "Abc_def-Xyz"},
	}

	input := "Abc_Def-XYZ"
	assert.Len(t, globalVars().defaultTransformers, len(cases))
	for i, c := range cases {
		tf := globalVars().defaultTransformers[i]
		require.Equal(t, c.name, tf.Name)
		assert.Equal(t, c.expected, tf.Transform(input), "case %s", c.name)
	}
}
