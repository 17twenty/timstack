package CustomeTypes

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"timstack/database/store" // Assuming this is the correct path

	"github.com/duo-labs/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/sqlc-dev/pqtype"
)

// ErrCredentialNotFound is returned when a credential cannot be found
var ErrCredentialNotFound = errors.New("credential not found")

// PasskeyUser interface
type PasskeyUser interface {
	webauthn.User
	AddCredential(*webauthn.Credential)
	UpdateCredential(*webauthn.Credential) error
	Credentials() []webauthn.Credential
}

// User struct
type User struct {
	ID             string                `json:"id"`
	DisplayName    string                `json:"display_name"`
	Name           string                `json:"name"`
	CredentialsRaw json.RawMessage       `json:"credentials"`
	CreatedAt      time.Time             `json:"created_at"`
	UpdatedAt      time.Time             `json:"updated_at"`
	credentials    []webauthn.Credential // private field for parsed credentials
}

// WebAuthnID implements webauthn.User interface
func (u *User) WebAuthnID() []byte {
	return []byte(u.ID)
}

// WebAuthnName implements webauthn.User interface
func (u *User) WebAuthnName() string {
	return u.Name
}

// WebAuthnDisplayName implements webauthn.User interface
func (u *User) WebAuthnDisplayName() string {
	return u.DisplayName
}

// WebAuthnIcon implements webauthn.User interface
func (u *User) WebAuthnIcon() string {
	return "" // Implement if needed
}

// WebAuthnCredentials implements webauthn.User interface
func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return u.credentials
}

// Credentials method
func (u *User) Credentials() []webauthn.Credential {
	return u.credentials
}

// AddCredential method
func (u *User) AddCredential(credential *webauthn.Credential) {
	u.credentials = append(u.credentials, *credential)
	// Update CredentialsRaw
	raw, _ := json.Marshal(u.credentials)
	u.CredentialsRaw = raw
}

// UpdateCredential method
func (u *User) UpdateCredential(credential *webauthn.Credential) error {
	for i, c := range u.credentials {
		if bytes.Equal(c.ID, credential.ID) {
			u.credentials[i] = *credential
			// Update CredentialsRaw
			raw, _ := json.Marshal(u.credentials)
			u.CredentialsRaw = raw
			return nil
		}
	}
	return ErrCredentialNotFound
}

// PasskeyStore interface
type PasskeyStore interface {
	GetUser(username string) (PasskeyUser, error)
	SaveUser(PasskeyUser) error
	CreateSession(userID string) (string, error)
	GetSession(token string) (webauthn.SessionData, error)
	DeleteSession(token string) error
}

// DBStore struct
type DBStore struct {
	queries *store.Queries
}

// NewDBStore function
func NewDBStore(db *sql.DB) (*DBStore, error) {
	queries := store.New(db)
	// Return a new DBStore instance with the initialized queries
	return &DBStore{queries: queries}, nil
}

// GetUser function
func (s *DBStore) GetUser(username string) (PasskeyUser, error) {
	dbUser, err := s.queries.GetUserByName(context.Background(), username)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	user := &User{
		ID:             dbUser.ID,
		DisplayName:    dbUser.DisplayName,
		Name:           dbUser.Name,
		CredentialsRaw: nil,
		CreatedAt:      dbUser.CreatedAt,
		UpdatedAt:      dbUser.UpdatedAt,
	}

	// Unmarshal the CredentialsRaw field into the credentials slice
	if err := json.Unmarshal(dbUser.Credentials.RawMessage, &user.credentials); err != nil {
		return nil, fmt.Errorf("failed to unmarshal credentials: %w", err)
	}

	return user, nil
}

// SaveUser function
func (s *DBStore) SaveUser(user PasskeyUser) error {
	u, ok := user.(*User)
	if !ok {
		return fmt.Errorf("invalid user type")
	}

	// Ensure CredentialsRaw is up to date
	raw, err := json.Marshal(u.credentials)
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}
	u.CredentialsRaw = raw

	dbUser := store.CreateUserParams{
		ID:          u.ID,
		DisplayName: u.DisplayName,
		Name:        u.Name,
		Credentials: pqtype.NullRawMessage{RawMessage: u.CredentialsRaw, Valid: true},
		// CreatedAt and UpdatedAt will be set by the database
	}
	_, err = s.queries.CreateUser(context.Background(), dbUser)
	if err != nil {
		return err
	}
	return nil
}

// CreateSession function
func (s *DBStore) CreateSession(userID string) (string, error) {

	sessionID := uuid.NewString()
	session := store.InsertIntoSessionsParams{
		ID:      sessionID,
		UserID:  userID,
		Data:    pqtype.NullRawMessage{RawMessage: []byte("{}"), Valid: true}, // Initialize with empty JSON object
		Expires: time.Now().Add(time.Hour),                                    // Adjust expiration as needed
	}
	_, err := s.queries.InsertIntoSessions(context.Background(), session)
	return sessionID, err
}

// GetSession function
func (s *DBStore) GetSession(token string) (webauthn.SessionData, error) {
	session, err := s.queries.GetSession(context.Background(), token)
	if err != nil {
		return webauthn.SessionData{}, err
	}

	var sessionData webauthn.SessionData
	err = json.Unmarshal(session.Data.RawMessage, &sessionData) // Use Bytes() method
	if err != nil {
		return webauthn.SessionData{}, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	sessionData.UserID = []byte(session.UserID)

	return sessionData, nil
}

// DeleteSession function
func (s *DBStore) DeleteSession(token string) error {
	return s.queries.DeleteSession(context.Background(), token)
}