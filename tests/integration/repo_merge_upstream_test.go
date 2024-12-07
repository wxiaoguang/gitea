// Copyright 2024 The Gitea Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package integration

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	auth_model "code.gitea.io/gitea/models/auth"
	repo_model "code.gitea.io/gitea/models/repo"
	"code.gitea.io/gitea/models/unittest"
	user_model "code.gitea.io/gitea/models/user"
	api "code.gitea.io/gitea/modules/structs"
	"code.gitea.io/gitea/modules/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepoMergeUpstream(t *testing.T) {
	onGiteaRun(t, func(*testing.T, *url.URL) {
		forkUser := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: 4})

		baseRepo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{ID: 1})
		baseUser := unittest.AssertExistsAndLoadBean(t, &user_model.User{ID: baseRepo.OwnerID})

		checkFileContent := func(exp string) {
			req := NewRequest(t, "GET", fmt.Sprintf("/%s/test-repo-fork/raw/branch/master/new-file.txt", forkUser.Name))
			resp := MakeRequest(t, req, http.StatusOK)
			require.Equal(t, resp.Body.String(), exp)
		}

		session := loginUser(t, forkUser.Name)
		token := getTokenForLoggedInUser(t, session, auth_model.AccessTokenScopeWriteRepository)

		// create a fork
		req := NewRequestWithJSON(t, "POST", fmt.Sprintf("/api/v1/repos/%s/%s/forks", baseUser.Name, baseRepo.Name), &api.CreateForkOption{
			Name: util.ToPointer("test-repo-fork"),
		}).AddTokenAuth(token)
		MakeRequest(t, req, http.StatusAccepted)
		forkRepo := unittest.AssertExistsAndLoadBean(t, &repo_model.Repository{OwnerID: forkUser.ID, Name: "test-repo-fork"})

		// add a file in base repo
		require.NoError(t, createOrReplaceFileInBranch(baseUser, baseRepo, "new-file.txt", "master", "test-content-1"))

		// the repo shows a prompt to "sync fork"
		resp := session.MakeRequest(t, NewRequestf(t, "GET", "/%s/test-repo-fork", forkUser.Name), http.StatusOK)
		htmlDoc := NewHTMLParser(t, resp.Body)
		respMsg, _ := htmlDoc.Find(".ui.message").Html()
		assert.Contains(t, respMsg, `This branch is 1 commit behind <a href="/user2/repo1/src/branch/master">user2/repo1:master</a>`)
		mergeUpstreamLink := htmlDoc.Find("button[data-url*='merge-upstream']").AttrOr("data-url", "")
		require.NotEmpty(t, mergeUpstreamLink)

		// click the "sync fork" button
		req = NewRequestWithValues(t, "POST", mergeUpstreamLink, map[string]string{"_csrf": GetUserCSRFToken(t, session)})
		session.MakeRequest(t, req, http.StatusOK)
		checkFileContent("test-content-1")

		// update the files
		require.NoError(t, createOrReplaceFileInBranch(forkUser, forkRepo, "new-file-other.txt", "master", "test-content-other"))
		require.NoError(t, createOrReplaceFileInBranch(baseUser, baseRepo, "new-file.txt", "master", "test-content-2"))
		resp = session.MakeRequest(t, NewRequestf(t, "GET", "/%s/test-repo-fork", forkUser.Name), http.StatusOK)
		htmlDoc = NewHTMLParser(t, resp.Body)
		respMsg, _ = htmlDoc.Find(".ui.message:not(.positive)").Html()
		assert.Contains(t, respMsg, `The base branch <a href="/user2/repo1/src/branch/master">user2/repo1:master</a> has new changes`)
		// and do the merge-upstream by API
		req = NewRequestWithJSON(t, "POST", fmt.Sprintf("/api/v1/repos/%s/test-repo-fork/merge-upstream", forkUser.Name), &api.MergeUpstreamRequest{
			Branch: "master",
		}).AddTokenAuth(token)
		resp = MakeRequest(t, req, http.StatusOK)
		checkFileContent("test-content-2")

		var mergeResp api.MergeUpstreamResponse
		DecodeJSON(t, resp, &mergeResp)
		assert.Equal(t, "merge", mergeResp.MergeStyle)
	})
}
