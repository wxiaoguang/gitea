// Copyright 2021 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package common

import (
	"net/http"
	"time"

	giteaContext "code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/log"
)

// NewLoggerHandlerV2 is a handler that will log the routing to the default gitea log
func NewLoggerHandlerV2(level log.Level) func(next http.Handler) http.Handler {
	lh := logContextHandler{
		logLevel:         level,
		requestRecordMap: map[uint64]*logRequestRecord{},
	}
	lh.startSlowQueryDetector(3 * time.Second)

	lh.printLog = func(trigger LogRequestTrigger, reqRec *logRequestRecord) {
		if trigger == LogRequestStart {
			return
		}

		funcFileShort := ""
		funcLine := 0
		funcNameShort := ""

		reqRec.funcInfoMu.RLock()
		if reqRec.funcInfo != nil {
			funcFileShort, funcLine, funcNameShort = reqRec.funcInfo.funcFileShort, reqRec.funcInfo.funcLine, reqRec.funcInfo.funcNameShort
		} else {
			// we might not find all handlers, so if a handler is not processed by our `UpdateContextHandlerFuncInfo`, we won't know its information
			// in such case, we should debug to find what handler it is and use `UpdateContextHandlerFuncInfo` to report its information
			funcFileShort = "unknown-handler"
		}
		reqRec.funcInfoMu.RUnlock()

		var status int
		if v, ok := reqRec.responseWriter.(giteaContext.ResponseWriter); ok {
			status = v.Status()
		}

		logger := log.GetLogger("router")
		req := reqRec.httpRequest
		if trigger == LogRequestExecuting {
			_ = logger.Log(0, lh.logLevel, "handler: %s:%d(%s) still-executing %v %s for %s, elapsed %v",
				funcFileShort, funcLine, funcNameShort,
				log.ColoredMethod(req.Method), req.RequestURI, req.RemoteAddr,
				log.ColoredTime(time.Since(reqRec.startTime)),
			)
		} else {
			if reqRec.panicError != nil {
				_ = logger.Log(0, lh.logLevel, "handler: %s:%d(%s) failed %v %s for %s, panic in %v, err=%v",
					funcFileShort, funcLine, funcNameShort,
					log.ColoredMethod(req.Method), req.RequestURI, req.RemoteAddr,
					log.ColoredTime(time.Since(reqRec.startTime)),
					reqRec.panicError,
				)
			} else {
				_ = logger.Log(0, lh.logLevel, "handler: %s:%d(%s) completed %v %s for %s, %v %v in %v",
					funcFileShort, funcLine, funcNameShort,
					log.ColoredMethod(req.Method), req.RequestURI, req.RemoteAddr,
					log.ColoredStatus(status), log.ColoredStatus(status, http.StatusText(status)), log.ColoredTime(time.Since(reqRec.startTime)))
			}
		}
	}

	return lh.handler
}
