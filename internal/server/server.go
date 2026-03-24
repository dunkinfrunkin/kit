package server

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/dunkinfrunkin/kit/internal/auth"
	"github.com/dunkinfrunkin/kit/internal/crypto"
	"github.com/dunkinfrunkin/kit/internal/store"
)

type itemResponse struct {
	ID        int       `json:"id"`
	Namespace string    `json:"namespace"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	Content   string    `json:"content"`
	Author    string    `json:"author"`
	Version   int       `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toItemResponse(item *store.Item, content []byte) itemResponse {
	return itemResponse{
		ID:        item.ID,
		Namespace: item.Namespace,
		Type:      item.Type,
		Name:      item.Name,
		Content:   base64.StdEncoding.EncodeToString(content),
		Author:    item.Author,
		Version:   item.Version,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

type Server struct {
	store  *store.Store
	secret string
	mux    *http.ServeMux
}

func New(s *store.Store, secret string) *Server {
	srv := &Server{
		store:  s,
		secret: secret,
		mux:    http.NewServeMux(),
	}

	srv.mux.HandleFunc("POST /login", srv.handleLogin)

	srv.mux.HandleFunc("GET /skills", srv.handleListAll("skill"))
	srv.mux.HandleFunc("GET /hooks", srv.handleListAll("hook"))
	srv.mux.HandleFunc("GET /configs", srv.handleListAll("config"))
	srv.mux.HandleFunc("GET /search", srv.handleSearch)

	srv.mux.HandleFunc("GET /{namespace}/skills", srv.handleList("skill"))
	srv.mux.HandleFunc("GET /{namespace}/hooks", srv.handleList("hook"))
	srv.mux.HandleFunc("GET /{namespace}/configs", srv.handleList("config"))

	srv.mux.HandleFunc("GET /{namespace}/skills/{name}", srv.handleGet("skill"))
	srv.mux.HandleFunc("GET /{namespace}/hooks/{name}", srv.handleGet("hook"))
	srv.mux.HandleFunc("GET /{namespace}/configs/{name}", srv.handleGet("config"))

	srv.mux.HandleFunc("POST /{namespace}/skills", srv.handlePush("skill"))
	srv.mux.HandleFunc("POST /{namespace}/hooks", srv.handlePush("hook"))
	srv.mux.HandleFunc("POST /{namespace}/configs", srv.handlePush("config"))

	srv.mux.HandleFunc("DELETE /{namespace}/skills/{name}", srv.handleDelete("skill"))
	srv.mux.HandleFunc("DELETE /{namespace}/hooks/{name}", srv.handleDelete("hook"))
	srv.mux.HandleFunc("DELETE /{namespace}/configs/{name}", srv.handleDelete("config"))

	srv.mux.HandleFunc("GET /{namespace}/profiles", srv.handleListProfiles)
	srv.mux.HandleFunc("GET /{namespace}/profiles/{name}", srv.handleGetProfile)
	srv.mux.HandleFunc("POST /{namespace}/profiles", srv.handleCreateProfile)
	srv.mux.HandleFunc("POST /{namespace}/profiles/{name}/items", srv.handleAddProfileItem)
	srv.mux.HandleFunc("DELETE /{namespace}/profiles/{name}", srv.handleDeleteProfile)

	return srv
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) authenticate(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")
	token := strings.TrimPrefix(h, "Bearer ")
	return auth.Verify(s.secret, token)
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Email == "" {
		s.error(w, http.StatusBadRequest, "email is required")
		return
	}

	token, err := auth.Sign(s.secret, body.Email)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "failed to sign token")
		return
	}

	s.json(w, http.StatusOK, map[string]string{
		"token": token,
		"email": body.Email,
	})
}

func (s *Server) handlePush(itemType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email, err := s.authenticate(r)
		if err != nil {
			s.error(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		var body struct {
			Name    string `json:"name"`
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			s.error(w, http.StatusBadRequest, "invalid request body")
			return
		}

		namespace := r.PathValue("namespace")
		content, err := base64.StdEncoding.DecodeString(body.Content)
		if err != nil {
			s.error(w, http.StatusBadRequest, "invalid base64 content")
			return
		}

		if isPersonal(namespace) {
			if ownerEmail(namespace) != email {
				s.error(w, http.StatusForbidden, "namespace does not match authenticated user")
				return
			}
			key := crypto.DeriveKey(s.secret, email)
			content, err = crypto.Encrypt(key, content)
			if err != nil {
				s.error(w, http.StatusInternalServerError, "encryption failed")
				return
			}
		}

		item, err := s.store.PushItem(namespace, itemType, body.Name, content, email)
		if err != nil {
			s.error(w, http.StatusInternalServerError, "failed to push item")
			return
		}

		s.json(w, http.StatusOK, toItemResponse(item, content))
	}
}

func (s *Server) handleGet(itemType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email, err := s.authenticate(r)
		if err != nil {
			s.error(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		namespace := r.PathValue("namespace")
		name := r.PathValue("name")

		item, err := s.store.GetItem(namespace, itemType, name)
		if err != nil {
			s.error(w, http.StatusNotFound, "item not found")
			return
		}

		content := item.Content
		if isPersonal(namespace) {
			if ownerEmail(namespace) != email {
				s.error(w, http.StatusForbidden, "namespace does not match authenticated user")
				return
			}
			key := crypto.DeriveKey(s.secret, email)
			content, err = crypto.Decrypt(key, content)
			if err != nil {
				s.error(w, http.StatusInternalServerError, "decryption failed")
				return
			}
		}

		s.json(w, http.StatusOK, toItemResponse(item, content))
	}
}

func (s *Server) handleList(itemType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := s.authenticate(r); err != nil {
			s.error(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		namespace := r.PathValue("namespace")
		items, err := s.store.ListItems(namespace, itemType)
		if err != nil {
			s.error(w, http.StatusInternalServerError, "failed to list items")
			return
		}

		s.json(w, http.StatusOK, items)
	}
}

func (s *Server) handleListAll(itemType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := s.authenticate(r); err != nil {
			s.error(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		items, err := s.store.ListAllItems(itemType)
		if err != nil {
			s.error(w, http.StatusInternalServerError, "failed to list items")
			return
		}

		s.json(w, http.StatusOK, items)
	}
}

func (s *Server) handleDelete(itemType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email, err := s.authenticate(r)
		if err != nil {
			s.error(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		namespace := r.PathValue("namespace")
		name := r.PathValue("name")

		if err := s.store.DeleteItem(namespace, itemType, name, email); err != nil {
			s.error(w, http.StatusInternalServerError, "failed to delete item")
			return
		}

		s.json(w, http.StatusOK, map[string]string{"status": "deleted"})
	}
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if _, err := s.authenticate(r); err != nil {
		s.error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	query := r.URL.Query().Get("q")
	items, err := s.store.SearchItems(query)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "search failed")
		return
	}

	s.json(w, http.StatusOK, items)
}

func (s *Server) handleListProfiles(w http.ResponseWriter, r *http.Request) {
	if _, err := s.authenticate(r); err != nil {
		s.error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	namespace := r.PathValue("namespace")
	profiles, err := s.store.ListProfiles(namespace)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "failed to list profiles")
		return
	}

	s.json(w, http.StatusOK, profiles)
}

func (s *Server) handleGetProfile(w http.ResponseWriter, r *http.Request) {
	if _, err := s.authenticate(r); err != nil {
		s.error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	namespace := r.PathValue("namespace")
	name := r.PathValue("name")

	profile, err := s.store.GetProfile(namespace, name)
	if err != nil {
		s.error(w, http.StatusNotFound, "profile not found")
		return
	}

	s.json(w, http.StatusOK, profile)
}

func (s *Server) handleCreateProfile(w http.ResponseWriter, r *http.Request) {
	email, err := s.authenticate(r)
	if err != nil {
		s.error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		s.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	namespace := r.PathValue("namespace")
	profile, err := s.store.CreateProfile(namespace, body.Name, email)
	if err != nil {
		s.error(w, http.StatusInternalServerError, "failed to create profile")
		return
	}

	s.json(w, http.StatusCreated, profile)
}

func (s *Server) handleAddProfileItem(w http.ResponseWriter, r *http.Request) {
	if _, err := s.authenticate(r); err != nil {
		s.error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var ref store.ProfileRef
	if err := json.NewDecoder(r.Body).Decode(&ref); err != nil {
		s.error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	namespace := r.PathValue("namespace")
	name := r.PathValue("name")

	if err := s.store.AddProfileItem(namespace, name, ref); err != nil {
		s.error(w, http.StatusInternalServerError, "failed to add profile item")
		return
	}

	s.json(w, http.StatusOK, map[string]string{"status": "added"})
}

func (s *Server) handleDeleteProfile(w http.ResponseWriter, r *http.Request) {
	email, err := s.authenticate(r)
	if err != nil {
		s.error(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	namespace := r.PathValue("namespace")
	name := r.PathValue("name")

	if err := s.store.DeleteProfile(namespace, name, email); err != nil {
		s.error(w, http.StatusInternalServerError, "failed to delete profile")
		return
	}

	s.json(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) json(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (s *Server) error(w http.ResponseWriter, status int, msg string) {
	s.json(w, status, map[string]string{"error": msg})
}

func isPersonal(namespace string) bool {
	return strings.HasPrefix(namespace, "@")
}

func ownerEmail(namespace string) string {
	return strings.TrimPrefix(namespace, "@")
}
