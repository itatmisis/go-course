package linkservice

import (
	"context"
	"errors"
	"fmt"

	"linkshortener/internal/repos/linkrepo"
	"linkshortener/internal/utils/randutil"
)

const shortLinkLength = 5

var ErrLinkNotFound = errors.New("short link not found")

type Service struct {
	linkRepo *linkrepo.Repo
}

func New(linkRepo *linkrepo.Repo) *Service {
	return &Service{linkRepo: linkRepo}
}

func (s *Service) CreateLink(ctx context.Context, realLink string) (string, error) {
	shortLink, err := randutil.Alphanumeric(shortLinkLength)
	if err != nil {
		return "", fmt.Errorf("generate short link: %w", err)
	}

	err = s.linkRepo.CreateLink(ctx, shortLink, realLink)
	if err != nil {
		return "", fmt.Errorf("save link: %w", err)
	}

	return shortLink, nil
}

func (s *Service) GetRealLink(ctx context.Context, shortLink string) (string, error) {
	link, err := s.linkRepo.GetLink(ctx, shortLink)
	if errors.Is(err, linkrepo.ErrNotFound) {
		return "", ErrLinkNotFound
	}

	if err != nil {
		return "", fmt.Errorf("get link: %w", err)
	}

	return link.RealLink, nil
}
