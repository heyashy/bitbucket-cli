package bitbucket

import "time"

type Paginated[T any] struct {
	Size     int    `json:"size"`
	Page     int    `json:"page"`
	PageLen  int    `json:"pagelen"`
	Next     string `json:"next,omitempty"`
	Previous string `json:"previous,omitempty"`
	Values   []T    `json:"values"`
}

type User struct {
	DisplayName string `json:"display_name"`
	UUID        string `json:"uuid"`
	Nickname    string `json:"nickname"`
	AccountID   string `json:"account_id"`
	Links       Links  `json:"links"`
}

type Links struct {
	Self   *Link `json:"self,omitempty"`
	HTML   *Link `json:"html,omitempty"`
	Avatar *Link `json:"avatar,omitempty"`
}

type Link struct {
	Href string `json:"href"`
}

type Branch struct {
	Name string `json:"name"`
}

type Ref struct {
	Branch     Branch     `json:"branch"`
	Repository Repository `json:"repository"`
}

type Repository struct {
	FullName string `json:"full_name"`
	Name     string `json:"name"`
	UUID     string `json:"uuid"`
	Links    Links  `json:"links"`
}

type PR struct {
	ID                int        `json:"id"`
	Title             string     `json:"title"`
	Description       string     `json:"description"`
	State             string     `json:"state"`
	Author            User       `json:"author"`
	Source            Ref        `json:"source"`
	Destination       Ref        `json:"destination"`
	CloseSourceBranch bool       `json:"close_source_branch"`
	CreatedOn         time.Time  `json:"created_on"`
	UpdatedOn         time.Time  `json:"updated_on"`
	MergeCommit       *Commit    `json:"merge_commit,omitempty"`
	CommentCount      int        `json:"comment_count"`
	TaskCount         int        `json:"task_count"`
	Reviewers         []User     `json:"reviewers"`
	Participants      []Participant `json:"participants"`
	Links             Links      `json:"links"`
}

type Participant struct {
	User     User   `json:"user"`
	Role     string `json:"role"`
	Approved bool   `json:"approved"`
	State    string `json:"state"`
}

type Commit struct {
	Hash string `json:"hash"`
}

type Comment struct {
	ID        int       `json:"id"`
	Content   Content   `json:"content"`
	User      User      `json:"user"`
	CreatedOn time.Time `json:"created_on"`
	UpdatedOn time.Time `json:"updated_on"`
	Inline    *Inline   `json:"inline,omitempty"`
}

type Content struct {
	Raw    string `json:"raw"`
	Markup string `json:"markup"`
	HTML   string `json:"html"`
}

type Inline struct {
	Path string `json:"path"`
	From *int   `json:"from,omitempty"`
	To   *int   `json:"to,omitempty"`
}

type CreatePRRequest struct {
	Title             string `json:"title"`
	Description       string `json:"description,omitempty"`
	Source            Ref    `json:"source"`
	Destination       Ref    `json:"destination"`
	CloseSourceBranch bool   `json:"close_source_branch"`
	Reviewers         []User `json:"reviewers,omitempty"`
}

type MergeOpts struct {
	MergeStrategy     string `json:"merge_strategy,omitempty"`
	CloseSourceBranch *bool  `json:"close_source_branch,omitempty"`
	Message           string `json:"message,omitempty"`
}
