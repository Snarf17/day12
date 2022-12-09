package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"personal-web/connection"
	"personal-web/middleware"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	route := mux.NewRouter()

	connection.DatabaseConnect()

	// 	Route inisialisasi Folder Public and Upload
	route.PathPrefix("/public/").Handler(http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))
	route.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	route.HandleFunc("/", home).Methods("GET")
	route.HandleFunc("/contact", contact).Methods("GET")
	route.HandleFunc("/project-detail/{id}", projectdetail).Methods("GET")
	route.HandleFunc("/add-article", middleware.UploadFile(formAddArticle)).Methods("POST")
	route.HandleFunc("/add-article", addArticle).Methods("GET")
	route.HandleFunc("/delete-article/{id}", deleteArticle).Methods("GET")
	route.HandleFunc("/edit-article/{index}", editFormArticle).Methods("GET")
	route.HandleFunc("/edit-article/{index}", middleware.UploadFile(editArticle)).Methods("POST")

	// Route Authentication Login and Register
	route.HandleFunc("/register", FormRegister).Methods("GET")
	route.HandleFunc("/register", register).Methods("POST")

	route.HandleFunc("/login", formLogin).Methods("GET")
	route.HandleFunc("/login", login).Methods("POST")

	route.HandleFunc("/logout", logout).Methods("GET")

	fmt.Println("Server On Running in port 5000")
	http.ListenAndServe("localhost:5000", route)
}

type MetaData struct {
	Id        int
	Title     string
	IsLogin   bool
	UserName  string
	FlashData string
}

var Data = MetaData{
	Title: "Personal Web",
}

// Var Blog
type User struct {
	Id       int
	Name     string
	Email    string
	Password string
}

type Project struct {
	Id               int
	NameProject      string
	Creator          string
	StarDate         time.Time
	StarDateEdit     string
	EndDate          time.Time
	EndDateEdit      string
	Duration         string
	Format_Star_Date string
	Format_End_Date  string
	Desc             string
	Image            string
	Tech             []string
	NodeJs           string
	ReactJs          string
	NextJs           string
	Typescript       string
	IsLogin          bool
	NodeCheked       string
	CobaCheked       string
	TryCheked        string
	TypescriptCheked string
}
type Login struct {
	Er           string
	InputInvalid string
}

var DataErr = Login{}

// var projects = []Project{}

func formAddArticle(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	name_project := r.PostForm.Get("name")
	start_date := r.PostForm.Get("starDate")
	end_date := r.PostForm.Get("endDate")
	description := r.PostForm.Get("desc")

	dataContext := r.Context().Value("dataImages")
	image := dataContext.(string)

	nodeJs := r.PostForm.Get("node")
	nextJs := r.PostForm.Get("next")
	reactJs := r.PostForm.Get("react")
	typescript := r.PostForm.Get("typescript")

	checked := []string{
		nodeJs,
		nextJs,
		reactJs,
		typescript,
	}

	var store = sessions.NewCookieStore([]byte("SESSIONS_ID"))
	session, _ := store.Get(r, "SESSIONS_ID")

	userPost := session.Values["Id"]

	// fmt.Println(userPost)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_project(name_project, start_date, end_date, description, technologies, image, user_id) VALUES ($1, $2, $3, $4, $5, $6, $7)", name_project, start_date, end_date, description, checked, image, userPost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Massage : " + err.Error()))
		return
	}
	// projects = append(projects, newProjects)
	// fmt.Println(nodeChecked)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func addArticle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text-html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/addproject.html")

	if err != nil {
		w.Write([]byte("Massage : " + err.Error()))
		return
	}
	// Setting SESSION LOGIN
	var store = sessions.NewCookieStore([]byte("SESSIONS_ID"))
	session, _ := store.Get(r, "SESSIONS_ID")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Names"].(string)
		Data.Id = session.Values["Id"].(int)
	}

	fm := session.Flashes("massage")

	var flashes []string
	if len(fm) > 0 {
		session.Save(r, w)

		for _, fl := range fm {
			flashes = append(flashes, fl.(string))
		}
	}

	Data.FlashData = strings.Join(flashes, "")
	Data := map[string]interface{}{
		"FlashData": Data,
	}

	tmpt.Execute(w, Data)
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text-html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/index.html")

	if err != nil {
		w.Write([]byte("Massage : " + err.Error()))
		return
	}

	// Setting SESSION LOGIN
	var store = sessions.NewCookieStore([]byte("SESSIONS_ID"))
	session, _ := store.Get(r, "SESSIONS_ID")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Names"].(string)
		Data.Id = session.Values["Id"].(int)
	}

	fm := session.Flashes("massage")

	var flashes []string
	if len(fm) > 0 {
		session.Save(r, w)

		for _, fl := range fm {
			flashes = append(flashes, fl.(string))
		}
	}

	Data.FlashData = strings.Join(flashes, "")

	dataProject, errQuery := connection.Conn.Query(context.Background(), "SELECT tb_project.id, name_project, start_date, end_date, description, technologies, image, tb_user.name as creator FROM tb_project LEFT JOIN tb_user ON tb_project.user_id = tb_user.id ORDER BY id DESC")
	if errQuery != nil {
		// w.Write([]byte("Massage : " + errQuery.Error()))
		fmt.Println("Massage : " + errQuery.Error())
		return
	}

	var result = []Project{}

	for dataProject.Next() {
		var each = Project{}

		err := dataProject.Scan(&each.Id, &each.NameProject, &each.StarDate, &each.EndDate, &each.Desc, &each.Tech, &each.Image, &each.Creator)
		if err != nil {
			// w.Write([]byte("Massage : " + err.Error()))
			fmt.Println("Massage : " + err.Error())
			return
		}
		diff := each.EndDate.Sub(each.StarDate)
		days := diff.Hours() / 24
		month := math.Floor(diff.Hours() / 24 / 30)

		dy := strconv.FormatFloat(days, 'f', 0, 64)
		mo := strconv.FormatFloat(month, 'f', 0, 64)

		if days < 30 {
			each.Duration = dy + " Days"
		} else if days > 30 {
			each.Duration = mo + " Month"
		}

		if each.Tech[0] == "nodejs" {
			each.NodeJs = "nodejs.svg"
		} else {
			each.NodeJs = "d-none"
		}
		if each.Tech[1] == "nextjs" {
			each.NextJs = "nextjs.svg"
		} else {
			each.NextJs = "d-none"
		}
		if each.Tech[2] == "reactjs" {
			each.ReactJs = "reactjs.svg"
		} else {
			each.ReactJs = "d-none"
		}
		if each.Tech[3] == "typescript" {
			each.Typescript = "typescript.svg"
		} else {
			each.Typescript = "d-none"
		}

		each.Format_Star_Date = each.StarDate.Format("2 January 2006")
		each.Format_End_Date = each.EndDate.Format("2 January 2006")
		// result = append(result, each)

		if session.Values["IsLogin"] != true {
			each.IsLogin = false
		} else {
			each.IsLogin = session.Values["IsLogin"].(bool)
		}

		result = append(result, each)
	}

	// fmt.Println(result)
	data := map[string]interface{}{
		"Projects": result,
		"Data":     Data,
	}

	tmpt.Execute(w, data)
}

func deleteArticle(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"])

	// projects = append(projects[:index], projects[index+1:]...)
	_, err := connection.Conn.Exec(context.Background(), "DELETE FROM tb_project WHERE id=$1", id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Massage : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func projectdetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text-html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/project-detail.html")

	if err != nil {
		w.Write([]byte("Massage : " + err.Error()))
		return
	}

	id, _ := strconv.Atoi(mux.Vars(r)["id"])
	// Setting SESSION LOGIN
	var store = sessions.NewCookieStore([]byte("SESSIONS_ID"))
	session, _ := store.Get(r, "SESSIONS_ID")

	if session.Values["IsLogin"] != true {
		Data.IsLogin = false
	} else {
		Data.IsLogin = session.Values["IsLogin"].(bool)
		Data.UserName = session.Values["Names"].(string)
		Data.Id = session.Values["Id"].(int)
	}

	fm := session.Flashes("massage")

	var flashes []string
	if len(fm) > 0 {
		session.Save(r, w)

		for _, fl := range fm {
			flashes = append(flashes, fl.(string))
		}
	}

	Data.FlashData = strings.Join(flashes, "")

	// Ojebct in golang
	// var result = []Project{}
	var ProjectDetail = Project{}

	// for index, data := range projects {
	// 	if index == id {
	// 		ProjectDetail = Project{
	// 			NameProject: data.NameProject,
	// 			StarDate:    data.StarDate,
	// 			EndDate:     data.EndDate,
	// 			Desc:        data.Desc,
	// 		}
	// 	}
	// }

	err = connection.Conn.QueryRow(context.Background(), "SELECT id, name_project, start_date, end_date, technologies, description,image FROM tb_project WHERE id=$1", id).Scan(
		&ProjectDetail.Id, &ProjectDetail.NameProject, &ProjectDetail.StarDate, &ProjectDetail.EndDate, &ProjectDetail.Tech, &ProjectDetail.Desc, &ProjectDetail.Image,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("massage : " + err.Error()))
		return
	}
	diff := ProjectDetail.EndDate.Sub(ProjectDetail.StarDate)
	days := diff.Hours() / 24
	month := math.Floor(diff.Hours() / 24 / 30)

	dy := strconv.FormatFloat(days, 'f', 0, 64)
	mo := strconv.FormatFloat(month, 'f', 0, 64)

	if days < 30 {
		ProjectDetail.Duration = dy + " Days"
	} else if days > 30 {
		ProjectDetail.Duration = mo + " Month"
	}

	if ProjectDetail.Tech[0] == "nodejs" {
		ProjectDetail.NodeJs = "nodejs.svg"
	} else {
		ProjectDetail.NodeJs = "d-none"
	}
	if ProjectDetail.Tech[1] == "nextjs" {
		ProjectDetail.NextJs = "nextjs.svg"
	} else {
		ProjectDetail.NextJs = "d-none"
	}
	if ProjectDetail.Tech[2] == "reactjs" {
		ProjectDetail.ReactJs = "reactjs.svg"
	} else {
		ProjectDetail.ReactJs = "d-none"
	}
	if ProjectDetail.Tech[3] == "typescript" {
		ProjectDetail.Typescript = "typescript.svg"
	} else {
		ProjectDetail.Typescript = "d-none"
	}

	ProjectDetail.Format_Star_Date = ProjectDetail.StarDate.Format("2 January 2006")
	ProjectDetail.Format_End_Date = ProjectDetail.EndDate.Format("2 January 2006")

	if session.Values["IsLogin"] != true {
		ProjectDetail.IsLogin = false
	} else {
		ProjectDetail.IsLogin = session.Values["IsLogin"].(bool)
	}

	// result = append(result, ProjectDetail)

	fmt.Println(ProjectDetail)

	dataDetail := map[string]interface{}{
		"Project": ProjectDetail,
		"Data":    Data,
	}

	tmpt.Execute(w, dataDetail)
}

func editFormArticle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text-html ; charset=utf-8")
	tmpt, err := template.ParseFiles("views/edit.html")

	if err != nil {
		w.Write([]byte("Massage : " + err.Error()))
		return
	}
	index, _ := strconv.Atoi(mux.Vars(r)["index"])

	// Ojebct in golang
	var EditArticle = Project{}
	var Check = Project{
		NodeCheked:       "",
		CobaCheked:       "",
		TryCheked:        "",
		TypescriptCheked: "",
		// NodeIcon:         "",
		// ReactIcon:        "",
		// NextIcon:         "",
		// TypescriptIcon:   "",
	}

	// fmt.Println(Check)
	err = connection.Conn.QueryRow(context.Background(), "SELECT id,name_project,start_date,end_date,description, technologies, image FROM tb_project WHERE id=$1", index).Scan(&EditArticle.Id, &EditArticle.NameProject, &EditArticle.StarDate, &EditArticle.EndDate, &EditArticle.Desc, &EditArticle.Tech, &EditArticle.Image)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Massage : " + err.Error()))
		return
	}

	if EditArticle.Tech[0] == "nodejs" {
		Check.NodeCheked = "checked"

	} else {
		Check.NodeCheked = ""
	}
	if EditArticle.Tech[2] == "reactjs" {
		Check.CobaCheked = "checked"
	} else {
		Check.CobaCheked = ""
	}
	if EditArticle.Tech[1] == "nextjs" {
		Check.TryCheked = "checked"
	} else {
		Check.TryCheked = ""
	}
	if EditArticle.Tech[3] == "typescript" {
		Check.TypescriptCheked = "checked"
	} else {
		Check.TypescriptCheked = ""
	}

	EditArticle.StarDateEdit = EditArticle.StarDate.Format("2006-01-02")
	EditArticle.EndDateEdit = EditArticle.EndDate.Format("2006-01-02")

	editProject := map[string]interface{}{
		"Project": EditArticle,
		"Check":   Check,
	}
	// fmt.Println(EditArticle)
	tmpt.Execute(w, editProject)

}
func editArticle(w http.ResponseWriter, r *http.Request) {
	index, _ := strconv.Atoi(mux.Vars(r)["index"])
	err := r.ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	name_project := r.PostForm.Get("name")
	start_date := r.PostForm.Get("starDate")
	end_date := r.PostForm.Get("endDate")
	description := r.PostForm.Get("desc")

	dataContext := r.Context().Value("dataImages")
	image := dataContext.(string)

	nodeJs := r.PostForm.Get("node")
	nextJs := r.PostForm.Get("next")
	reactJs := r.PostForm.Get("react")
	typescript := r.PostForm.Get("typescript")

	checked := []string{
		nodeJs,
		nextJs,
		reactJs,
		typescript,
	}

	_, errUpdate := connection.Conn.Exec(context.Background(), "UPDATE tb_project SET name_project=$1, start_date=$2, end_date=$3, technologies=$4, description=$5, image=$6 WHERE id=$7", name_project, start_date, end_date, checked, description, image, index)

	if errUpdate != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Massage : " + errUpdate.Error()))
		return
	}

	http.Redirect(w, r, "/", http.StatusMovedPermanently)
}

func contact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text-html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/contact.html")

	if err != nil {
		w.Write([]byte("Massage : " + err.Error()))
		return
	}

	tmpt.Execute(w, nil)
}

func FormRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/register.html")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Massage : " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	tmpt.Execute(w, nil)

}

func register(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	name := r.PostForm.Get("name")
	email := r.PostForm.Get("email")

	password := r.PostForm.Get("password")
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte(password), 10)

	_, err = connection.Conn.Exec(context.Background(), "INSERT INTO tb_user(name, email,password) VALUES ($1, $2, $3)", name, email, passwordHash)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Massage : " + err.Error()))
		return
	}

	http.Redirect(w, r, "/login", http.StatusMovedPermanently)
}

func formLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpt, err := template.ParseFiles("views/login.html")

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Massage : " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	tmpt.Execute(w, nil)
}

func login(w http.ResponseWriter, r *http.Request) {
	var store = sessions.NewCookieStore([]byte("SESSIONS_ID"))
	session, _ := store.Get(r, "SESSIONS_ID")

	err := r.ParseForm()

	if err != nil {
		log.Fatal(err)
	}

	email := r.PostForm.Get("email")
	password := r.PostForm.Get("password")

	user := User{}

	err = connection.Conn.QueryRow(context.Background(), "SELECT * FROM tb_user WHERE email = $1", email).Scan(
		&user.Id, &user.Name, &user.Email, &user.Password,
	)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Message : " + err.Error()))
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		DataErr.Er = "Password Wrong"
		DataErr.InputInvalid = "is-invalid"
		t := template.Must(template.ParseFiles("views/login.html"))
		Data := map[string]interface{}{
			"DataErr": DataErr,
		}
		t.Execute(w, Data)
	}
	// fmt.Println(user)
	session.Values["IsLogin"] = true
	session.Values["Names"] = user.Name
	session.Values["Id"] = user.Id
	session.Options.MaxAge = 10800 // 3 Jam / Hours

	session.AddFlash("Login Succesfuly!", "massage")
	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusMovedPermanently)

}
func logout(w http.ResponseWriter, r *http.Request) {
	var store = sessions.NewCookieStore([]byte("SESSIONS_ID"))
	session, _ := store.Get(r, "SESSIONS_ID")
	session.Options.MaxAge = -1

	session.Save(r, w)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
