package texthandler

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"effective_mobile/internal/lib/api/response"
	"effective_mobile/internal/lib/logger/sl"
	"effective_mobile/internal/service"
	"effective_mobile/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type Response struct {
	response.Response
	Text string `json:"text,omitempty"`
}

type TextProvider interface {
	Text(id, verse int) (string, error)
}

func New(log *slog.Logger, textProvider TextProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.song.text.New"

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

		verse := r.URL.Query().Get("verse")
		verseNum, err := strconv.Atoi(verse)
		if err != nil {
			log.Error("invalid verse number format", sl.Err(err))

			response.Error(w, r, http.StatusBadRequest, "invalid verse number format")

			return
		}

		text, err := textProvider.Text(id, verseNum)
		if err != nil {
			if errors.Is(err, storage.ErrSongNotFound) {
				log.Info("songs not found")

				response.Error(w, r, http.StatusNotFound, "songs not found")

				return
			}

			if errors.Is(err, service.ErrInvalidVerseNumber) {
				response.Error(w, r, http.StatusBadRequest, "invalid verse parameter")

				return
			}

			log.Error("failed to find songs", sl.Err(err))

			response.Error(w, r, http.StatusInternalServerError, "failed to find songs")

			return
		}

		log.Info("verse founded")

		responseOK(w, r, text)
	}
}

func responseOK(w http.ResponseWriter, r *http.Request, text string) {
	render.JSON(w, r, Response{
		Response: response.OK(),
		Text:     text,
	})
}
