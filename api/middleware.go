package api

import "encore.dev/middleware"

//encore:middleware target=tag:acknowledge
func AcknowledgeReferrerPolicyMiddleware(req middleware.Request, next middleware.Next) middleware.Response {
	resp := next(req)
	resp.Header().Set("Referrer-Policy", "no-referrer")
	return resp
}
