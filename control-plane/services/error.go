package services

import "fmt"

type RouteUUIDMatchError struct {
	Err error
}

func (e *RouteUUIDMatchError) Error() string { return e.Err.Error() }

var BadRouteRegistrationRequest = fmt.Errorf("bad register route request")
