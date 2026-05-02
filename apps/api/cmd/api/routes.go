package main

import (
	"net/http"

	"github.com/Chinsusu/ERP-v2/apps/api/internal/shared/auth"
)

type routeHandler = http.HandlerFunc

type routeGroup struct {
	mux          *http.ServeMux
	authSessions *auth.SessionManager
}

func newRouteGroup(mux *http.ServeMux, authSessions *auth.SessionManager) routeGroup {
	return routeGroup{
		mux:          mux,
		authSessions: authSessions,
	}
}

func (r routeGroup) public(path string, handler routeHandler) {
	r.mux.HandleFunc(path, handler)
}

func (r routeGroup) token(path string, handler routeHandler) {
	r.mux.Handle(path, auth.RequireSessionToken(r.authSessions, handler))
}

func (r routeGroup) permission(path string, permission auth.PermissionKey, handler routeHandler) {
	r.mux.Handle(path, auth.RequireSessionPermission(r.authSessions, permission, handler))
}
