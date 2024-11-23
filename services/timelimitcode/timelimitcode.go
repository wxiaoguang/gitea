package timelimitcode

import (
	user_model "code.gitea.io/gitea/models/user"
	"code.gitea.io/gitea/modules/cache"
	"context"
)

type Purpose string

const PurposeActivateAccount Purpose = "ActivateAccount"
const PurposeResetPassword Purpose = "ResetPassword"
const PurposeActivateEmail Purpose = "ActivateEmail"

type TimeLimitCode[T any] struct {
	CreateTimestamp uint64

	Purpose Purpose
	Payload T
}

func generateTimeLimitCode(ctx context.Context, purpose Purpose, key string, payload any) string {
	cache.GetCache().GetJSON()
	return ""
}

func getTimeLimitCode[T any](ctx context.Context, purpose Purpose) (ret T) {
	//return EncodeSha256(email)
	return ""
}

func VerifyUserTimeLimitCode(ctx context.Context, purpose Purpose, code string) *user_model.User {
	//return EncodeSha256(email)
	return nil
}

func VerifyEmailTimeLimitCode(ctx context.Context, purpose Purpose, code string) *user_model.EmailAddress {
	//return EncodeSha256(email)
	return nil
}
