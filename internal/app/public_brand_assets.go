package app

import (
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	engage "github.com/ranakdinesh/spur-engage"
	storage "github.com/ranakdinesh/spur-storage"
)

func registerPublicBrandAssetRoutes(r chi.Router, engageModule *engage.Module, storageModule *storage.Module) {
	r.Get("/engage/public/lead-capture/{formKey}/assets/{kind}", func(w http.ResponseWriter, r *http.Request) {
		if engageModule == nil || engageModule.Services == nil || engageModule.Services.Brands == nil ||
			storageModule == nil || storageModule.Services == nil || storageModule.Services.Storage == nil {
			http.NotFound(w, r)
			return
		}

		kind := chi.URLParam(r, "kind")
		config, err := engageModule.Services.Brands.GetPublicLeadCaptureConfig(r.Context(), chi.URLParam(r, "formKey"))
		if err != nil || config == nil || config.Brand == nil {
			http.NotFound(w, r)
			return
		}

		assetURL := ""
		switch kind {
		case "logo":
			assetURL = stringValue(config.Brand.LogoURL)
		case "favicon":
			assetURL = stringValue(config.Brand.FaviconURL)
		default:
			http.NotFound(w, r)
			return
		}

		objectID, ok := storageObjectIDFromURL(assetURL)
		if !ok {
			http.NotFound(w, r)
			return
		}
		object, content, err := storageModule.Services.Storage.GetObject(r.Context(), config.Brand.TenantID, objectID)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer content.Close()

		w.Header().Set("Content-Type", object.ContentType)
		w.Header().Set("Content-Disposition", mime.FormatMediaType("inline", map[string]string{"filename": object.FileName}))
		w.Header().Set("Cache-Control", "public, max-age=300")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		_, _ = io.Copy(w, content)
	})
}

func storageObjectIDFromURL(rawURL string) (uuid.UUID, bool) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return uuid.Nil, false
	}
	const prefix = "/storage/objects/"
	if !strings.HasPrefix(rawURL, prefix) {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(strings.TrimPrefix(rawURL, prefix))
	if err != nil || id == uuid.Nil {
		return uuid.Nil, false
	}
	return id, true
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
