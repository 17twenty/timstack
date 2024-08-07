package passkey

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	CustomeTypes "timstack/types"

	// Adjust this import path as needed

	"github.com/duo-labs/webauthn/webauthn"
)

var (
	l            log.Logger                // Assume this is properly initialized
	webAuthn     *webauthn.WebAuthn        // Assume this is properly initialized
	passkeyStore CustomeTypes.PasskeyStore // This should be initialized with NewDBStore
	passKeyUser  CustomeTypes.PasskeyUser
)

func init() {
	// Initialize passkeyStore in the init function
	dbHost := "postgres://joshtheeuf:jc194980@localhost:5432/passkey?sslmode=disable"
	db, err := sql.Open("postgres", dbHost)
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	passkeyStore, err = CustomeTypes.NewDBStore(db)
	if err != nil {
		log.Fatal("Error creating DBStore:", err)
	}

	// Initialize webAuthn
	webAuthn, err = webauthn.New(&webauthn.Config{
		RPID:          "localhost",
		RPDisplayName: "Example Website",
		RPOrigin:      "http://localhost:9005",
		Timeout:       30,
	})
	if err != nil {
		log.Fatal("Error creating webAuthn instance:", err)
	}
}

func BeginRegistration(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] begin registration ----------------------\\")

	username, err := getUsername(r)
	log.Println(username)
	if err != nil {
		log.Printf("[ERROR] can't get user name: %s", err.Error())
		http.Error(w, "Invalid username", http.StatusBadRequest)
		return
	}
	log.Println(username)

	user, err := passkeyStore.GetUser(username)
	if err != nil {
		// create a user of CustomeTypes.user
		log.Println("line 63")
		user = &CustomeTypes.User{
			ID:          username,
			DisplayName: username,
			Name:        username,
		}
		err = passkeyStore.SaveUser(user)
		log.Println(user)
		if err != nil {
			log.Printf("[ERROR] can't save user: %s", err.Error())
			JSONResponse(w, "", "Internal server error", http.StatusInternalServerError)
			return
		}
	} else {
		log.Printf("[INFO] user already exists")
	}

	options, _, err := webAuthn.BeginRegistration(user)
	if err != nil {
		log.Printf("[ERROR] can't begin registration: %s", err.Error())
		JSONResponse(w, "", fmt.Sprintf("Can't begin registration: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	// Create session
	sessionID, err := passkeyStore.CreateSession(string(user.WebAuthnID()))
	if err != nil {
		log.Printf("[ERROR] can't create session: %s", err.Error())
		JSONResponse(w, "", "Internal server error", http.StatusInternalServerError)
		return
	}

	JSONResponse(w, sessionID, options, http.StatusOK)
}

func FinishRegistration(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("Session-Key")
	session, err := passkeyStore.GetSession(sessionID)
	if err != nil {
		log.Printf("[ERROR] can't get session: %s", err.Error())
		JSONResponse(w, "", "Invalid session", http.StatusBadRequest)
		return
	}

	user, err := passkeyStore.GetUser(string(session.UserID))
	if err != nil {
		log.Printf("[ERROR] can't get user: %s", err.Error())
		JSONResponse(w, "", "User not found", http.StatusNotFound)
		return
	}

	credential, err := webAuthn.FinishRegistration(user, session, r)
	if err != nil {
		log.Printf("[ERROR] can't finish registration: %s", err.Error())
		JSONResponse(w, "", fmt.Sprintf("Can't finish registration: %s", err.Error()), http.StatusBadRequest)
		return
	}

	user.AddCredential(credential)
	err = passkeyStore.SaveUser(user)
	if err != nil {
		log.Printf("[ERROR] can't save user: %s", err.Error())
		JSONResponse(w, "", "Internal server error", http.StatusInternalServerError)
		return
	}

	err = passkeyStore.DeleteSession(sessionID)
	if err != nil {
		log.Printf("[WARN] can't delete session: %s", err.Error())
	}

	log.Printf("[INFO] finish registration ----------------------/")
	JSONResponse(w, "", "Registration Success", http.StatusOK)
}

func BeginLogin(w http.ResponseWriter, r *http.Request) {
	log.Printf("[INFO] begin login ----------------------\\")

	username, err := getUsername(r)
	if err != nil {
		log.Printf("[ERROR] can't get user name: %s", err.Error())
		http.Error(w, "Invalid username", http.StatusBadRequest)
		return
	}

	user, err := passkeyStore.GetUser(username)
	if err != nil {
		log.Printf("[ERROR] user not found: %s", err.Error())
		JSONResponse(w, "", "User not found", http.StatusNotFound)
		return
	}

	options, _, err := webAuthn.BeginLogin(user)
	if err != nil {
		log.Printf("[ERROR] can't begin login: %s", err.Error())
		JSONResponse(w, "", fmt.Sprintf("Can't begin login: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	sessionID, err := passkeyStore.CreateSession(string(user.WebAuthnID()))
	if err != nil {
		log.Printf("[ERROR] can't create session: %s", err.Error())
		JSONResponse(w, "", "Internal server error", http.StatusInternalServerError)
		return
	}

	JSONResponse(w, sessionID, options, http.StatusOK)
}

func FinishLogin(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("Session-Key")
	session, err := passkeyStore.GetSession(sessionID)
	if err != nil {
		log.Printf("[ERROR] can't get session: %s", err.Error())
		JSONResponse(w, "", "Invalid session", http.StatusBadRequest)
		return
	}

	user, err := passkeyStore.GetUser(string(session.UserID))
	if err != nil {
		log.Printf("[ERROR] can't get user: %s", err.Error())
		JSONResponse(w, "", "User not found", http.StatusNotFound)
		return
	}

	credential, err := webAuthn.FinishLogin(user, session, r)
	if err != nil {
		log.Printf("[ERROR] can't finish login: %s", err.Error())
		JSONResponse(w, "", fmt.Sprintf("Can't finish login: %s", err.Error()), http.StatusBadRequest)
		return
	}

	if credential.Authenticator.CloneWarning {
		log.Printf("[WARN] potential cloned authenticator detected")
	}

	err = user.UpdateCredential(credential)
	if err != nil {
		log.Printf("[ERROR] can't update credential: %s", err.Error())
		JSONResponse(w, "", "Internal server error", http.StatusInternalServerError)
		return
	}

	err = passkeyStore.SaveUser(user)
	if err != nil {
		log.Printf("[ERROR] can't save user: %s", err.Error())
		JSONResponse(w, "", "Internal server error", http.StatusInternalServerError)
		return
	}

	err = passkeyStore.DeleteSession(sessionID)
	if err != nil {
		log.Printf("[WARN] can't delete session: %s", err.Error())
	}

	log.Printf("[INFO] finish login ----------------------/")
	JSONResponse(w, "", "Login Success", http.StatusOK)
}

func JSONResponse(w http.ResponseWriter, sessionKey string, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Session-Key", sessionKey)
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func getUsername(r *http.Request) (string, error) {
	var u struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		return "", err
	}
	return u.Username, nil
}
