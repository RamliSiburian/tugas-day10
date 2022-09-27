package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"personal-web/connection"
	"strconv"
	"text/template"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	// membuat route
	route := mux.NewRouter()

	// memanggil package connection
	connection.DatabaseConnect()

	// Membuat route path folder public
	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	// membuat routing ke setiap halaman yang akan ditampilkan (html)
	route.HandleFunc("/", index).Methods("GET")                                        // index.html
	route.HandleFunc("/form-project", formAddProject).Methods("GET")                   // addproject.html
	route.HandleFunc("/add-project", addProject).Methods("POST")                       // menyimpan data dari form project ke db
	route.HandleFunc("/detail-project/{id_project}", detailProject).Methods("GET")     // menampilkan halaman detail project
	route.HandleFunc("/form-editproject/{id_project}", formEditProject).Methods("GET") // menampilkan form editproject.html
	route.HandleFunc("/edit-project/{id_project}", editProject).Methods("POST")        // Menjalankan fungsi edit
	route.HandleFunc("/delete-project/{id_project}", deleteProject).Methods("GET")     // menjalankan fungsi delete

	// menjalankan server (port opsional)
	fmt.Println("Server running on port 5050")
	http.ListenAndServe("localhost:5050", route)

}

// untuk membuat stuct yang mendefenisikan tipe data yang akan ditampilka (dto = data transformation object)
type Project struct {
	ID               int
	ProjectName      string
	StartDate        time.Time
	EndDate          time.Time
	Format_startdate string
	Format_enddate   string
	Description      string
	Technologies     string
	Duration         string
}

// function index / home
func index(w http.ResponseWriter, r *http.Request) {
	// membuat header/type html
	w.Header().Set("Description-Type", "text/html; charset=utf-8")

	// memanggil index.html dari folder views
	indeksTemplate, err := template.ParseFiles("views/index.html")

	if err != nil {
		// []byte untuk memberitahu bahwa data yang dikirim adalah tipe string
		w.Write([]byte("message : " + err.Error()))
		return
	}

	// mengambil data dari database
	queryData, _ := connection.Conn.Query(context.Background(), "SELECT id_project, project_name, start_date, end_date, description FROM tb_project ")

	// untuk menampung data dari database yang disiman di struct Project
	var resultData []Project
	for queryData.Next() {
		newData := Project{}
		// men scan data dari struck Project "urutan data pada scan harus sesuai dengan query data"
		err := queryData.Scan(&newData.ID, &newData.ProjectName, &newData.StartDate, &newData.EndDate, &newData.Description)

		if err != nil {
			fmt.Println(err.Error())
			return
		}
		// data tanggal bentuk string
		newData.Format_startdate = newData.StartDate.Format("2006-01-02")
		newData.Format_enddate = newData.EndDate.Format("2006-01-02")
		startDateFormat := newData.StartDate.Format("2006-01-02")
		endDateFormat := newData.EndDate.Format("2006-01-02")
		layout := "2006-01-02"
		// data tanggal betuk time/date
		startDateParse, _ := time.Parse(layout, startDateFormat)
		endDateParse, _ := time.Parse(layout, endDateFormat)

		hours := endDateParse.Sub(startDateParse).Hours()
		days := hours / 24
		weeks := math.Round(days / 7)
		months := math.Round(days / 30)
		years := math.Round(days / 365)

		var duration string

		if days >= 1 && days <= 6 {
			duration = strconv.Itoa(int(days)) + " days"
		} else if days >= 7 && days <= 29 {
			duration = strconv.Itoa(int(weeks)) + " weeks"
		} else if days >= 30 && days <= 364 {
			duration = strconv.Itoa(int(months)) + " months"
		} else if days >= 365 {
			duration = strconv.Itoa(int(years)) + " years"
		}

		newData.Duration = duration

		// mem push data ke resultData
		resultData = append(resultData, newData)
	}

	// menalpilkan data ke html
	data := map[string]interface{}{
		"Projects": resultData,
	}

	// menampilkan index.html
	indeksTemplate.Execute(w, data)
}

// function form add project
func formAddProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Description-Type", "text/html; charset=utf-8")

	formAdd, err := template.ParseFiles("views/addproject.html")

	if err != nil {
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	formAdd.Execute(w, nil)
}

// funtion add project
func addProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	// mengambil / menangkap data yang di input dari form
	projectName := r.PostForm.Get("projectName")
	startDate := r.PostForm.Get("startDate")
	endDate := r.PostForm.Get("endDate")
	description := r.PostForm.Get("description")

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_project(project_name , start_date , end_date , description) VALUES ($1, $2, $3, $4) ", projectName, startDate, endDate, description)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)

}

// funtion detail project
func detailProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	detailProjectTemplate, err := template.ParseFiles("views/detailproject.html")

	if err != nil {
		w.Write([]byte("message : " + err.Error()))
		return
	}

	var DetailProject = Project{}

	id_project, _ := strconv.Atoi(mux.Vars(r)["id_project"])

	err = connection.Conn.QueryRow(context.Background(), "SELECT id_project, project_name, start_date, end_date, description FROM tb_project WHERE id_project = $1", id_project).Scan(&DetailProject.ID, &DetailProject.ProjectName, &DetailProject.StartDate, &DetailProject.EndDate, &DetailProject.Description)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	DetailProject.Format_startdate = DetailProject.StartDate.Format("2006-01-02")
	DetailProject.Format_enddate = DetailProject.EndDate.Format("2006-01-02")
	startDateFormat := DetailProject.StartDate.Format("2006-01-02")
	endDateFormat := DetailProject.EndDate.Format("2006-01-02")
	layout := "2006-01-02"
	startDateParse, _ := time.Parse(layout, startDateFormat)
	endDateParse, _ := time.Parse(layout, endDateFormat)

	hours := endDateParse.Sub(startDateParse).Hours()
	days := hours / 24
	weeks := math.Round(days / 7)
	months := math.Round(days / 30)
	years := math.Round(days / 365)

	var duration string

	if days >= 1 && days <= 6 {
		duration = strconv.Itoa(int(days)) + " days"
	} else if days >= 7 && days <= 29 {
		duration = strconv.Itoa(int(weeks)) + " weeks"
	} else if days >= 30 && days <= 364 {
		duration = strconv.Itoa(int(months)) + " months"
	} else if days >= 365 {
		duration = strconv.Itoa(int(years)) + " years"
	}

	DetailProject.Duration = duration

	data := map[string]interface{}{
		"Project": DetailProject,
	}

	detailProjectTemplate.Execute(w, data)
}

// form edit project
func formEditProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Description-Type", "text/html; charset=utf-8")
	formEditTemplate, err := template.ParseFiles("views/editproject.html")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Message :" + err.Error()))
		return
	}

	id_project, _ := strconv.Atoi(mux.Vars(r)["id_project"])

	EditProject := Project{}
	err = connection.Conn.QueryRow(context.Background(), "SELECT id_project, project_name, start_date, end_date, description FROM tb_project WHERE id_project = $1", id_project).Scan(&EditProject.ID, &EditProject.ProjectName, &EditProject.StartDate, &EditProject.EndDate, &EditProject.Description)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	EditProject.Format_startdate = EditProject.StartDate.Format("2006-01-02")
	EditProject.Format_enddate = EditProject.EndDate.Format("2006-01-02")

	data := map[string]interface{}{
		"Edits": EditProject,
	}

	formEditTemplate.Execute(w, data)
}

// edit project
func editProject(w http.ResponseWriter, r *http.Request) {
	id_project, _ := strconv.Atoi(mux.Vars(r)["id_project"])

	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	// mengambil / menangkap data yang di input dari form
	projectNameEdit := r.PostForm.Get("projectName")
	startDateEdit := r.PostForm.Get("startDate")
	endDateEdit := r.PostForm.Get("endDate")
	descriptionEdit := r.PostForm.Get("description")

	_, err = connection.Conn.Exec(context.Background(), "UPDATE tb_project SET project_name = $1, start_date = $2, end_date = $3, description = $4 WHERE id_project = $5", projectNameEdit, startDateEdit, endDateEdit, descriptionEdit, id_project)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}
	fmt.Println(startDateEdit)
	fmt.Println(projectNameEdit)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)

}

// funtion delete project
func deleteProject(w http.ResponseWriter, r *http.Request) {
	id_project, _ := strconv.Atoi(mux.Vars(r)["id_project"])

	_, err := connection.Conn.Exec(context.Background(), "DELETE FROM tb_project WHERE id_project=$1", id_project)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}
