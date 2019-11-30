package actions

import (
	"fmt"

	"github.com/google/go-github/github"
)

//getReleaseInfo gets the release info
func getReleaseInfo(payload []byte) (*github.RepositoryRelease, error) {
	e, err := github.ParseWebHook("release", payload)
	if err != nil {
		return nil, err
	}

	event, ok := e.(*github.ReleaseEvent)
	if !ok {
		return nil, fmt.Errorf("invalid event data")
	}

	if event.Release == nil {
		return nil, fmt.Errorf("event.Release is nil %s", string(payload))
	}

	if len(event.Release.Assets) == 0 {
		return nil, fmt.Errorf("no release assets found")
	}

	return event.Release, nil
}
