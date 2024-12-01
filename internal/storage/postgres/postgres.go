package postgres

import (
	"database/sql"
	"effective_mobile/internal/domain/models"
	"effective_mobile/internal/storage"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Storage struct {
	db *sqlx.DB
}

func New(port int, host, username, dbname, password, sslMode string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sqlx.Open("postgres",
		fmt.Sprintf("host=%s port =%d user=%s dbname=%s password=%s sslmode=%s",
			host, port, username, dbname, password, sslMode))

	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) SaveSong(songData models.SongData) (int, error) {
	const op = "storage.postgres.SaveSong"

	var id int

	query := fmt.Sprintf(`
		INSERT INTO %s ("group", song, release_date, lyrics, link) 
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, songsTable,
	)

	var releaseDate interface{}
	if songData.ReleaseDate == "" {
		releaseDate = nil
	} else {
		releaseDate = songData.ReleaseDate
	}

	if err := s.db.QueryRowx(query, songData.Group, songData.Song, releaseDate, songData.Text, songData.Link).Scan(&id); err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return 0, storage.ErrSongExists
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

func (s *Storage) Songs(filter models.FilterSongData) ([]models.SongData, error) {
	const op = "storage.postgres.Songs"

	query := strings.Builder{}
	query.WriteString(fmt.Sprintf("SELECT id, \"group\", song, release_date, lyrics, link FROM %s WHERE 1=1", songsTable))

	args := make([]interface{}, 0)
	argId := 1

	if filter.Group != nil {
		query.WriteString(fmt.Sprintf(" AND \"group\"=$%d", argId))
		args = append(args, *filter.Group)
		argId++
	}

	if filter.Song != nil {
		query.WriteString(fmt.Sprintf(" AND song=$%d", argId))
		args = append(args, *filter.Song)
		argId++
	}

	if filter.ReleaseDate != nil {
		query.WriteString(fmt.Sprintf(" AND release_date=$%d", argId))
		args = append(args, *filter.ReleaseDate)
		argId++
	}

	offset := (filter.Page - 1) * filter.PerPage
	query.WriteString(fmt.Sprintf(" ORDER BY id DESC LIMIT $%d OFFSET $%d", argId, argId+1))
	args = append(args, filter.PerPage, offset)

	rows, err := s.db.Queryx(query.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	songs := make([]models.SongData, 0)
	for rows.Next() {
		var song models.SongData
		err := rows.StructScan(&song)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		songs = append(songs, song)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if len(songs) == 0 {
		return nil, fmt.Errorf("%s: %w", op, storage.ErrSongNotFound)
	}

	return songs, nil
}

func (s *Storage) Text(id int) (string, error) {
	const op = "storage.postgres.Text"

	var text string
	query := fmt.Sprintf(`SELECT lyrics FROM %s WHERE id = $1`, songsTable)

	err := s.db.Get(&text, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("%s: %w", op, storage.ErrSongNotFound)
		}

		return "", fmt.Errorf("%s: %w", op, err)
	}

	return text, nil
}

func (s *Storage) UpdateSong(id int, updateSong models.UpdateSongData) error {
	const op = "storage.postgres.UpdateSong"

	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argId := 1

	if updateSong.Group != nil {
		setValues = append(setValues, fmt.Sprintf("\"group\"=$%d", argId))
		args = append(args, *updateSong.Group)
		argId++
	}

	if updateSong.Song != nil {
		setValues = append(setValues, fmt.Sprintf("song=$%d", argId))
		args = append(args, *updateSong.Song)
		argId++
	}

	if updateSong.Link != nil {
		setValues = append(setValues, fmt.Sprintf("link=$%d", argId))
		args = append(args, *updateSong.Link)
		argId++
	}

	if updateSong.ReleaseDate != nil {
		setValues = append(setValues, fmt.Sprintf("release_date=$%d", argId))
		args = append(args, *updateSong.ReleaseDate)
		argId++
	}

	if updateSong.Text != nil {
		setValues = append(setValues, fmt.Sprintf("lyrics=$%d", argId))
		args = append(args, *updateSong.Text)
		argId++
	}

	query := fmt.Sprintf(`UPDATE %s SET %s WHERE id=$%d`, songsTable, strings.Join(setValues, ", "), argId)
	args = append(args, id)

	result, err := s.db.Exec(query, args...)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return storage.ErrSongExists
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrSongNotFound)
	}

	return nil
}

func (s *Storage) DeleteSong(id int) error {
	const op = "storage.postgres.DeleteSong"

	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, songsTable)

	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: %w", op, storage.ErrSongNotFound)
	}

	return nil
}
