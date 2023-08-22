package hndlr

import (
	"net/http"
)

type Route struct {
	Name    string
	Path    string
	Method  string
	Handler http.HandlerFunc
}

var DirectoryRoutes = []Route{
	Route{
		"Add User",
		"/api/v1/user",
		"POST",
		AddUser,
	},
	Route{
		"Update User",
		"/api/v1/user",
		"PUT",
		UpdateUser,
	},
	Route{
		"Get Users",
		"/api/v1/user",
		"GET",
		GetUser,
	},

	Route{
		"Delete User",
		"/api/v1/user",
		"DELETE",
		DeleteUser,
	},
}
