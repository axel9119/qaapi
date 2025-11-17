package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

/* ============================
   Models
   ============================ */

type Question struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Text      string    `gorm:"type:text;not null" json:"text"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	Answers   []Answer  `gorm:"constraint:OnDelete:CASCADE" json:"answers,omitempty"`
}

type Answer struct {
	ID         int       `gorm:"primaryKey;autoIncrement" json:"id"`
	QuestionID int       `gorm:"not null;index" json:"question_id"`
	UserID     string    `gorm:"type:text;not null" json:"user_id"`
	Text       string    `gorm:"type:text;not null" json:"text"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

/* ============================
   Global DB connection
   ============================ */

var DB *gorm.DB

func connectDB() {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "host=localhost user=qa password=qa_pass dbname=qa_db port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("DB connect error: %v", err)
	}

	if err := db.AutoMigrate(&Question{}, &Answer{}); err != nil {
		log.Fatalf("AutoMigrate error: %v", err)
	}

	DB = db
}

/* ============================
   Repository layer
   ============================ */

func CreateQuestion(q *Question) error {
	return DB.Create(q).Error
}

func GetAllQuestions() ([]Question, error) {
	var qs []Question
	return qs, DB.Find(&qs).Error
}

func GetQuestionWithAnswers(id int) (*Question, error) {
	var q Question
	err := DB.Preload("Answers").First(&q, id).Error
	if err != nil {
		return nil, err
	}
	return &q, nil
}

func DeleteQuestion(id int) error {
	res := DB.Delete(&Question{}, id)
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return res.Error
}

func CreateAnswer(a *Answer) error {
	var q Question
	if err := DB.First(&q, a.QuestionID).Error; err != nil {
		return err
	}
	return DB.Create(a).Error
}

func GetAnswer(id int) (*Answer, error) {
	var a Answer
	if err := DB.First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func DeleteAnswer(id int) error {
	res := DB.Delete(&Answer{}, id)
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return res.Error
}

/* ============================
   Handlers
   ============================ */

func CreateQuestionHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	q := Question{Text: req.Text}
	if err := CreateQuestion(&q); err != nil {
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(q)
}

func GetAllQuestionsHandler(w http.ResponseWriter, r *http.Request) {
	qs, err := GetAllQuestions()
	if err != nil {
		http.Error(w, "internal", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(qs)
}

func GetQuestionHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/questions/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	q, err := GetQuestionWithAnswers(id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(q)
}

func DeleteQuestionHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/questions/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	if err := DeleteQuestion(id); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func CreateAnswerHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/questions/")
	parts := strings.Split(path, "/")

	if len(parts) < 2 || parts[1] != "answers" {
		http.Error(w, "bad path", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	var req struct {
		UserID string `json:"user_id"`
		Text   string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	a := Answer{QuestionID: id, UserID: req.UserID, Text: req.Text}

	if err := CreateAnswer(&a); err != nil {
		http.Error(w, "question not found", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(a)
}

func GetAnswerHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/answers/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	a, err := GetAnswer(id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(a)
}

func DeleteAnswerHandler(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/answers/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}

	if err := DeleteAnswer(id); err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

/* ============================
   Router + main()
   ============================ */

func main() {
	connectDB()

	mux := http.NewServeMux()

	mux.HandleFunc("/questions/", func(w http.ResponseWriter, r *http.Request) {
		// порядок важен: сначала POST /questions/{id}/answers
		if r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/answers") {
			CreateAnswerHandler(w, r)
			return
		}

		switch {
		case r.Method == http.MethodGet && (r.URL.Path == "/questions/" || r.URL.Path == "/questions"):
			GetAllQuestionsHandler(w, r)
		case r.Method == http.MethodPost && (r.URL.Path == "/questions/" || r.URL.Path == "/questions"):
			CreateQuestionHandler(w, r)
		case r.Method == http.MethodGet:
			GetQuestionHandler(w, r)
		case r.Method == http.MethodDelete:
			DeleteQuestionHandler(w, r)
		default:
			http.NotFound(w, r)
		}
	})

	mux.HandleFunc("/answers/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			GetAnswerHandler(w, r)
		case http.MethodDelete:
			DeleteAnswerHandler(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
