package kubernetes

import (
	"context"
	"errors"
)

type ServerVersion struct {
	Major      string
	Minor      string
	GitVersion string
	Platform   string
}

func GetServerVersion(ctx context.Context, client Client) (*ServerVersion, error) {
	set, err := client.Clientset()
	if err != nil {
		return nil, err
	}

	info, err := set.Discovery().ServerVersion()
	if err != nil {
		return nil, wrapVersionError(err)
	}
	if info == nil {
		return nil, &ErrVersionUnavailable{Err: errors.New("server version response was nil")}
	}

	return &ServerVersion{
		Major:      info.Major,
		Minor:      info.Minor,
		GitVersion: info.GitVersion,
		Platform:   info.Platform,
	}, nil
}
