package httpmw

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ogen-go/ogen/ogenerrors"

	"github.com/rofleksey/dredge/internal/entity"
	"github.com/rofleksey/dredge/internal/http/gen"
)

// OgenErrorHandler wraps the ogen default handler to map known sentinel errors to HTTP statuses.
func OgenErrorHandler() gen.ErrorHandler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
		if errors.Is(err, ErrLoginRateLimited) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"too many login attempts"}`))

			return
		}

		if errors.Is(err, entity.ErrInvalidRule) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)

			b, _ := json.Marshal(map[string]string{"message": err.Error()})
			_, _ = w.Write(b)

			return
		}

		ogenerrors.DefaultErrorHandler(ctx, w, r, err)
	}
}
