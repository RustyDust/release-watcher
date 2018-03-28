package providers

import (
	"encoding/json"
	"fmt"
	"github.com/rycus86/release-watcher/config"
	"github.com/rycus86/release-watcher/model"
	"net/http"
	"time"
)

type PyPIProvider struct {
	client *http.Client
}

type pypiResponse struct {
	Releases map[string][]struct {
		UploadTime string `json:"upload_time"`
	} `json:"releases"`
}

func (provider *PyPIProvider) Initialize() {
	provider.client = &http.Client{
		Timeout: config.GetTimeout("HTTP_TIMEOUT", "/var/secrets/pypi"),
	}

	RegisterProvider(provider)
}

func (provider *PyPIProvider) GetName() string {
	return "PyPI"
}

func (provider *PyPIProvider) FetchReleases(project model.Project) ([]model.Release, error) {
	var releases []model.Release

	apiUrl := fmt.Sprintf("https://pypi.python.org/pypi/%s/json", project.Repo)
	response, err := provider.client.Get(apiUrl)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var apiResponse = pypiResponse{}
	err = json.NewDecoder(response.Body).Decode(&apiResponse)
	if err != nil {
		return nil, err
	}

	for name, items := range apiResponse.Releases {
		for _, release := range items {
			published, err := time.Parse("2006-01-02T15:04:05", release.UploadTime)
			if err != nil {
				published = time.Now()
			}

			releases = append(releases, model.Release{
				Name: name,
				URL:  fmt.Sprintf("https://pypi.python.org/pypi/%s/%s", project.Repo, name),
				Date: published,

				Provider: provider,
				Project:  project,
			})

		}
	}

	return releases, nil
}
