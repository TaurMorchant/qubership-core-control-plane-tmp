package entity

import "github.com/go-errors/errors"

var LegacyRouteDisallowed = errors.Errorf("registering new routes for Legacy or Archived versions is prohibited")
