package devtest

import (
	actions_model "code.gitea.io/gitea/models/actions"
	"code.gitea.io/gitea/modules/web"
	"code.gitea.io/gitea/routers/web/repo/actions"
	"code.gitea.io/gitea/services/context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"
)

func MockActionsRunsJobs(ctx *context.Context) {
	req := web.GetForm(ctx).(*actions.ViewRequest)

	resp := &actions.ViewResponse{}
	resp.State.CurrentJob.Steps = append(resp.State.CurrentJob.Steps, &actions.ViewJobStep{
		Summary:  "job-step-summary-0",
		Duration: time.Hour.String(),
		Status:   actions_model.StatusRunning.String(),
	})
	resp.State.CurrentJob.Steps = append(resp.State.CurrentJob.Steps, &actions.ViewJobStep{
		Summary:  "job-step-summary-1",
		Duration: time.Hour.String(),
		Status:   actions_model.StatusRunning.String(),
	})

	resp.Logs.StepsLog = []*actions.ViewStepLog{}
	for _, logCur := range req.LogCursors {
		if !logCur.Expanded {
			continue
		}
		cur := logCur.Cursor
		for i := 0; i < 3; i++ {
			cur++
			resp.Logs.StepsLog = append(resp.Logs.StepsLog, &actions.ViewStepLog{
				Step:    logCur.Step,
				Cursor:  cur,
				Started: time.Now().Unix(),
				Lines: []*actions.ViewStepLogLine{
					{
						Index:     cur,
						Message:   fmt.Sprintf("message for step %d cur %d", logCur.Step, cur),
						Timestamp: float64(time.Now().UnixNano()) / float64(time.Second),
					},
				},
			})
		}
	}
	time.Sleep(time.Duration(500+rand.IntN(500)) * time.Millisecond)
	ctx.JSON(http.StatusOK, resp)
}

func MockActionsRunsArtifacts(ctx *context.Context) {
	artifactsResponse := &actions.ArtifactsViewResponse{}
	ctx.JSON(http.StatusOK, artifactsResponse)
}
