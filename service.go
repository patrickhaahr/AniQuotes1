package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type Service struct {
	store  Storage
	mailer Mailer
}

func NewService(store Storage, mailer Mailer) *Service {
	return &Service{store: store, mailer: mailer}
}

func (s *Service) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/users/{userID}/parse-kindle-file", s.handleParseKindleFile).Methods("POST")
	router.HandleFunc("/cloud/send-daily-insights", s.handleSendDailyInsights).Methods("GET")
}

func (s *Service) handleParseKindleFile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["userID"]

	// Convert userID to int
	userIDint, err := strconv.Atoi(userID)
	if err != nil {
		log.Printf("Invalid user ID: %v", err)
		WriteJSON(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// Check if the user exists
	user, err := s.store.GetUserByID(userIDint)
	if err != nil {
		log.Printf("Error fetching user: %v", err)
		WriteJSON(w, http.StatusInternalServerError, "Error fetching user")
		return
	}
	if user == nil {
		log.Printf("User with ID %d does not exist", userIDint)
		WriteJSON(w, http.StatusBadRequest, "User does not exist")
		return
	}

	// Parse the file
	file, _, err := r.FormFile("file")
	if err != nil {
		log.Printf("Error parsing file: %v", err)
		WriteJSON(w, http.StatusBadRequest, fmt.Sprintf("Error parsing file: %v", err))
		return
	}
	defer file.Close()

	// Parse the multipart file
	raw, err := parseKindleExtractFile(file)
	if err != nil {
		log.Printf("Error parsing Kindle extract file: %v", err)
		WriteJSON(w, http.StatusBadRequest, fmt.Sprintf("Error parsing file: %v", err))
		return
	}

	if err := s.createDataFromRawBook(raw, userIDint); err != nil {
		log.Printf("Error creating data from raw book: %v", err)
		WriteJSON(w, http.StatusInternalServerError, fmt.Sprintf("Error creating data from raw book: %v", err))
		return
	}

	WriteJSON(w, http.StatusOK, "Successfully parsed file")
}

func (s *Service) handleSendDailyInsights(w http.ResponseWriter, r *http.Request) {
	log.Println("handleSendDailyInsights called")

	// get users
	users, err := s.store.GetUsers()
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		WriteJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// loop over users and get random highlights
	for _, u := range users {
		log.Printf("Processing user: %v", u.Email)
		hs, err := s.store.GetRandomHighlights(3, u.ID)
		if err != nil {
			log.Printf("Error fetching highlights for user %v: %v", u.Email, err)
			WriteJSON(w, http.StatusInternalServerError, err.Error())
			return
		}

		if len(hs) == 0 {
			log.Printf("No highlights found for user %v", u.Email)
			continue
		}

		insights, err := s.buildInsights(hs)
		if err != nil {
			log.Printf("Error building insights for user %v: %v", u.Email, err)
			WriteJSON(w, http.StatusInternalServerError, err.Error())
			return
		}

		err = s.mailer.SendInsights(insights, u)
		if err != nil {
			log.Printf("Error sending insights to user %v: %v", u.Email, err)
			WriteJSON(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	log.Println("Successfully sent daily insights")
	WriteJSON(w, http.StatusOK, nil)
}

func parseKindleExtractFile(file multipart.File) (*RawExtractBook, error) {
	decoder := json.NewDecoder(file)

	raw := new(RawExtractBook)
	if err := decoder.Decode(raw); err != nil {
		return nil, err
	}

	return raw, nil
}

func (s *Service) createDataFromRawBook(raw *RawExtractBook, userID int) error {
	_, err := s.store.GetBookByISBN(raw.ASIN)
	if err != nil {
		s.store.CreateBook(Book{
			ISBN:    raw.ASIN,
			Title:   raw.Title,
			Authors: raw.Authors,
		})
	}

	// create highlights
	hs := make([]Highlight, len(raw.Highlights))
	for i, h := range raw.Highlights {
		hs[i] = Highlight{
			Text:     h.Text,
			Location: h.Location.URL,
			Note:     h.Note,
			UserID:   userID,
			BookID:   raw.ASIN,
		}
	}

	err = s.store.CreateHighlights(hs)
	if err != nil {
		log.Println("Error creating highlights: ", err)
		return err
	}

	return nil
}

func (s *Service) buildInsights(hs []*Highlight) ([]*DailyInsight, error) {
	var insights []*DailyInsight

	for _, h := range hs {
		book, err := s.store.GetBookByISBN(h.BookID)
		if err != nil {
			return nil, err
		}

		insights = append(insights, &DailyInsight{
			Text:        h.Text,
			Note:        h.Note,
			BookAuthors: book.Authors,
			BookTitle:   book.Title,
		})
	}

	return insights, nil
}
