package server

import (
	"fmt"
	netv1 "k8s.io/api/networking/v1"
	"strings"
)

type Route struct {
	Path     string
	Host     string
	PathType string
	Resource string
}

func NewRoute() *Route {
	return &Route{
		Path:     "",
		Host:     "",
		PathType: "prefix",
		Resource: "",
	}
}

func NewRouteFrom(path string, host string, pathType *netv1.PathType, resource string) *Route {
	r := NewRoute()
	r.Path = path
	r.Host = host
	if pathType != nil {
		switch *pathType {
		case netv1.PathTypeExact:
			r.PathType = "exact"
		case netv1.PathTypePrefix:
			r.PathType = "prefix"
		case netv1.PathTypeImplementationSpecific:
			r.PathType = "prefix"
		default:
			r.PathType = "prefix"
		}
	}
	r.Resource = resource
	return r
}

// creates an indented yaml string
func (r *Route) String() string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("\n  - path: %s", r.Path))
	if len(r.PathType) > 0 {
		b.WriteString(fmt.Sprintf("\n    pathType: %s", r.PathType))
	}
	if len(r.Host) > 0 {
		b.WriteString(fmt.Sprintf("\n    host: %s", r.PathType))
	}

	b.WriteString(fmt.Sprintf("\n    resource: %s", r.Resource))

	return b.String()
}
