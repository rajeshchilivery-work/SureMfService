package firebase

import (
	"SureMFService/config"
	"context"
	"encoding/json"
	"log"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

var FirestoreClient *firestore.Client
var AuthClient *auth.Client

func Init() error {
	ctx := context.Background()

	credJSON, err := json.Marshal(map[string]string{
		"type":                        "service_account",
		"project_id":                  config.AppConfig.FirebaseProjectID,
		"private_key":                 config.AppConfig.FirebasePrivateKey,
		"client_email":                config.AppConfig.FirebaseClientEmail,
		"token_uri":                   "https://oauth2.googleapis.com/token",
		"auth_uri":                    "https://accounts.google.com/o/oauth2/auth",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
	})
	if err != nil {
		return err
	}

	opt := option.WithCredentialsJSON(credJSON)
	conf := &firebase.Config{ProjectID: config.AppConfig.FirebaseProjectID}

	app, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		log.Printf("Error initializing Firebase app: %v", err)
		return err
	}

	FirestoreClient, err = app.Firestore(ctx)
	if err != nil {
		log.Printf("Error initializing Firestore client: %v", err)
		return err
	}

	AuthClient, err = app.Auth(ctx)
	if err != nil {
		log.Printf("Error initializing Auth client: %v", err)
		return err
	}

	log.Println("Firebase (Firestore + Auth) initialized successfully")
	return nil
}

func Close() {
	if FirestoreClient != nil {
		if err := FirestoreClient.Close(); err != nil {
			log.Printf("Error closing Firestore client: %v", err)
		}
	}
}

// SetDocFields merges the given fields into a Firestore document (upsert).
func SetDocFields(collection, docID string, fields map[string]interface{}) error {
	ctx := context.Background()
	_, err := FirestoreClient.Collection(collection).Doc(docID).Set(ctx, fields, firestore.MergeAll)
	return err
}

// GetDoc retrieves a Firestore document into dest (pointer to struct).
func GetDoc(collection, docID string, dest interface{}) (bool, error) {
	ctx := context.Background()
	doc, err := FirestoreClient.Collection(collection).Doc(docID).Get(ctx)
	if err != nil {
		return false, nil // treat not-found as empty
	}
	return true, doc.DataTo(dest)
}
