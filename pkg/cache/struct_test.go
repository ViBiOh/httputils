package cache_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func fetchRepository(_ context.Context, id int) (Repository, error) {
	var output Repository

	err := json.Unmarshal([]byte(githubRepoPayload), &output)
	output.ID = id

	return output, err
}

func noFetch(_ context.Context, _ int) (Repository, error) {
	return Repository{}, errors.New("not implemented")
}

func getRepository(t testing.TB) Repository {
	t.Helper()

	var output Repository

	err := json.Unmarshal([]byte(githubRepoPayload), &output)
	assert.NoError(t, err)

	return output
}

type jsonErrorItem struct {
	ID    int           `json:"id"`
	Value func() string `json:"value"`
}

type Repository struct {
	ID       int    `json:"id"`
	NodeID   string `json:"node_id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Private  bool   `json:"private"`
	Owner    struct {
		Login             string `json:"login"`
		ID                int    `json:"id"`
		NodeID            string `json:"node_id"`
		AvatarURL         string `json:"avatar_url"`
		GravatarID        string `json:"gravatar_id"`
		URL               string `json:"url"`
		HTMLURL           string `json:"html_url"`
		FollowersURL      string `json:"followers_url"`
		FollowingURL      string `json:"following_url"`
		GistsURL          string `json:"gists_url"`
		StarredURL        string `json:"starred_url"`
		SubscriptionsURL  string `json:"subscriptions_url"`
		OrganizationsURL  string `json:"organizations_url"`
		ReposURL          string `json:"repos_url"`
		EventsURL         string `json:"events_url"`
		ReceivedEventsURL string `json:"received_events_url"`
		Type              string `json:"type"`
		SiteAdmin         bool   `json:"site_admin"`
	} `json:"owner"`
	HTMLURL          string    `json:"html_url"`
	Description      string    `json:"description"`
	Fork             bool      `json:"fork"`
	URL              string    `json:"url"`
	ForksURL         string    `json:"forks_url"`
	KeysURL          string    `json:"keys_url"`
	CollaboratorsURL string    `json:"collaborators_url"`
	TeamsURL         string    `json:"teams_url"`
	HooksURL         string    `json:"hooks_url"`
	IssueEventsURL   string    `json:"issue_events_url"`
	EventsURL        string    `json:"events_url"`
	AssigneesURL     string    `json:"assignees_url"`
	BranchesURL      string    `json:"branches_url"`
	TagsURL          string    `json:"tags_url"`
	BlobsURL         string    `json:"blobs_url"`
	GitTagsURL       string    `json:"git_tags_url"`
	GitRefsURL       string    `json:"git_refs_url"`
	TreesURL         string    `json:"trees_url"`
	StatusesURL      string    `json:"statuses_url"`
	LanguagesURL     string    `json:"languages_url"`
	StargazersURL    string    `json:"stargazers_url"`
	ContributorsURL  string    `json:"contributors_url"`
	SubscribersURL   string    `json:"subscribers_url"`
	SubscriptionURL  string    `json:"subscription_url"`
	CommitsURL       string    `json:"commits_url"`
	GitCommitsURL    string    `json:"git_commits_url"`
	CommentsURL      string    `json:"comments_url"`
	IssueCommentURL  string    `json:"issue_comment_url"`
	ContentsURL      string    `json:"contents_url"`
	CompareURL       string    `json:"compare_url"`
	MergesURL        string    `json:"merges_url"`
	ArchiveURL       string    `json:"archive_url"`
	DownloadsURL     string    `json:"downloads_url"`
	IssuesURL        string    `json:"issues_url"`
	PullsURL         string    `json:"pulls_url"`
	MilestonesURL    string    `json:"milestones_url"`
	NotificationsURL string    `json:"notifications_url"`
	LabelsURL        string    `json:"labels_url"`
	ReleasesURL      string    `json:"releases_url"`
	DeploymentsURL   string    `json:"deployments_url"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	PushedAt         time.Time `json:"pushed_at"`
	GitURL           string    `json:"git_url"`
	SSHURL           string    `json:"ssh_url"`
	CloneURL         string    `json:"clone_url"`
	SvnURL           string    `json:"svn_url"`
	Homepage         string    `json:"homepage"`
	Size             int       `json:"size"`
	StargazersCount  int       `json:"stargazers_count"`
	WatchersCount    int       `json:"watchers_count"`
	Language         string    `json:"language"`
	HasIssues        bool      `json:"has_issues"`
	HasProjects      bool      `json:"has_projects"`
	HasDownloads     bool      `json:"has_downloads"`
	HasWiki          bool      `json:"has_wiki"`
	HasPages         bool      `json:"has_pages"`
	HasDiscussions   bool      `json:"has_discussions"`
	ForksCount       int       `json:"forks_count"`
	MirrorURL        any       `json:"mirror_url"`
	Archived         bool      `json:"archived"`
	Disabled         bool      `json:"disabled"`
	OpenIssuesCount  int       `json:"open_issues_count"`
	License          struct {
		Key    string `json:"key"`
		Name   string `json:"name"`
		SpdxID string `json:"spdx_id"`
		URL    string `json:"url"`
		NodeID string `json:"node_id"`
	} `json:"license"`
	AllowForking             bool   `json:"allow_forking"`
	IsTemplate               bool   `json:"is_template"`
	WebCommitSignoffRequired bool   `json:"web_commit_signoff_required"`
	Topics                   []any  `json:"topics"`
	Visibility               string `json:"visibility"`
	Forks                    int    `json:"forks"`
	OpenIssues               int    `json:"open_issues"`
	Watchers                 int    `json:"watchers"`
	DefaultBranch            string `json:"default_branch"`
	Permissions              struct {
		Admin    bool `json:"admin"`
		Maintain bool `json:"maintain"`
		Push     bool `json:"push"`
		Triage   bool `json:"triage"`
		Pull     bool `json:"pull"`
	} `json:"permissions"`
	TempCloneToken            string `json:"temp_clone_token"`
	AllowSquashMerge          bool   `json:"allow_squash_merge"`
	AllowMergeCommit          bool   `json:"allow_merge_commit"`
	AllowRebaseMerge          bool   `json:"allow_rebase_merge"`
	AllowAutoMerge            bool   `json:"allow_auto_merge"`
	DeleteBranchOnMerge       bool   `json:"delete_branch_on_merge"`
	AllowUpdateBranch         bool   `json:"allow_update_branch"`
	UseSquashPrTitleAsDefault bool   `json:"use_squash_pr_title_as_default"`
	SquashMergeCommitMessage  string `json:"squash_merge_commit_message"`
	SquashMergeCommitTitle    string `json:"squash_merge_commit_title"`
	MergeCommitMessage        string `json:"merge_commit_message"`
	MergeCommitTitle          string `json:"merge_commit_title"`
	SecurityAndAnalysis       struct {
		SecretScanning struct {
			Status string `json:"status"`
		} `json:"secret_scanning"`
		SecretScanningPushProtection struct {
			Status string `json:"status"`
		} `json:"secret_scanning_push_protection"`
		DependabotSecurityUpdates struct {
			Status string `json:"status"`
		} `json:"dependabot_security_updates"`
	} `json:"security_and_analysis"`
	NetworkCount     int `json:"network_count"`
	SubscribersCount int `json:"subscribers_count"`
}

const githubRepoPayload = `{"id":99679090,"node_id":"MDEwOlJlcG9zaXRvcnk5OTY3OTA5MA==","name":"httputils","full_name":"ViBiOh/httputils","private":false,"owner":{"login":"ViBiOh","id":2349470,"node_id":"MDQ6VXNlcjIzNDk0NzA=","avatar_url":"https://avatars.githubusercontent.com/u/2349470?v=4","gravatar_id":"","url":"https://api.github.com/users/ViBiOh","html_url":"https://github.com/ViBiOh","followers_url":"https://api.github.com/users/ViBiOh/followers","following_url":"https://api.github.com/users/ViBiOh/following{/other_user}","gists_url":"https://api.github.com/users/ViBiOh/gists{/gist_id}","starred_url":"https://api.github.com/users/ViBiOh/starred{/owner}{/repo}","subscriptions_url":"https://api.github.com/users/ViBiOh/subscriptions","organizations_url":"https://api.github.com/users/ViBiOh/orgs","repos_url":"https://api.github.com/users/ViBiOh/repos","events_url":"https://api.github.com/users/ViBiOh/events{/privacy}","received_events_url":"https://api.github.com/users/ViBiOh/received_events","type":"User","site_admin":false},"html_url":"https://github.com/ViBiOh/httputils","description":"Golang Web Server utilities","fork":false,"url":"https://api.github.com/repos/ViBiOh/httputils","forks_url":"https://api.github.com/repos/ViBiOh/httputils/forks","keys_url":"https://api.github.com/repos/ViBiOh/httputils/keys{/key_id}","collaborators_url":"https://api.github.com/repos/ViBiOh/httputils/collaborators{/collaborator}","teams_url":"https://api.github.com/repos/ViBiOh/httputils/teams","hooks_url":"https://api.github.com/repos/ViBiOh/httputils/hooks","issue_events_url":"https://api.github.com/repos/ViBiOh/httputils/issues/events{/number}","events_url":"https://api.github.com/repos/ViBiOh/httputils/events","assignees_url":"https://api.github.com/repos/ViBiOh/httputils/assignees{/user}","branches_url":"https://api.github.com/repos/ViBiOh/httputils/branches{/branch}","tags_url":"https://api.github.com/repos/ViBiOh/httputils/tags","blobs_url":"https://api.github.com/repos/ViBiOh/httputils/git/blobs{/sha}","git_tags_url":"https://api.github.com/repos/ViBiOh/httputils/git/tags{/sha}","git_refs_url":"https://api.github.com/repos/ViBiOh/httputils/git/refs{/sha}","trees_url":"https://api.github.com/repos/ViBiOh/httputils/git/trees{/sha}","statuses_url":"https://api.github.com/repos/ViBiOh/httputils/statuses/{sha}","languages_url":"https://api.github.com/repos/ViBiOh/httputils/languages","stargazers_url":"https://api.github.com/repos/ViBiOh/httputils/stargazers","contributors_url":"https://api.github.com/repos/ViBiOh/httputils/contributors","subscribers_url":"https://api.github.com/repos/ViBiOh/httputils/subscribers","subscription_url":"https://api.github.com/repos/ViBiOh/httputils/subscription","commits_url":"https://api.github.com/repos/ViBiOh/httputils/commits{/sha}","git_commits_url":"https://api.github.com/repos/ViBiOh/httputils/git/commits{/sha}","comments_url":"https://api.github.com/repos/ViBiOh/httputils/comments{/number}","issue_comment_url":"https://api.github.com/repos/ViBiOh/httputils/issues/comments{/number}","contents_url":"https://api.github.com/repos/ViBiOh/httputils/contents/{+path}","compare_url":"https://api.github.com/repos/ViBiOh/httputils/compare/{base}...{head}","merges_url":"https://api.github.com/repos/ViBiOh/httputils/merges","archive_url":"https://api.github.com/repos/ViBiOh/httputils/{archive_format}{/ref}","downloads_url":"https://api.github.com/repos/ViBiOh/httputils/downloads","issues_url":"https://api.github.com/repos/ViBiOh/httputils/issues{/number}","pulls_url":"https://api.github.com/repos/ViBiOh/httputils/pulls{/number}","milestones_url":"https://api.github.com/repos/ViBiOh/httputils/milestones{/number}","notifications_url":"https://api.github.com/repos/ViBiOh/httputils/notifications{?since,all,participating}","labels_url":"https://api.github.com/repos/ViBiOh/httputils/labels{/name}","releases_url":"https://api.github.com/repos/ViBiOh/httputils/releases{/id}","deployments_url":"https://api.github.com/repos/ViBiOh/httputils/deployments","created_at":"2017-08-08T10:09:11Z","updated_at":"2023-05-07T12:57:20Z","pushed_at":"2023-08-13T12:08:03Z","git_url":"git://github.com/ViBiOh/httputils.git","ssh_url":"git@github.com:ViBiOh/httputils.git","clone_url":"https://github.com/ViBiOh/httputils.git","svn_url":"https://github.com/ViBiOh/httputils","homepage":"","size":2255,"stargazers_count":3,"watchers_count":3,"language":"Go","has_issues":true,"has_projects":false,"has_downloads":false,"has_wiki":false,"has_pages":false,"has_discussions":false,"forks_count":0,"mirror_url":null,"archived":false,"disabled":false,"open_issues_count":0,"license":{"key":"mit","name":"MIT License","spdx_id":"MIT","url":"https://api.github.com/licenses/mit","node_id":"MDc6TGljZW5zZTEz"},"allow_forking":true,"is_template":false,"web_commit_signoff_required":false,"topics":[],"visibility":"public","forks":0,"open_issues":0,"watchers":3,"default_branch":"main","permissions":{"admin":true,"maintain":true,"push":true,"triage":true,"pull":true},"temp_clone_token":"","allow_squash_merge":false,"allow_merge_commit":false,"allow_rebase_merge":true,"allow_auto_merge":false,"delete_branch_on_merge":true,"allow_update_branch":false,"use_squash_pr_title_as_default":false,"squash_merge_commit_message":"COMMIT_MESSAGES","squash_merge_commit_title":"COMMIT_OR_PR_TITLE","merge_commit_message":"PR_TITLE","merge_commit_title":"MERGE_MESSAGE","security_and_analysis":{"secret_scanning":{"status":"enabled"},"secret_scanning_push_protection":{"status":"enabled"},"dependabot_security_updates":{"status":"enabled"}},"network_count":0,"subscribers_count":3}`
