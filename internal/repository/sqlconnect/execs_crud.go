package sqlconnect

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"restapi/internal/models"
	"restapi/pkg/utils"
	"strconv"
	"strings"
	"time"
)

func GetExecsDbHandler(execs []models.Exec, r *http.Request) ([]models.Exec, error) {
	db, err := ConnectDb()
	if err != nil {
		// http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	query := "SELECT id, first_name, last_name, email, username, user_created_at, inactive_status, role FROM execs WHERE 1=1"
	var args []interface{}

	query, args = utils.AddFilters(r, query, args)
	query = utils.AddSorting(r, query)

	rows, err := db.Query(query, args...)
	if err != nil {
		fmt.Println(err)
		// http.Error(w, "Database Query Error", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "error retrieving data")
	}
	defer rows.Close()

	//teacherList := make([]models.Teacher, 0)
	for rows.Next() {
		var exec models.Exec
		err := rows.Scan(&exec.ID, &exec.FirstName, &exec.LastName, &exec.Email, &exec.Username, &exec.UserCreatedAt, &exec.InactiveStatus, &exec.Role)
		if err != nil {
			return nil, utils.ErrorHandler(err, "error retrieving data")
		}
		execs = append(execs, exec)
	}
	return execs, nil
}

func GetExecByID(id int) (models.Exec, error) {
	db, err := ConnectDb()
	if err != nil {
		//http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return models.Exec{}, utils.ErrorHandler(err, "error retrieving data")
	}
	defer db.Close()

	var exec models.Exec
	db.QueryRow("SELECT id, first_name, last_name, email, username, user_created_at, inactive_status, role FROM execs WHERE id = ?", id).Scan(
		&exec.ID, &exec.FirstName, &exec.LastName, &exec.Email, &exec.Username, &exec.UserCreatedAt, &exec.InactiveStatus, &exec.Role)
	if err == sql.ErrNoRows {
		//http.Error(w, "Teacher not found", http.StatusNotFound)
		return models.Exec{}, utils.ErrorHandler(err, "error retrieving data")
	} else if err != nil {
		fmt.Println(err)
		//http.Error(w, "Database query error", http.StatusInternalServerError)
		return models.Exec{}, utils.ErrorHandler(err, "error retrieving data")
	}
	return exec, nil
}

func AddExecDbHandler(newExecs []models.Exec) ([]models.Exec, error) {
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error retrieving data 1")
	}
	defer db.Close()

	//stmt, err := db.Prepare("INSERT INTO teachers (first_name, last_name, email, class, subject) VALUES (?,?,?,?,?)")
	stmt, err := db.Prepare(utils.GenerateInsertQuery("execs", models.Exec{}))
	fmt.Println(utils.GenerateInsertQuery("execs", models.Exec{}))
	if err != nil {
		return nil, utils.ErrorHandler(err, "error retrieving data 2")
	}
	defer stmt.Close()

	addedExecs := make([]models.Exec, len(newExecs))
	for i, newExec := range newExecs {
		newExec.Password, err = utils.HashPassword(newExec.Password)
		if err != nil {
			return nil, utils.ErrorHandler(err, "error adding exec into database")
		}

		//res, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.Class, newTeacher.Subject)
		values := utils.GetStructValues(newExec)
		res, err := stmt.Exec(values...)
		if err != nil {
			//http.Error(w, "Error inserting data into database", http.StatusInternalServerError)
			if strings.Contains(err.Error(), "a foreign key constraint fails") {
				return nil, utils.ErrorHandler(err, "class/class teacher does not exist")
			}
			return nil, utils.ErrorHandler(err, "error retrieving data")
		}
		lastID, err := res.LastInsertId()
		if err != nil {
			//http.Error(w, "Error getting last insert ID", http.StatusInternalServerError)
			return nil, utils.ErrorHandler(err, "error retrieving data")
		}
		newExec.ID = int(lastID)
		addedExecs[i] = newExec
	}
	return addedExecs, nil
}

func PatchExecs(updates []map[string]interface{}) error {
	db, err := ConnectDb()
	if err != nil {
		log.Println(err)
		//http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return utils.ErrorHandler(err, "error retrieving data")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		//http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		return utils.ErrorHandler(err, "error retrieving data")
	}

	for _, update := range updates {
		idStr, ok := update["id"].(string)
		if !ok {
			tx.Rollback()
			//http.Error(w, "Invalid teacher ID in update", http.StatusBadRequest)
			return utils.ErrorHandler(err, "error retrieving data")
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			tx.Rollback()
			//http.Error(w, "Invalid teacher ID in update", http.StatusBadRequest)
			return utils.ErrorHandler(err, "error retrieving data")
		}

		fmt.Println(id)

		var execFromDb models.Exec
		err = db.QueryRow("SELECT id, first_name, last_name, email, username FROM execs WHERE id = ?", id).Scan(&execFromDb.ID, &execFromDb.FirstName, &execFromDb.LastName, &execFromDb.Email, &execFromDb.Username)
		if err != nil {
			if err == sql.ErrNoRows {
				tx.Rollback()
				fmt.Println(err)
				//http.Error(w, "Teacher not found", http.StatusInternalServerError)
			}
			//http.Error(w, "Error retrieving teacher", http.StatusInternalServerError)
			return utils.ErrorHandler(err, "error retrieving data")
		}

		// Apply updates using reflection
		studentVal := reflect.ValueOf(&execFromDb).Elem()
		studentType := studentVal.Type()

		for k, v := range update {
			if k == "id" {
				continue // skip updating the id field
			}
			for i := 0; i < studentVal.NumField(); i++ {
				field := studentType.Field(i)
				if field.Tag.Get("json") == k+",omitempty" {
					fieldVal := studentVal.Field(i)
					if fieldVal.CanSet() {
						val := reflect.ValueOf(v)
						if val.Type().ConvertibleTo(fieldVal.Type()) {
							fieldVal.Set(val.Convert(fieldVal.Type()))
						} else {
							tx.Rollback()
							log.Printf("Cannot convert %v to %v", val.Type(), fieldVal.Type())
						}
					}
					break
				}
			}
		}

		_, err = tx.Exec("UPDATE execs SET first_name = ?, last_name = ?, email = ? WHERE id = ?", execFromDb.FirstName, execFromDb.LastName, execFromDb.Email, execFromDb.ID)
		if err != nil {
			fmt.Println(err)
			tx.Rollback()
			//http.Error(w, "Error updating teacher", http.StatusInternalServerError)
			return utils.ErrorHandler(err, "error retrieving data")
		}
	}

	err = tx.Commit()
	if err != nil {
		//http.Error(w, "Error commiting transaction", http.StatusInternalServerError)
		return utils.ErrorHandler(err, "error retrieving data")
	}
	return nil
}

func PatchOneExec(id int, updates map[string]interface{}) (models.Exec, error) {
	db, err := ConnectDb()
	if err != nil {
		log.Println(err)
		//http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return models.Exec{}, utils.ErrorHandler(err, "error patch data")
	}
	defer db.Close()

	var existingExec models.Exec
	err = db.QueryRow("SELECT id, first_name, last_name, email FROM execs WHERE id = ?", id).Scan(&existingExec.ID, &existingExec.FirstName, &existingExec.LastName, &existingExec.Email)

	if err != nil {
		if err == sql.ErrNoRows {
			//http.Error(w, "Teacher not found", http.StatusNotFound)
			return models.Exec{}, utils.ErrorHandler(err, "error patch data")
		}
		fmt.Println(err)
		//http.Error(w, "Unable to retrieve data", http.StatusInternalServerError)
		return models.Exec{}, utils.ErrorHandler(err, "error patch data")
	}

	// Apply updates using reflect
	studentVal := reflect.ValueOf(&existingExec).Elem()
	studentType := studentVal.Type()
	fmt.Println("Teacher Val:", studentVal.Type().Field(1))

	for k, v := range updates {
		for i := 0; i < studentVal.NumField(); i++ {
			fmt.Println("k from reflect mechanism", k)
			field := studentType.Field(i)
			fmt.Println(field.Tag.Get("json"))
			if field.Tag.Get("json") == k+",omitempty" {
				if studentVal.Field(i).CanSet() {
					fieldVal := studentVal.Field(i)
					fmt.Println("fieldVal:", fieldVal)
					fmt.Println("teacherVal.Field(i).Type():", studentVal.Field(i).Type())
					fmt.Println("reflect.ValueOf(v):", reflect.ValueOf(v))
					fieldVal.Set(reflect.ValueOf(v).Convert(studentVal.Field(i).Type()))
				}
			}
		}
	}

	_, err = db.Exec("UPDATE execs SET first_name = ?, last_name = ?, email = ? WHERE id = ?", existingExec.FirstName, existingExec.LastName, existingExec.Email, existingExec.ID)
	if err != nil {
		fmt.Println(err)
		//http.Error(w, "Error updating teacher", http.StatusInternalServerError)
		return models.Exec{}, utils.ErrorHandler(err, "error patch data")
	}
	return existingExec, err
}

func DeleteOneExec(id int) error {
	db, err := ConnectDb()
	if err != nil {
		log.Println(err)
		//http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return utils.ErrorHandler(err, "error delete data")
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM execs WHERE id = ?", id)
	if err != nil {
		//http.Error(w, "Error deleting teacher", http.StatusInternalServerError)
		return utils.ErrorHandler(err, "error delete data")
	}

	fmt.Println(result.RowsAffected())
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		//http.Error(w, "Error deleting teacher", http.StatusInternalServerError)
		return utils.ErrorHandler(err, "error delete data")
	}

	if rowsAffected == 0 {
		//http.Error(w, "Teacher not found", http.StatusNotFound)
		return utils.ErrorHandler(err, "error delete data")
	}
	return nil
}

func DeleteExecs(ids []int) ([]int, error) {
	db, err := ConnectDb()
	if err != nil {
		log.Println(err)
		//http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "error delete data")
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		//http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "error delete data")
	}

	stmt, err := tx.Prepare("DELETE FROM execs WHERE id = ?")
	if err != nil {
		log.Println(err)
		tx.Rollback()
		//http.Error(w, "Error preparing delete statement", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "error delete data")
	}
	defer stmt.Close()

	deletedIds := []int{}

	for _, id := range ids {
		result, err := stmt.Exec(id)
		if err != nil {
			tx.Rollback()
			log.Println(err)
			//http.Error(w, "Error deleting teacher", http.StatusInternalServerError)
			return nil, utils.ErrorHandler(err, "error delete data")
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			tx.Rollback()
			//http.Error(w, "Error retrieving deleted result", http.StatusInternalServerError)
			return nil, utils.ErrorHandler(err, "error delete data")
		}

		// if teacher was deleted then add the ID to the deletedIDs slice
		if rowsAffected > 0 {
			deletedIds = append(deletedIds, id)
		}

		if rowsAffected < 1 {
			tx.Rollback()
			//http.Error(w, fmt.Sprintf("ID does not exist", id), http.StatusInternalServerError)
			return nil, utils.ErrorHandler(err, "error delete data")
		}
	}

	// Commit
	err = tx.Commit()
	if err != nil {
		log.Println(err)
		//http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "error delete data")
	}

	if len(deletedIds) < 1 {
		//http.Error(w, "IDs do not exist", http.StatusBadRequest)
		return nil, utils.ErrorHandler(err, "error delete data")
	}
	return deletedIds, nil
}

func GetUserByUsername(username string) (*models.Exec, error) {
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error updating data")
	}
	defer db.Close()

	user := &models.Exec{}
	err = db.QueryRow("SELECT id, first_name, last_name, email, username, password, inactive_status, role FROM execs WHERE username = ?", username).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Username, &user.Password, &user.InactiveStatus, &user.Role)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, utils.ErrorHandler(err, "user not found")
		}
		return nil, utils.ErrorHandler(err, "user not found")
	}
	return user, nil
}

func UpdatePasswordInDb(userId int, currentPassword, newPassword string) (bool, error) {
	db, err := ConnectDb()
	if err != nil {
		return false, utils.ErrorHandler(err, "database connection error")
	}
	defer db.Close()

	var username string
	var userPassword string
	var userRole string

	err = db.QueryRow("SELECT username, password, role FROM execs WHERE id = ?", userId).Scan(&username, &userPassword, &userRole)
	if err != nil {
		return false, utils.ErrorHandler(err, "user not found")
	}

	err = utils.VerifyPassword(currentPassword, userPassword)
	if err != nil {
		return false, utils.ErrorHandler(err, "The password does not match")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return false, utils.ErrorHandler(err, "Internal error")
	}

	currentTime := time.Now().Format(time.RFC3339)
	_, err = db.Exec("UPDATE execs SET password = ?, password_changed_at = ? WHERE id = ?", hashedPassword, currentTime, userId)
	if err != nil {
		return false, utils.ErrorHandler(err, "failed to update the password")
	}

	//token, err := utils.SignToken(userId, username, userRole)
	//if err != nil {
	//	utils.ErrorHandler(err, "Could not create token")
	//	return
	//}

	return false, nil
}
