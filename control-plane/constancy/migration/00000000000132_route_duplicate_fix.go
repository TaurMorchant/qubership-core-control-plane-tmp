package migration

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/netcracker/qubership-core-control-plane/domain"
	"github.com/uptrace/bun"
	"strings"
)

func init() {
	migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
		err := db.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
			log.Info("#132 route duplicate fix")

			if err := removeConfigServerRoutes(ctx, &tx); err != nil {
				return err
			}

			if err := removeDuplicateRoutes(ctx, &tx); err != nil {
				return err
			}

			log.Info("Route duplicate fix (migration #132) applied successfully")
			return nil
		})
		return err
	}, func(ctx context.Context, db *bun.DB) error {
		return nil
	})
}

func removeConfigServerRoutes(ctx context.Context, db *bun.Tx) error {
	log.Info("Delete config server routes")

	routes := make([]*V16Route, 0)
	err := db.NewSelect().
		Model(&routes).
		Relation("HeaderMatchers").
		Where("rm_prefix = ?", "/api/v1/config-server/").
		Scan(ctx)
	if err != nil {
		log.Errorf("Error selecting routes by non-empty regexp: %v", err)
		return err
	}

	for _, route := range routes {
		log.Info("Delete config server route: %v", route)
		_, err := db.NewDelete().Model(route).WherePK().Exec(ctx)
		if err != nil {
			log.Errorf("couldn't delete config server route with PK '%d': %v", route.Id, err)
			return err
		}
	}

	log.Info("Delete config server routes successfully")
	return nil
}

func removeDuplicateRoutes(ctx context.Context, db *bun.Tx) error {
	routes := make([]*V132Route, 0)
	err := db.NewSelect().
		Model(&routes).
		Relation("HeaderMatchers").
		Scan(ctx)
	if err != nil {
		log.Errorf("Error selecting routes by non-empty regexp: %v", err)
		return err
	}

	deletedRouteIds := make(map[int32]bool, 0)
	for _, route := range routes {
		if _, exist := deletedRouteIds[route.Id]; exist {
			continue
		}

		log.Info("Found route: %v", route)
		saveRoute := false
		if strings.HasSuffix(route.Prefix, "/") {
			route.Prefix = route.Prefix[:len(route.Prefix)-1]
			saveRoute = true
			log.Info("Remove slash")
		}

		routeKey := GenerateKey(*route)
		duplicateRoutes := make([]*V16Route, 0)
		err := db.NewSelect().
			Model(&duplicateRoutes).
			Relation("HeaderMatchers").
			Where("routekey = ? and virtualhostid = ? and id != ?", routeKey, route.VirtualHostId, route.Id).
			Scan(ctx)
		if err != nil {
			log.Errorf("Error selecting routes: %v", err)
			return err
		}

		for _, duplicateRoute := range duplicateRoutes {
			log.Info("Delete headerMatchers for duplicated route: %v", duplicateRoute)
			_, err1 := db.NewDelete().Model(&V16HeaderMatcher{}).Where("routeid = ?", duplicateRoute.Id).Exec(ctx)
			if err1 != nil {
				log.Errorf("couldn't delete headerMatchers for route id = '%d': %v", duplicateRoute.Id, err1)
				return err1
			}
			log.Info("Delete route duplicate: %v", duplicateRoute)
			_, err := db.NewDelete().Model(duplicateRoute).WherePK().Exec(ctx)
			if err != nil {
				log.Errorf("couldn't delete route with PK '%d': %v", duplicateRoute.Id, err)
				return err
			}
			deletedRouteIds[duplicateRoute.Id] = true
		}

		if saveRoute {
			log.Info("Save route without suffix")
			route.RouteKey = routeKey
			_, err := db.NewUpdate().Model(route).Column("rm_prefix", "routekey").WherePK().Exec(ctx)
			if err != nil {
				log.Errorf("couldn't update route with PK '%d': %v", route.Id, err)
				return err
			}
		}
	}

	return nil
}

type V132Route struct {
	bun.BaseModel            `bun:"routes"`
	Id                       int32                `bun:",pk" json:"id"`
	Uuid                     string               `bun:"uuid,notnull,type:varchar,unique" json:"uuid"`
	VirtualHostId            int32                `bun:"virtualhostid,notnull" json:"virtualHostId"`
	RouteKey                 string               `bun:"routekey,notnull" json:"routeKey"`
	DirectResponseCode       uint32               `bun:"directresponse_status" json:"directResponseCode"`
	Prefix                   string               `bun:"rm_prefix" json:"prefix"`
	Regexp                   string               `bun:"rm_regexp" json:"regexp"`
	Path                     string               `bun:"rm_path" json:"path"`
	ClusterName              string               `bun:"ra_clustername" json:"clusterName"`
	HostRewrite              string               `bun:"ra_hostrewrite" json:"hostRewrite"`
	HostAutoRewrite          domain.NullBool      `bun:"ra_hostautorewrite" json:"hostAutoRewrite" swaggertype:"boolean"`
	PrefixRewrite            string               `bun:"ra_prefixrewrite" json:"prefixRewrite"`
	RegexpRewrite            string               `bun:"ra_regexprewrite" json:"regexpRewrite"`
	PathRewrite              string               `bun:"ra_pathrewrite" json:"pathRewrite"`
	Version                  int32                `bun:",notnull" json:"version"`
	Timeout                  domain.NullInt       `json:"timeout" swaggertype:"integer"`
	IdleTimeout              domain.NullInt       `bun:"idle_timeout" json:"idleTimeout" swaggertype:"integer"`
	DeploymentVersion        string               `bun:"deployment_version,notnull" json:"deploymentVersionString"`
	InitialDeploymentVersion string               `bun:"initialdeploymentversion,notnull" json:"initialDeploymentVersion"`
	Autogenerated            bool                 `bun:"autogenerated" json:"autogenerated"`
	HeaderMatchers           []*V132HeaderMatcher `bun:"rel:has-many,join:id=routeid" json:"headerMatchers"`
	RequestHeadersToAdd      []V132Header         `bun:"request_header_to_add,type:jsonb"`
	RequestHeadersToRemove   []string             `bun:"request_header_to_remove,type:jsonb"`
	Fallback                 sql.NullBool         `bun:"fallback,type:boolean" swaggertype:"boolean"`
	RateLimitId              string               `bun:"rate_limit_id,nullzero,notnull"`
	StatefulSessionId        int32                `bun:"statefulsessionid,nullzero,notnull" json:"statefulSessionId"`
}

type V132HeaderMatcher struct {
	bun.BaseModel  `bun:"header_matchers"`
	Id             int32           `bun:",pk" json:"id"`
	Name           string          `bun:",notnull" json:"name"`
	Version        int32           `bun:",notnull" json:"version"`
	ExactMatch     string          `bun:"exactmatch" json:"exactMatch"`
	SafeRegexMatch string          `bun:"saferegexmatch" json:"safeRegexMatch"`
	RangeMatch     V132RangeMatch  `bun:"rangematch,type:jsonb" json:"rangeMatch"`
	PresentMatch   domain.NullBool `bun:"presentmatch" json:"presentMatch" swaggertype:"boolean"`
	PrefixMatch    string          `bun:"prefixmatch" json:"prefixMatch"`
	SuffixMatch    string          `bun:"suffixmatch" json:"suffixMatch"`
	InvertMatch    bool            `bun:"invertmatch,default:false" json:"invertMatch" swaggertype:"boolean"`
	RouteId        int32           `bun:"routeid,notnull" json:"-"`
	Route          *V132Route      `bun:"rel:belongs-to,join:routeid=id" json:"-" yaml:"-"`
}

type V132RangeMatch struct {
	Start domain.NullInt `json:"start" swaggertype:"integer"`
	End   domain.NullInt `json:"end" swaggertype:"integer"`
}

type V132Header struct {
	Name  string
	Value string
}

func GenerateKey(route V132Route) string {
	routeKey := fromRoute(route)
	return Generate(routeKey)
}

func Generate(routeMatch V132RouteMatch) string {
	b, _ := json.Marshal(routeMatch)
	return fmt.Sprintf("%x", sha256.Sum256(b))
}

type V132RouteMatch struct {
	Prefix  string            `json:",omitempty"`
	Regexp  string            `json:",omitempty"`
	Path    string            `json:",omitempty"`
	Headers []V132HeaderMatch `json:",omitempty"`
	Version string            `json:",omitempty"`
}

type V132HeaderMatch struct {
	Name           string            `json:",omitempty"`
	ExactMatch     string            `json:",omitempty"`
	SafeRegexMatch string            `json:",omitempty"`
	RangeMatch     *V132RangeMatcher `json:",omitempty"`
	PresentMatch   *bool             `json:",omitempty"`
	PrefixMatch    string            `json:",omitempty"`
	SuffixMatch    string            `json:",omitempty"`
	InvertMatch    bool              `json:",omitempty"`
}

type V132RangeMatcher struct {
	Start int64 `json:"start"`
	End   int64 `json:"end"`
}

func fromRoute(route V132Route) V132RouteMatch {
	result := V132RouteMatch{
		Prefix:  route.Prefix,
		Regexp:  route.Regexp,
		Path:    route.Path,
		Version: route.InitialDeploymentVersion,
	}
	if route.HeaderMatchers != nil {
		hmKeys := make([]V132HeaderMatch, len(route.HeaderMatchers))
		for i, hm := range route.HeaderMatchers {
			hmKeys[i] = fromHeaderMatcher(*hm)
		}
		result.Headers = hmKeys
	}
	return result
}

func fromHeaderMatcher(hm V132HeaderMatcher) V132HeaderMatch {
	result := V132HeaderMatch{
		Name:           hm.Name,
		ExactMatch:     hm.ExactMatch,
		SafeRegexMatch: hm.SafeRegexMatch,
		PrefixMatch:    hm.PrefixMatch,
		SuffixMatch:    hm.SuffixMatch,
		InvertMatch:    hm.InvertMatch,
	}
	if hm.RangeMatch.Start.Valid && hm.RangeMatch.End.Valid {
		result.RangeMatch = &V132RangeMatcher{
			Start: hm.RangeMatch.Start.Int64,
			End:   hm.RangeMatch.End.Int64,
		}
	}
	if hm.PresentMatch.Valid {
		boolValue := hm.PresentMatch.Bool
		result.PresentMatch = &boolValue
	}
	return result
}
