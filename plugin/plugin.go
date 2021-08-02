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
func New(client *github.Client, org string, team, access_team int64) admission.Plugin {
	return &plugin{
		client:      client,
		org:         org,
		team:        team,
		access_team: access_team,
	}
}

type plugin struct {
	client *github.Client

	org         string // members of this org are granted access
	team        int64  // members of this team are granted admin access
	access_team int64  // members of this team are granted access
}

func (p *plugin) Admit(ctx context.Context, req *admission.Request) (*drone.User, error) {
	u := req.User

	logrus.WithField("user", u.Login).
		Debugln("requesting system access")

	// check organization membership
	m, _, err := p.client.Organizations.GetOrgMembership(ctx, u.Login, p.org)
	if err != nil {
		logrus.WithError(err).
			WithField("user", u.Login).
			Debugln("cannot get organization membership")
		return nil, ErrAccessDenied
	}

	// if the user is an organization administrator
	// they are granted admin access to the system.
	if *m.Role == "admin" {
		logrus.WithField("user", u.Login).
			WithField("org", p.org).
			WithField("role", "admin").
			Debugln("granted admin system access")
		u.Admin = true
		return &u, err
	}

	// if an admin team is defined ...
	if p.team != 0 {
		// check team membership. if the user is a member
		// of the team they are granted administrator access
		// to the system.
		_, _, err = p.client.Teams.GetTeamMembership(ctx, p.team, u.Login)
		if err == nil {
			logrus.WithField("user", u.Login).
				WithField("org", p.org).
				WithField("team", p.team).
				Debugln("granted admin system access")
			u.Admin = true
			return &u, err
		}
	}

	if p.access_team != 0 {
		// check team membership. if the user is a member
		// of the team they are granted access to the system, otherwise they are not
		_, _, err = p.client.Teams.GetTeamMembership(ctx, p.access_team, u.Login)
		if err == nil {
			logrus.WithField("user", u.Login).
				WithField("org", p.org).
				WithField("team", p.access_team).
				Debugln("grant system access")
			return nil, nil
		} else {
			logrus.WithField("user", u.Login).
				WithField("org", p.org).
				WithField("access_team", p.access_team).
				Debugln("deny standard system access")

			return nil, ErrAccessDenied
		}
	}
	logrus.WithField("user", u.Login).
		WithField("org", p.org).
		Debugln("grant standard system access")

	return nil, nil

}
