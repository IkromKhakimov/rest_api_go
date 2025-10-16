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
)

func GetStudentsDbHandler(students []models.Student, r *http.Request) ([]models.Student, error) {
	db, err := ConnectDb()
	if err != nil {
		// http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return nil, utils.ErrorHandler(err, "Error connecting to database")
	}
	defer db.Close()

	query := "SELECT id, first_name, last_name, email, class FROM students WHERE 1=1"
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
		var student models.Student
		err := rows.Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class)
		if err != nil {
			return nil, utils.ErrorHandler(err, "error retrieving data")
		}
		students = append(students, student)
	}
	return students, nil
}

func GetStudentByID(id int) (models.Student, error) {
	db, err := ConnectDb()
	if err != nil {
		//http.Error(w, "Error connecting to database", http.StatusInternalServerError)
		return models.Student{}, utils.ErrorHandler(err, "error retrieving data")
	}
	defer db.Close()

	var student models.Student
	db.QueryRow("SELECT id, first_name, last_name, class FROM students WHERE id = ?", id).Scan(&student.ID, &student.FirstName, &student.LastName, &student.Email, &student.Class)
	if err == sql.ErrNoRows {
		//http.Error(w, "Teacher not found", http.StatusNotFound)
		return models.Student{}, utils.ErrorHandler(err, "error retrieving data")
	} else if err != nil {
		fmt.Println(err)
		//http.Error(w, "Database query error", http.StatusInternalServerError)
		return models.Student{}, utils.ErrorHandler(err, "error retrieving data")
	}
	return student, nil
}

func AddStudentDbHandler(newStudents []models.Student) ([]models.Student, error) {
	db, err := ConnectDb()
	if err != nil {
		return nil, utils.ErrorHandler(err, "error retrieving data 1")
	}
	defer db.Close()

	//stmt, err := db.Prepare("INSERT INTO teachers (first_name, last_name, email, class, subject) VALUES (?,?,?,?,?)")
	stmt, err := db.Prepare(utils.GenerateInsertQuery("students", models.Student{}))
	fmt.Println(utils.GenerateInsertQuery("students", models.Student{}))
	if err != nil {
		return nil, utils.ErrorHandler(err, "error retrieving data 2")
	}
	defer stmt.Close()

	addedStudents := make([]models.Student, len(newStudents))
	for i, newStudent := range newStudents {
		//res, err := stmt.Exec(newTeacher.FirstName, newTeacher.LastName, newTeacher.Email, newTeacher.Class, newTeacher.Subject)
		values := utils.GetStructValues(newStudent)
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
		newStudent.ID = int(lastID)
		addedStudents[i] = newStudent
	}
	return addedStudents, nil
}

func UpdateStudent(id int, updatedStudent models.Student) (models.Student, error) {
	db, err := ConnectDb()
	if err != nil {
		log.Println(err)
		//http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return models.Student{}, utils.ErrorHandler(err, "error retrieving data")
	}
	defer db.Close()

	var existingStudent models.Student
	err = db.QueryRow("SELECT id, first_name, last_name, email, class FROM students WHERE id = ?", id).Scan(&existingStudent.ID, &existingStudent.FirstName, &existingStudent.LastName, &existingStudent.Email, &existingStudent.Class)

	if err != nil {
		if err == sql.ErrNoRows {
			//http.Error(w, "Teacher not found", http.StatusNotFound)
			return models.Student{}, utils.ErrorHandler(err, "error retrieving data")
		}
		fmt.Println(err)
		//http.Error(w, "Unable to retrieve data", http.StatusInternalServerError)
		return models.Student{}, utils.ErrorHandler(err, "error retrieving data")
	}

	updatedStudent.ID = existingStudent.ID
	_, err = db.Exec("UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ? WHERE id = ?", updatedStudent.FirstName, updatedStudent.LastName, updatedStudent.Email, updatedStudent.Class, updatedStudent.ID)
	if err != nil {
		fmt.Println(err)
		//http.Error(w, "Error updating teacher", http.StatusInternalServerError)
		return models.Student{}, utils.ErrorHandler(err, "error retrieving data")
	}
	return models.Student{}, nil
}

func PatchStudent(updates []map[string]interface{}) error {
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

		var studentFromDb models.Student
		err = db.QueryRow("SELECT id, first_name, last_name, email, class FROM students WHERE id = ?", id).Scan(&studentFromDb.ID, &studentFromDb.FirstName, &studentFromDb.LastName, &studentFromDb.Email, &studentFromDb.Class)
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
		studentVal := reflect.ValueOf(&studentFromDb).Elem()
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

		_, err = tx.Exec("UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ? WHERE id = ?", studentFromDb.FirstName, studentFromDb.LastName, studentFromDb.Email, studentFromDb.Class, studentFromDb.ID)
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

func PatchOneStudent(id int, updates map[string]interface{}) (models.Student, error) {
	db, err := ConnectDb()
	if err != nil {
		log.Println(err)
		//http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return models.Student{}, utils.ErrorHandler(err, "error patch data")
	}
	defer db.Close()

	var existingStudent models.Student
	err = db.QueryRow("SELECT id, first_name, last_name, email, class FROM students WHERE id = ?", id).Scan(&existingStudent.ID, &existingStudent.FirstName, &existingStudent.LastName, &existingStudent.Email, &existingStudent.Class)

	if err != nil {
		if err == sql.ErrNoRows {
			//http.Error(w, "Teacher not found", http.StatusNotFound)
			return models.Student{}, utils.ErrorHandler(err, "error patch data")
		}
		fmt.Println(err)
		//http.Error(w, "Unable to retrieve data", http.StatusInternalServerError)
		return models.Student{}, utils.ErrorHandler(err, "error patch data")
	}

	// Apply updates using reflect
	studentVal := reflect.ValueOf(&existingStudent).Elem()
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

	_, err = db.Exec("UPDATE students SET first_name = ?, last_name = ?, email = ?, class = ? WHERE id = ?", existingStudent.FirstName, existingStudent.LastName, existingStudent.Email, existingStudent.Class, existingStudent.ID)
	if err != nil {
		fmt.Println(err)
		//http.Error(w, "Error updating teacher", http.StatusInternalServerError)
		return models.Student{}, utils.ErrorHandler(err, "error patch data")
	}
	return existingStudent, err
}

func DeleteOneStudent(id int) error {
	db, err := ConnectDb()
	if err != nil {
		log.Println(err)
		//http.Error(w, "Unable to connect to database", http.StatusInternalServerError)
		return utils.ErrorHandler(err, "error delete data")
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM students WHERE id = ?", id)
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

func DeleteStudents(ids []int) ([]int, error) {
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

	stmt, err := tx.Prepare("DELETE FROM students WHERE id = ?")
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
