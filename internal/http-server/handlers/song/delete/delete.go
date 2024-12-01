package deletehandler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"effective_mobile/internal/lib/api/response"
	"effective_mobile/internal/lib/logger/sl"
	"effective_mobile/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type SongDeleter interface {
	DeleteSong(id int) error
}

func New(log *slog.Logger, songDeleter SongDeleter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.song.delete.New"

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

		if err := songDeleter.DeleteSong(id); err != nil {
			if errors.Is(err, storage.ErrSongNotFound) {
				log.Info("song not found", slog.String("id", idString))

				response.Error(w, r, http.StatusNotFound, "song not found")

				return
			}

			log.Error("failed to delete song", sl.Err(err))

			response.Error(w, r, http.StatusInternalServerError, "failed to delete song")

			return
		}

		log.Info("song deleted", slog.Int("id", id))

		render.JSON(w, r, response.OK())
	}
}
