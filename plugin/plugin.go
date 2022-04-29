// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Polyform License
// that can be found in the LICENSE file.

package plugin

import (
	"context"
	"errors"

	"github.com/drone/drone-go/drone"
	"github.com/drone/drone-go/plugin/admission"

	"github.com/google/go-github/v28/github"
	"github.com/sirupsen/logrus"
)

// ErrAccessDenied is returned if access is denied.
var ErrAccessDenied = errors.New("admission: access denied")

// New returns a new admission plugin.
func New(client *github.Client, orgs []string, teams []int64) admission.Plugin {
	return &plugin{
		client: client,
		orgs:   orgs,
		teams:  teams,
	}
}

type plugin struct {
	client *github.Client

	orgs  []string // members of those orgs are granted access
	teams []int64  // members of those teams are granted admin access
}

func (p *plugin) Admit(ctx context.Context, req *admission.Request) (*drone.User, error) {
	u := req.User

	logrus.WithField("user", u.Login).
		Debugln("requesting system access")

	// check organizations membership
	var e []error
	var userOrg string
	for _, org := range p.orgs {
		m, _, err := p.client.Organizations.GetOrgMembership(ctx, u.Login, org)
		if err != nil {
			e = append(e, err)
			continue
		}

		userOrg = org

		// if the user is an organization administrator
		// they are granted admin access to the system.
		if *m.Role == "admin" {
			logrus.WithField("user", u.Login).
				WithField("org", org).
				WithField("role", "admin").
				Debugln("granted admin system access")
			u.Admin = true
			return &u, err
		}
	}

	if userOrg == "" {
		logrus.WithField("errors", e).
			WithField("user", u.Login).
			WithField("orgs", p.orgs).
			Debugln("cannot get organization membership")
		return nil, ErrAccessDenied
	}

	for _, team := range p.teams {
		// check teams membership. if the user is a member
		// of the team they are granted administrator access
		// to the system.
		_, _, err := p.client.Teams.GetTeamMembership(ctx, team, u.Login)
		if err == nil {
			logrus.WithField("user", u.Login).
				WithField("team", team).
				Debugln("granted admin system access")
			u.Admin = true
			return &u, err
		}
	}

	logrus.WithField("user", u.Login).
		WithField("org", userOrg).
		Debugln("granted standard system access")

	return nil, nil
}
