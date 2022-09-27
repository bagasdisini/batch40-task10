package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"personal-web/connection"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	route := mux.NewRouter()

	connection.DatabaseConnect()

	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))

	route.HandleFunc("/", home).Methods("GET")
	route.HandleFunc("/contact", contact).Methods("GET")
	route.HandleFunc("/detail-project/{id}", blogDetail).Methods("GET")
	route.HandleFunc("/add-project", formAddBlog).Methods("GET")
	route.HandleFunc("/add-blog", addBlog).Methods("POST")
	route.HandleFunc("/delete-project/{id}", deleteProject).Methods("GET")
	route.HandleFunc("/edit-project/{id}", editProject).Methods("GET")
	route.HandleFunc("/update-project/{id}", updateProject).Methods("POST")

	fmt.Println("Server berjalan di port 8080")

	http.ListenAndServe("localhost:8080", route)
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var template, error = template.ParseFiles("views/index.html")

	if error != nil {
		w.Write([]byte(error.Error()))
		return
	}

	data, _ := connection.Conn.Query(context.Background(), "SELECT name, description, technologies, duration, id FROM tb_blog")

	var result []Project
	for data.Next() {
		var each = Project{}

		err := data.Scan(&each.ProjectName, &each.Description, &each.Technologies, &each.Duration, &each.ID)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		result = append(result, each)
	}
	resData := map[string]interface{}{
		"Project": result,
	}

	w.WriteHeader(http.StatusOK)
	template.Execute(w, resData)
}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var template, error = template.ParseFiles("views/contact.html")

	if error != nil {
		w.Write([]byte(error.Error()))
		return
	}

	template.Execute(w, nil)
}

func blogDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var template, error = template.ParseFiles("views/detail-project.html")

	if error != nil {
		w.Write([]byte(error.Error()))
		return
	}

	var BlogDetail = Project{}

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	error = connection.Conn.QueryRow(context.Background(), "SELECT name, start_date, end_date, description, duration FROM tb_blog WHERE id=$1", id).Scan(&BlogDetail.ProjectName, &BlogDetail.StartDate, &BlogDetail.EndDate, &BlogDetail.Description, &BlogDetail.Duration)

	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(error.Error()))
	}

	data := map[string]interface{}{
		"BlogDetail": BlogDetail,
	}

	template.Execute(w, data)
}

func formAddBlog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var template, error = template.ParseFiles("views/add-project.html")

	if error != nil {
		w.Write([]byte(error.Error()))
		return
	}

	template.Execute(w, nil)
}

type Project struct {
	ID           int
	ProjectName  string
	Description  string
	StartDate    string
	EndDate      string
	Technologies []string
	Duration     string
}

func addBlog(w http.ResponseWriter, r *http.Request) {
	error := r.ParseForm()
	if error != nil {
		log.Fatal(error)
	}

	var duration string
	var projectName = r.PostForm.Get("projectName")
	var deskripsi = r.PostForm.Get("deskripsi")
	var startDate = r.PostForm.Get("startDate")
	var endDate = r.PostForm.Get("endDate")
	var tech = r.Form["checkbox"]

	var layout = "2006-01-02"
	var startDateParse, _ = time.Parse(layout, startDate)
	var endDateParse, _ = time.Parse(layout, endDate)
	var startDateConvert = startDateParse.Format("02 Jan 2006")
	var endDateConvert = endDateParse.Format("02 Jan 2006")

	var hours = endDateParse.Sub(startDateParse).Hours()
	var days = hours / 24
	var weeks = math.Round(days / 7)
	var months = math.Round(days / 30)
	var years = math.Round(days / 365)

	if days >= 1 && days <= 6 {
		duration = strconv.Itoa(int(days)) + " day(s)"
	} else if days >= 7 && days <= 29 {
		duration = strconv.Itoa(int(weeks)) + " week(s)"
	} else if days >= 30 && days <= 364 {
		duration = strconv.Itoa(int(months)) + " month(s)"
	} else if days >= 365 {
		duration = strconv.Itoa(int(years)) + " year(s)"
	}

	_, error = connection.Conn.Exec(context.Background(), "INSERT INTO public.tb_blog(name, start_date, end_date, description, technologies, duration) VALUES ($1, $2, $3, $4, $5, $6)", projectName, startDateConvert, endDateConvert, deskripsi, tech, duration)

	if error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(error.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func deleteProject(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	_, err := connection.Conn.Exec(context.Background(), "DELETE FROM tb_blog WHERE id=$1", id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func editProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	var tmpl, err = template.ParseFiles("views/edit-project.html")

	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	var editProject = Project{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT id, name, description FROM tb_blog WHERE id = $1", id).Scan(&editProject.ID, &editProject.ProjectName, &editProject.Description)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("message : " + err.Error()))
		return
	}

	data := map[string]interface{}{
		"EditProject": editProject,
	}

	tmpl.Execute(w, data)
}

func updateProject(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}

	var durationUpdate string
	var projectNameUpdate = r.PostForm.Get("projectName")
	var deskripsiUpdate = r.PostForm.Get("deskripsi")
	var startDateUpdate = r.PostForm.Get("startDate")
	var endDateUpdate = r.PostForm.Get("endDate")
	var techUpdate = r.Form["checkbox"]

	var layoutUpdate = "2006-01-02"
	var startDateParseUpdate, _ = time.Parse(layoutUpdate, startDateUpdate)
	var endDateParseUpdate, _ = time.Parse(layoutUpdate, endDateUpdate)
	var startDateConvert = startDateParseUpdate.Format("02 Jan 2006")
	var endDateConvertUpdate = endDateParseUpdate.Format("02 Jan 2006")

	var hoursUpdate = endDateParseUpdate.Sub(startDateParseUpdate).Hours()
	var daysUpdate = hoursUpdate / 24
	var weeksUpdate = math.Round(daysUpdate / 7)
	var monthsUpdate = math.Round(daysUpdate / 30)
	var yearsUpdate = math.Round(daysUpdate / 365)

	if daysUpdate >= 1 && daysUpdate <= 6 {
		durationUpdate = strconv.Itoa(int(daysUpdate)) + " day(s)"
	} else if daysUpdate >= 7 && daysUpdate <= 29 {
		durationUpdate = strconv.Itoa(int(weeksUpdate)) + " week(s)"
	} else if daysUpdate >= 30 && daysUpdate <= 364 {
		durationUpdate = strconv.Itoa(int(monthsUpdate)) + " month(s)"
	} else if daysUpdate >= 365 {
		durationUpdate = strconv.Itoa(int(yearsUpdate)) + " year(s)"
	}

	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	sqlStatement := `
	UPDATE public.tb_blog
	SET name=$7, description=$4, technologies=$5, duration=$6, start_date=$2, end_date=$3
	WHERE id=$1;
	`

	_, err = connection.Conn.Exec(context.Background(), sqlStatement, id, startDateConvert, endDateConvertUpdate, deskripsiUpdate, techUpdate, durationUpdate, projectNameUpdate)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}
