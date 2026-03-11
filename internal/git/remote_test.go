package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		workspace string
		repoSlug  string
		wantErr   bool
	}{
		{
			name:      "ssh with .git suffix",
			url:       "git@bitbucket.org:myworkspace/my-repo.git",
			workspace: "myworkspace",
			repoSlug:  "my-repo",
		},
		{
			name:      "ssh without .git suffix",
			url:       "git@bitbucket.org:myworkspace/my-repo",
			workspace: "myworkspace",
			repoSlug:  "my-repo",
		},
		{
			name:      "https with .git suffix",
			url:       "https://bitbucket.org/myworkspace/my-repo.git",
			workspace: "myworkspace",
			repoSlug:  "my-repo",
		},
		{
			name:      "https without .git suffix",
			url:       "https://bitbucket.org/myworkspace/my-repo",
			workspace: "myworkspace",
			repoSlug:  "my-repo",
		},
		{
			name:      "https with username",
			url:       "https://user@bitbucket.org/myworkspace/my-repo.git",
			workspace: "myworkspace",
			repoSlug:  "my-repo",
		},
		{
			name:      "url with trailing whitespace",
			url:       "git@bitbucket.org:myworkspace/my-repo.git\n",
			workspace: "myworkspace",
			repoSlug:  "my-repo",
		},
		{
			name:    "github url is not bitbucket",
			url:     "git@github.com:owner/repo.git",
			wantErr: true,
		},
		{
			name:    "empty url",
			url:     "",
			wantErr: true,
		},
		{
			name:    "random string",
			url:     "not-a-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := ParseRemoteURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.workspace, info.Workspace)
			assert.Equal(t, tt.repoSlug, info.RepoSlug)
		})
	}
}
