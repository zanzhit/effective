package songservice

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"effective_mobile/internal/clients"
	"effective_mobile/internal/domain/models"
	"effective_mobile/internal/service"
)

type SongService struct {
	songSaver    SongSaver
	songProvider SongProvider
	externalAPI  ExternalRequester
}

type SongSaver interface {
	SaveSong(song models.SongData) (int, error)
	UpdateSong(id int, updateSong models.UpdateSongData) error
}

type SongProvider interface {
	Songs(filter models.FilterSongData) ([]models.SongData, error)
	Text(id int) (string, error)
}

type ExternalRequester interface {
	FetchSongDetails(group, song string) (*models.SongData, error)
}

func New(songSaver SongSaver, songProvider SongProvider, externalAPI ExternalRequester) *SongService {
	return &SongService{
		songSaver:    songSaver,
		songProvider: songProvider,
		externalAPI:  externalAPI,
	}
}

func (s *SongService) SaveSong(group, song string) (int, error) {
	const op = "service/song-service/SaveSong"

	songDetail, err := s.externalAPI.FetchSongDetails(group, song)
	if err != nil {
		if errors.Is(err, clients.ErrBadRequest) {
			return 0, fmt.Errorf("%s: %w", op, clients.ErrBadRequest)
		}

		if errors.Is(err, clients.ErrInternal) {
			return 0, fmt.Errorf("%s: %w", op, clients.ErrInternal)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	parsedDate, err := time.Parse("02.01.2006", songDetail.ReleaseDate)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, service.ErrInvalidDateFormat)
	}
	songDetail.ReleaseDate = parsedDate.Format("2006-01-02")

	songData := models.SongData{
		Group:       group,
		Song:        song,
		ReleaseDate: songDetail.ReleaseDate,
		Text:        songDetail.Text,
		Link:        songDetail.Link,
	}

	id, err := s.songSaver.SaveSong(songData)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *SongService) UpdateSong(id int, updateSong models.UpdateSongData) error {
	const op = "service/song-service/UpdateSong"

	if isEmptyUpdate(updateSong) {
		return fmt.Errorf("%s: %w", op, service.ErrEmptyUpdate)
	}

	if updateSong.ReleaseDate != nil {
		parsedDate, err := time.Parse("02.01.2006", *updateSong.ReleaseDate)
		if err != nil {
			return fmt.Errorf("%s: %w", op, service.ErrInvalidDateFormat)
		}

		formattedDate := parsedDate.Format("2006-01-02")
		updateSong.ReleaseDate = &formattedDate
	}

	err := s.songSaver.UpdateSong(id, updateSong)
	if err != nil {
		return err
	}

	return nil
}

func (s *SongService) Songs(filter models.FilterSongData) ([]models.SongData, error) {
	const op = "service/song-service/Songs"

	if *filter.Group == "" {
		filter.Group = nil
	}

	if *filter.Song == "" {
		filter.Song = nil
	}

	if *filter.ReleaseDate == "" {
		filter.ReleaseDate = nil
	}

	if filter.ReleaseDate != nil {
		parsedDate, err := time.Parse("02.01.2006", *filter.ReleaseDate)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, service.ErrInvalidDateFormat)
		}

		formattedDate := parsedDate.Format("2006-01-02")
		filter.ReleaseDate = &formattedDate
	}

	songs, err := s.songProvider.Songs(filter)
	if err != nil {
		return nil, err
	}

	return songs, nil
}

func (s *SongService) Text(id, verse int) (string, error) {
	const op = "service/song-service/Text"

	if verse < 1 {
		return "", fmt.Errorf("%s: %w", op, service.ErrInvalidVerseNumber)
	}

	text, err := s.songProvider.Text(id)
	if err != nil {
		return "", err
	}

	verses := splitByVerses(text)

	if verse > len(verses) {
		return "", fmt.Errorf("%s: %w", op, service.ErrInvalidVerseNumber)
	}

	return verses[verse-1], nil
}

func isEmptyUpdate(req models.UpdateSongData) bool {
	return req.Group == nil && req.Song == nil &&
		req.ReleaseDate == nil && req.Link == nil && req.Text == nil
}

func splitByVerses(text string) []string {
	lines := strings.Split(text, "\n")

	verses := []string{}
	currentVerse := strings.Builder{}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			if currentVerse.Len() > 0 {
				verses = append(verses, currentVerse.String())
				currentVerse.Reset()
			}
		} else {
			if currentVerse.Len() > 0 {
				currentVerse.WriteString("\n")
			}
			currentVerse.WriteString(line)
		}
	}

	if currentVerse.Len() > 0 {
		verses = append(verses, currentVerse.String())
	}

	return verses
}
