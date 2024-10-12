package middleware

import (
	"database/sql"
	"log"
	"net/http"

	"timstack/internal/flash"
	CustomeTypes "timstack/types"

	"github.com/duo-labs/webauthn/webauthn"
	// Import the package where datastore is defined
	// Assuming this is the correct path
)

var (
	l            log.Logger                // Assume this is properly initialized
	webAuthn     *webauthn.WebAuthn        // Assume this is properly initialized
	passkeyStore CustomeTypes.PasskeyStore // This should be initialized with NewDBStore
)

func init() {
	// Initialize passkeyStore in the init function
	dbHost := "postgres://joshtheeuf:jc194980@localhost:5432/passkey?sslmode=disable"
	db, err := sql.Open("postgres", dbHost)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	// Create a new DBStore instance
	passkeyStore, err = CustomeTypes.NewDBStore(db)
	if err != nil {
		log.Fatal("Error creating DBStore:", err)
	}

	// Initialize webAuthn
	webAuthn, err = webauthn.New(&webauthn.Config{
		RPID:          "localhost",
		RPDisplayName: "Example Website",
		RPOrigin:      "http://localhost:9005",
		Timeout:       30000,
	})
	if err != nil {
		log.Fatal("Error creating webAuthn instance:", err)
	}
}

// LoggedInMiddleware checks if a user is logged in and redirects to the login page if not.
func LoggedInMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: url to redirect to should be passed as a parameter

		sid, err := r.Cookie("sid")
		if err != nil {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		//Retrieve the session from database, if the session is expiered redirect to login page
		// if no session is found return to landing page
		_, err = passkeyStore.GetSession(sid.Value)
		if err != nil {
			flash.Set(w, flash.Warning, "Error", err.Error())
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// User is logged in, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}
