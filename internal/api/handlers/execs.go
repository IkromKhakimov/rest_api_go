package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"restapi/internal/models"
	"restapi/internal/repository/sqlconnect"
	"restapi/pkg/utils"
	"strconv"
	"time"
)

func GetExecsHandler(w http.ResponseWriter, r *http.Request) {
	var execs []models.Exec
	execs, err := sqlconnect.GetExecsDbHandler(execs, r)
	if err != nil {
		return
	}

	response := struct {
		Status string        `json:"status"`
		Count  int           `json:"count"`
		Data   []models.Exec `json:"data"`
	}{
		Status: "success",
		Count:  len(execs),
		Data:   execs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetOneExecHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	// Handle Path parameter
	id, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println("Fatal error:", err)
		return
	}

	execs, err := sqlconnect.GetExecByID(id)
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(execs)

}

func AddExecHandler(w http.ResponseWriter, r *http.Request) {

	var newExecs []models.Exec
	var rawExecs []map[string]interface{}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &rawExecs)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	fields := GetFieldNames(models.Exec{})

	allowedFields := make(map[string]struct{})
	for _, field := range fields {
		allowedFields[field] = struct{}{}
	}

	for _, teacher := range rawExecs {
		for key := range teacher {
			_, ok := allowedFields[key]
			if !ok {
				http.Error(w, "Unacceptable field found in request", http.StatusBadRequest)
			}
		}
	}

	err = json.Unmarshal(body, &newExecs)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}

	for _, exec := range newExecs {
		err := CheckBlankFields(exec)
		if err != nil {
			http.Error(w, "Invalid Request Body", http.StatusBadRequest)
			return
		}
	}

	addedExecs, err := sqlconnect.AddExecDbHandler(newExecs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	response := struct {
		Status string        `json:"status"`
		Count  int           `json:"count"`
		Data   []models.Exec `json:"data"`
	}{
		Status: "success",
		Count:  len(addedExecs),
		Data:   addedExecs,
	}
	json.NewEncoder(w).Encode(response)
}

func PatchExecsHandlerOld(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Teacher ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}

	db, err := sqlconnect.ConnectDb()
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	var existingExec models.Exec
	err = db.QueryRow("SELECT id, first_name, last_name, email, class, subject FROM execs WHERE id = ?", id).Scan(&existingExec.ID, &existingExec.FirstName, &existingExec.LastName)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Exec not found", http.StatusNotFound)
			return
		}
		fmt.Println(err)
		http.Error(w, "Unable to retrieve data", http.StatusInternalServerError)
		return
	}

	//for k, v := range updates {
	//	switch k {
	//	case "first_name":
	//		existingTeacher.FirstName = v.(string)
	//	case "last_name":
	//		existingTeacher.LastName = v.(string)
	//	case "email":
	//		existingTeacher.Email = v.(string)
	//	case "class":
	//		existingTeacher.Class = v.(string)
	//	case "subject":
	//		existingTeacher.Subject = v.(string)
	//	}
	//}

	// Apply updates using reflect
	execVal := reflect.ValueOf(&existingExec).Elem()
	execType := execVal.Type()
	fmt.Println("Teacher Val:", execVal.Type().Field(1))

	for k, v := range updates {
		for i := 0; i < execVal.NumField(); i++ {
			fmt.Println("k from reflect mechanism", k)
			field := execType.Field(i)
			fmt.Println(field.Tag.Get("json"))
			if field.Tag.Get("json") == k+",omitempty" {
				if execVal.Field(i).CanSet() {
					fieldVal := execVal.Field(i)
					fmt.Println("fieldVal:", fieldVal)
					fmt.Println("teacherVal.Field(i).Type():", execVal.Field(i).Type())
					fmt.Println("reflect.ValueOf(v):", reflect.ValueOf(v))
					fieldVal.Set(reflect.ValueOf(v).Convert(execVal.Field(i).Type()))
				}
			}
		}
	}

	_, err = db.Exec("UPDATE teachers SET first_name = ?, last_name = ?, email = ? WHERE id = ?", existingExec.FirstName, existingExec.LastName, existingExec.Email, existingExec.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Error updating teacher", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingExec)
}

// PATCH /teachers
func PatchExecsHandler(w http.ResponseWriter, r *http.Request) {

	var updates []map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	err = sqlconnect.PatchExecs(updates)
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PATCH /teachers/{id}
func PatchOneExecHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Teacher ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Request Payload", http.StatusBadRequest)
		return
	}

	existingExec, err := sqlconnect.PatchOneExec(id, updates)
	if err != nil {
		log.Println(err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingExec)
}

func DeleteOneExecHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid Teacher ID", http.StatusBadRequest)
		return
	}

	err = sqlconnect.DeleteOneExec(id)
	if err != nil {
		log.Println(err)
		return
	}

	// Alternative Apporch
	//w.WriteHeader(http.StatusNoContent)

	// Response Body
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status string `json:"status"`
		ID     int    `json:"id"`
	}{
		Status: "Exec successfully Deleted",
		ID:     id,
	}
	json.NewEncoder(w).Encode(response)
}

func DeleteExecHandler(w http.ResponseWriter, r *http.Request) {

	var ids []int
	err := json.NewDecoder(r.Body).Decode(&ids)
	if err != nil {
		log.Println(err)
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	deletedIds, err := sqlconnect.DeleteExecs(ids)
	if err != nil {
		log.Println(err)
		return
	}

	// Response Body
	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Status     string `json:"status"`
		DeletedIDs []int  `json:"deleted_ids"`
	}{
		Status:     "Exec successfully Deleted",
		DeletedIDs: deletedIds,
	}
	json.NewEncoder(w).Encode(response)
}
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req = models.Exec{}
	// Data Validation
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username or password are required", http.StatusBadRequest)
		return
	}

	// Search for user if user actually exist
	user, err := sqlconnect.GetUserByUsername(req.Username)
	if err != nil {
		http.Error(w, "Username or password are required", http.StatusBadRequest)
		return
	}

	// Is user active
	if user.InactiveStatus {
		http.Error(w, "Account is inactive", http.StatusForbidden)
		return
	}

	// Verify password
	err = utils.VerifyPassword(req.Password, user.Password)
	if err != nil {
		http.Error(w, "The password you entered", http.StatusForbidden)
		return
	}
	// Generate token
	tokenString, err := utils.SignToken(user.ID, user.Username, user.Role)
	if err != nil {
		http.Error(w, "Could not create login token", http.StatusForbidden)
		return
	}

	// Send token as a response or as a cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(24 * time.Hour),
		SameSite: http.SameSiteStrictMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "test",
		Value:    "testing",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Now().Add(24 * time.Hour),
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Token string `json:"token"`
	}{
		Token: tokenString,
	}
	json.NewEncoder(w).Encode(response)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "Bearer",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		Expires:  time.Unix(0, 0),
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message": "Logged out successfully"}`))
}

func UpdatePasswordHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	userId, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Error exec ID", http.StatusBadRequest)
		return
	}

	var req models.UpdatePasswordRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid Request Body", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	if req.CurrentPassword == "" || req.NewPassword == "" {
		http.Error(w, "Please enter password", http.StatusBadRequest)
		return
	}

	_, err = sqlconnect.UpdatePasswordInDb(userId, req.CurrentPassword, req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	//
	//// Send token as a response or as a cookie
	//http.SetCookie(w, &http.Cookie{
	//	Name:     "Bearer",
	//	Value:    token,
	//	Path:     "/",
	//	HttpOnly: true,
	//	Secure:   true,
	//	Expires:  time.Now().Add(24 * time.Hour),
	//	SameSite: http.SameSiteStrictMode,
	//})

	w.Header().Set("Content-Type", "application/json")
	response := struct {
		Message string `json:"message"`
	}{
		Message: "Password updated successfully",
	}
	json.NewEncoder(w).Encode(response)

}
