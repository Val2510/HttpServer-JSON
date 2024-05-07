package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type User struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Age     string   `json:"age"`
	Friends []string `json:"friends"`
}

var (
	users      = make(map[string]*User)
	usersMutex sync.Mutex
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Post("/create", createUserHandler)
	r.Post("/make_friends", makeFriendsHandler)
	r.Delete("/user", deleteUserHandler)
	r.Get("/friends/{userID}", getFriendsHandler)
	r.Put("/{userID}", updateUserHandler)

	fmt.Println("Сервер запущен на http://localhost:8080")
	http.ListenAndServe(":8080", r)
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	var newUser User
	err := json.NewDecoder(r.Body).Decode(&newUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	usersMutex.Lock()
	defer usersMutex.Unlock()

	newUser.ID = fmt.Sprintf("%d", len(users)+1)
	users[newUser.ID] = &newUser

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": newUser.ID})
}

func makeFriendsHandler(w http.ResponseWriter, r *http.Request) {
	var friendship struct {
		SourceID string `json:"source_id"`
		TargetID string `json:"target_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&friendship)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	usersMutex.Lock()
	defer usersMutex.Unlock()

	sourceUser, ok1 := users[friendship.SourceID]
	targetUser, ok2 := users[friendship.TargetID]
	if !ok1 || !ok2 {
		http.Error(w, "One of the users not found", http.StatusBadRequest)
		return
	}

	sourceUser.Friends = append(sourceUser.Friends, friendship.TargetID)
	targetUser.Friends = append(targetUser.Friends, friendship.SourceID)

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "%s и %s теперь друзья\n", sourceUser.Name, targetUser.Name)
}

func deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	var target struct {
		TargetID string `json:"target_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	usersMutex.Lock()
	defer usersMutex.Unlock()

	user, ok := users[target.TargetID]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	delete(users, target.TargetID)

	for _, u := range users {
		for i, friendID := range u.Friends {
			if friendID == target.TargetID {
				u.Friends = append(u.Friends[:i], u.Friends[i+1:]...)
				break
			}
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Удалён пользователь: %s\n", user.Name)
}

func getFriendsHandler(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	usersMutex.Lock()
	defer usersMutex.Unlock()

	user, ok := users[userID]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user.Friends)
}

func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")

	var updateAge struct {
		NewAge string `json:"new_age"`
	}
	err := json.NewDecoder(r.Body).Decode(&updateAge)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	usersMutex.Lock()
	defer usersMutex.Unlock()

	user, ok := users[userID]
	if !ok {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	user.Age = updateAge.NewAge

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Successfully\n")
}
