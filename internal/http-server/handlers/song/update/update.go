package updatehandler

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"effective_mobile/internal/domain/models"
	"effective_mobile/internal/lib/api/response"
	"effective_mobile/internal/lib/logger/sl"
	"effective_mobile/internal/service"
	"effective_mobile/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type SongUpdater interface {
	UpdateSong(id int, updateSong models.UpdateSongData) error
}

func New(log *slog.Logger, songUpdater SongUpdater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.song.update.New"

		log := log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		idString := chi.URLParam(r, "id")
		if idString == "" {
			log.Error("missing id parameter in path")

			response.Error(w, r, http.StatusBadRequest, "missing id in path")

			return
		}

		id, err := strconv.Atoi(idString)
		if err != nil {
			log.Error("invalid id format", sl.Err(err))

			response.Error(w, r, http.StatusBadRequest, "invalid id format")

			return
		}

		log.Info("id decoded", slog.Any("id", id))

		var req models.UpdateSongData

		err = render.DecodeJSON(r.Body, &req)
		if errors.Is(err, io.EOF) {
			log.Error("request body is empty")

			response.Error(w, r, http.StatusBadRequest, "empty request")

			return
		}

		if err != nil {
			log.Error("failed to decode request body", sl.Err(err))

			response.Error(w, r, http.StatusInternalServerError, "failed to decode request")

			return
		}

		log.Info("request body decoded", slog.Any("request", req))

		if err := songUpdater.UpdateSong(id, req); err != nil {
			if errors.Is(err, storage.ErrSongNotFound) {
				log.Info("song not found", slog.String("id", idString))

				response.Error(w, r, http.StatusNotFound, "song not found")

				return
			}

			if errors.Is(err, service.ErrEmptyUpdate) {
				log.Info("empty update request", slog.String("date", *req.ReleaseDate))

				response.Error(w, r, http.StatusBadRequest, "empty update request")

				return
			}

			if errors.Is(err, service.ErrInvalidDateFormat) {
				log.Info("invalid date format for update", slog.String("date", *req.ReleaseDate))

				response.Error(w, r, http.StatusBadRequest, "invalid date format for update")

				return
			}

			log.Error("failed to update song", sl.Err(err))

			response.Error(w, r, http.StatusInternalServerError, "failed to update song")

			return
		}

		log.Info("song is updated", slog.Int("id", id))

		render.JSON(w, r, response.OK())
	}
}
