package bitbucket

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

type PRService struct {
	client    Client
	workspace string
	repoSlug  string
}

func NewPRService(client Client, workspace, repoSlug string) *PRService {
	return &PRService{
		client:    client,
		workspace: workspace,
		repoSlug:  repoSlug,
	}
}

func (s *PRService) basePath() string {
	return fmt.Sprintf("/repositories/%s/%s/pullrequests", s.workspace, s.repoSlug)
}

func (s *PRService) prPath(id int) string {
	return fmt.Sprintf("%s/%d", s.basePath(), id)
}

type ListPRsOpts struct {
	State string
	Page  int
}

func (s *PRService) List(ctx context.Context, opts ListPRsOpts) (*Paginated[PR], error) {
	query := url.Values{}
	if opts.State != "" {
		query.Set("state", opts.State)
	}
	if opts.Page > 0 {
		query.Set("page", strconv.Itoa(opts.Page))
	}

	resp, err := s.client.Get(ctx, s.basePath(), query)
	if err != nil {
		return nil, err
	}

	var result Paginated[PR]
	if err := DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *PRService) Get(ctx context.Context, id int) (*PR, error) {
	resp, err := s.client.Get(ctx, s.prPath(id), nil)
	if err != nil {
		return nil, err
	}

	var pr PR
	if err := DecodeResponse(resp, &pr); err != nil {
		return nil, err
	}

	return &pr, nil
}

func (s *PRService) Create(ctx context.Context, req CreatePRRequest) (*PR, error) {
	resp, err := s.client.Post(ctx, s.basePath(), req)
	if err != nil {
		return nil, err
	}

	var pr PR
	if err := DecodeResponse(resp, &pr); err != nil {
		return nil, err
	}

	return &pr, nil
}

func (s *PRService) Merge(ctx context.Context, id int, opts MergeOpts) (*PR, error) {
	path := fmt.Sprintf("%s/merge", s.prPath(id))
	resp, err := s.client.Post(ctx, path, opts)
	if err != nil {
		return nil, err
	}

	var pr PR
	if err := DecodeResponse(resp, &pr); err != nil {
		return nil, err
	}

	return &pr, nil
}

func (s *PRService) Approve(ctx context.Context, id int) error {
	path := fmt.Sprintf("%s/approve", s.prPath(id))
	resp, err := s.client.Post(ctx, path, nil)
	if err != nil {
		return err
	}

	return DecodeResponse(resp, nil)
}

func (s *PRService) Unapprove(ctx context.Context, id int) error {
	path := fmt.Sprintf("%s/approve", s.prPath(id))
	resp, err := s.client.Delete(ctx, path)
	if err != nil {
		return err
	}

	return DecodeResponse(resp, nil)
}

func (s *PRService) Decline(ctx context.Context, id int) (*PR, error) {
	path := fmt.Sprintf("%s/decline", s.prPath(id))
	resp, err := s.client.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var pr PR
	if err := DecodeResponse(resp, &pr); err != nil {
		return nil, err
	}

	return &pr, nil
}

func (s *PRService) Diff(ctx context.Context, id int) (string, error) {
	path := fmt.Sprintf("%s/diff", s.prPath(id))
	resp, err := s.client.Get(ctx, path, nil)
	if err != nil {
		return "", err
	}

	return ReadRawBody(resp)
}

func (s *PRService) ListComments(ctx context.Context, id int) (*Paginated[Comment], error) {
	path := fmt.Sprintf("%s/comments", s.prPath(id))
	resp, err := s.client.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var result Paginated[Comment]
	if err := DecodeResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (s *PRService) AddComment(ctx context.Context, id int, body string) (*Comment, error) {
	path := fmt.Sprintf("%s/comments", s.prPath(id))
	payload := map[string]interface{}{
		"content": map[string]string{
			"raw": body,
		},
	}

	resp, err := s.client.Post(ctx, path, payload)
	if err != nil {
		return nil, err
	}

	var comment Comment
	if err := DecodeResponse(resp, &comment); err != nil {
		return nil, err
	}

	return &comment, nil
}
