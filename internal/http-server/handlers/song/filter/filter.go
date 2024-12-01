package filterhandler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"effective_mobile/internal/domain/models"
	"effective_mobile/internal/lib/api/response"
	"effective_mobile/internal/lib/logger/sl"
	"effective_mobile/internal/storage"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	response.Response
	Songs []models.SongData `json:"songs,omitempty"`
}

type SongsProvider interface {
	Songs(models.FilterSongData) ([]models.SongData, error)
}

func New(log *slog.Logger, songsProvider SongsProvider, pageSizeLimit int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.song.filter.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		query := r.URL.Query()

		filter := models.FilterSongData{
			Group:       stringPtr(query.Get("group")),
			Song:        stringPtr(query.Get("song")),
			ReleaseDate: stringPtr(query.Get("releaseDate")),
			Page:        intOrDefault(query.Get("page"), 1),
			PerPage:     intOrDefault(query.Get("per_page"), pageSizeLimit),
		}

		songs, err := songsProvider.Songs(filter)
		if err != nil {
			if errors.Is(err, storage.ErrSongNotFound) {
				log.Info("songs not found")

				response.Error(w, r, http.StatusNotFound, "songs not found")

				return
			}

			log.Error("failed to find songs", sl.Err(err))

			response.Error(w, r, http.StatusInternalServerError, "failed to find songs")

			return
		}

		log.Info("songs founded")

		responseOK(w, r, songs)
	}
}

func stringPtr(s string) *string {
	return &s
}

func intOrDefault(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return parsed
}

func responseOK(w http.ResponseWriter, r *http.Request, songs []models.SongData) {
	render.JSON(w, r, Response{
		Response: response.OK(),
		Songs:    songs,
	})
}
