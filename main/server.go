package main

import (
	"crypto/md5"
	"database/sql" // Interactuación con la base de datos
	"encoding/json"
	"strconv"
	"strings"

	"fmt" // Imprimir en consola
	"io"

	"math/rand" //aleatorios
	// Ayuda a escribir en la respuesta
	"net/http" // El paquete HTTP

	_ "github.com/go-sql-driver/mysql" // La librería para mySQL
)

type User struct {
	nickname, full_name, email, pass, fav_airport string
	id                                            int
}

type Airport struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
	IATA     string `json:"IATA"`
}

type Route struct {
	Id          int     `json:"id"`
	Departure   string  `json:"departure"`
	Destination string  `json:"arrival"`
	Duration    int     `json:"duration"`
	Avg_price   float32 `json:"price"`
}

type Company struct {
	Id       int     `json:"id"`
	Name     string  `json:"name"`
	Multiply float32 `json:"multiply"`
}

type Fly struct {
	Airway  Route   `json:"route"`
	Corpor  Company `json:"company"`
	HourDep string  `json:"timeDep"`
	HourArr string  `json:"timeArr"`
	Price   float32 `json:"price"`
	Sales   int32   `json:"sales"`
}

func databaseConection() (db *sql.DB, e error) {
	user := "root"
	pass := "aaaa"
	host := "tcp(127.0.0.1:3306)"
	databaseName := "skycompare"

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@%s/%s", user, pass, host, databaseName))
	if err != nil {
		return nil, err
	}
	return db, nil
}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		return
	}
	r.ParseForm()
	fmt.Print(r)

	fmt.Print("\nPetition of login " + r.Form.Get("nickname") + "\n")
	db, err := databaseConection()
	if err != nil {
		fmt.Printf("Error getting database: %v", err)
		return
	}

	rows, err := db.Query("SELECT * FROM users WHERE nickname=? AND pass=?", r.Form.Get("nickname"), r.Form.Get("password"))
	var u User
	users := []User{}
	fmt.Printf("\nSearching user in db\n")
	fmt.Print(err)
	for rows.Next() {
		rows.Scan(&u.id, &u.nickname, &u.full_name, &u.email, &u.pass, &u.fav_airport)
		users = append(users, u)
		fmt.Printf("%v\n", u)
	}

	if len(users) == 0 {
		io.WriteString(w, "User or password are incorrect")
	} else {
		io.WriteString(w, "ok")
	}

}

func register(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Print(r)
	fmt.Print(r.Form.Get("nickname"))

	db, err := databaseConection()
	if err != nil {
		fmt.Printf("Error getting database: %v", err)
		return
	}

	data := []byte(r.Form.Get("password"))
	fmt.Printf("%x", md5.Sum(data))

	rows, err := db.Query("SELECT * FROM users WHERE nickname=?", r.Form.Get("nickname"))

	defer db.Close()

	var u User
	users := []User{}
	fmt.Printf("\nTesting if new user is in db\n")
	for rows.Next() {
		rows.Scan(&u.id, &u.nickname, &u.full_name, &u.email, &u.pass, &u.fav_airport)
		users = append(users, u)
		fmt.Printf("%v\n", u)
	}

	if len(users) != 0 {
		io.WriteString(w, "User is in db")
	} else {
		id := 0
		row, err := db.Query("SELECT MAX(id) FROM users")
		row.Next()
		row.Scan(&id)
		if err == nil {
			sentence, err := db.Prepare("INSERT INTO users (id, nickname, full_name, email, pass, fav_airport) VALUES(?,?,?,?,?,?)")
			if err == nil {
				_, err := sentence.Exec((id + 1), r.Form.Get("nickname"), r.Form.Get("full_name"), r.Form.Get("email") /*md5.Sum(data)*/, r.Form.Get("password"), nil)
				if err != nil {
					fmt.Printf("\nError in Exec: %v", err)
				}
				io.WriteString(w, "User registration confirmed")
			} else {
				fmt.Printf("\nError in query for id: %v\n", err)
			}
		} else {
			fmt.Printf("\nError in query: %v\n", err)
		}
	}
}

func airports(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\nCharging airports")
	db, err := databaseConection()
	if err != nil {
		fmt.Printf("Error getting database: %v", err)
		return
	}

	rows, err := db.Query("SELECT id, name, location, IATA FROM airports")

	defer rows.Close()
	defer db.Close()

	if err == nil {
		airports := []Airport{}

		for rows.Next() {
			var a Airport
			rows.Scan(&a.Id, &a.Name, &a.Location, &a.IATA)
			airports = append(airports, a)
			fmt.Printf("%v\n", a)
		}

		fmt.Print(airports)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(airports)
	} else {
		fmt.Printf("\n%v", err)
	}
}

func airportsWithoutOne(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var noAirport = r.Form.Get("IATA")
	fmt.Printf("\nCharging airports minus %s", noAirport)

	db, err := databaseConection()
	if err != nil {
		fmt.Printf("Error getting database: %v", err)
		return
	}

	rows, err := db.Query("SELECT id, name, location, IATA FROM airports")

	defer rows.Close()
	defer db.Close()

	if err == nil {
		airports := []Airport{}

		for rows.Next() {
			var a Airport
			rows.Scan(&a.Id, &a.Name, &a.Location, &a.IATA)
			if a.IATA != noAirport {
				airports = append(airports, a)
				fmt.Printf("%v\n", a)
			}
		}

		fmt.Print(airports)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(airports)
	} else {
		fmt.Printf("\n%v", err)
	}
}

func favourite(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Print(r)
	db, err := databaseConection()
	if err != nil {
		fmt.Printf("Error getting database: %v", err)
		return
	}

	fav := r.Form.Get("fav")
	fmt.Printf("%s", fav)
	if fav == "No favourite airport" {
		fav = ""
	}
	fmt.Printf("%s", fav)
	rows, err1 := db.Exec("UPDATE users SET fav_airport=? WHERE nickname=?", fav, r.Form.Get("user"))

	fmt.Print(rows)
	defer db.Close()

	if err1 != nil {
		fmt.Printf("%v", err1)
		io.WriteString(w, "An error ocurred, contact with one administrator (Jose or David)")
	} else {
		io.WriteString(w, "Favourite airport saved corretly, thanks")
	}
}

func getFavourite(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fmt.Print(r)
	db, err := databaseConection()
	if err != nil {
		fmt.Printf("Error getting database: %v", err)
		return
	}

	type airport struct {
		Airport string `json:"airport"`
	}

	var a airport
	nickName := r.Form.Get("nickName")
	fmt.Printf("\nSearching favourite airport for %s\n", nickName)
	rows, err1 := db.Query("SELECT fav_airport FROM users WHERE nickname=?", nickName)

	defer db.Close()
	if err1 == nil {
		for rows.Next() {
			rows.Scan(&a.Airport)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(a)
	} else {
		fmt.Printf("\nError: %v\n", err1)
	}
}

func getRoutes(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	var dep = r.Form.Get("dep")
	var arr = r.Form.Get("arr")
	fmt.Printf("\nSearching routes for %s-%s", dep, arr)

	db, err := databaseConection()
	if err != nil {
		fmt.Printf("Error getting database: %v", err)
		return
	}

	rows, err := db.Query("SELECT * FROM routes WHERE departure=? && destination=?", dep, arr)
	corporations, err2 := db.Query("SELECT * FROM company")

	defer rows.Close()
	defer corporations.Close()
	defer db.Close()

	if err == nil && err2 == nil {
		airways := []Fly{}
		companies := []Company{}

		for corporations.Next() {
			var c Company
			corporations.Scan(&c.Id, &c.Name, &c.Multiply)
			companies = append(companies, c)
		}

		var r Route
		for rows.Next() {
			rows.Scan(&r.Id, &r.Departure, &r.Destination, &r.Duration, &r.Avg_price)
			fmt.Printf("\n%v", r)
		}

		var rnd = rand.Intn(4)

		fmt.Printf("\n%d companies on route\n", rnd)
		comps := "    "
		for count := 0; count <= rnd; count++ {
			num := rand.Intn(len(companies))
			if !strings.Contains(comps, strconv.Itoa(num)) {
				comps = comps + strconv.Itoa(num) + "    "
				ranComponent := rand.Float32()*(1.35-0.65) + 0.65
				price := r.Avg_price * companies[num].Multiply * ranComponent
				var sales int32 = 0
				if ranComponent <= 0.8 {
					sales = 1
				}

				timeDep := rand.Float32() * 23.99
				minDep := timeDep * 60
				minArr := minDep + float32(r.Duration)
				if minArr > 1440 {
					minArr = minArr - 1440
				}

				hourDep := strconv.Itoa(int(timeDep)) + ":"   // +
				hourArr := strconv.Itoa(int(minArr/60)) + ":" // +

				if timeDep < 10 {
					hourDep = "0" + hourDep
				}
				if minArr/60 < 10 {
					hourArr = "0" + hourArr
				}

				remainDep := strconv.Itoa(int(minDep)%60) + ""
				remainArr := strconv.Itoa(int(minArr)%60) + ""

				if int(minDep)%60 < 10 {
					remainDep = "0" + remainDep
				}
				if int(minArr)%60 < 10 {
					remainArr = "0" + remainArr
				}

				hourDep = hourDep + remainDep
				hourArr = hourArr + remainArr

				f := Fly{r, companies[num], hourDep, hourArr, price, sales}
				airways = append(airways, f)
			} else {
				count = count - 1
			}
		}
		fmt.Print(airways)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(airways)
	} else {
		fmt.Printf("\n%v", err)
	}
}

func saveFlies(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fmt.Print(r)
	var nickName = r.Form.Get("nickname")
	var route_id = r.Form.Get("route_id")
	var companies = strings.Split(r.Form.Get("companies"), " ")
	var prices = strings.Split(r.Form.Get("prices"), " ")
	var hoursDep = strings.Split(r.Form.Get("hoursDep"), " ")
	var hoursArr = strings.Split(r.Form.Get("hoursArr"), " ")

	db, err := databaseConection()
	if err != nil {
		fmt.Printf("Error getting database: %v", err)
		return
	}

	defer db.Close()

	var id int

	row, err2 := db.Query("SELECT MAX(id) FROM routes_probed")

	if err2 == nil {
		defer row.Close()

		if row.Next() {
			row.Scan(&id)
		}
	}

	fmt.Printf("\n\n%d\n", id)
	sentence, err1 := db.Prepare("INSERT INTO routes_probed (id, id_user, route_id, company1, price1, timeDep1, timeArr1, company2, price2, timeDep2, timeArr2,  company3, price3, timeDep3, timeArr3,  company4, price4, timeDep4, timeArr4,  company5, price5, timeDep5, timeArr5) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")

	if err1 == nil && err2 == nil {
		var id_user int32

		user, errU := db.Query("SELECT id FROM users WHERE nickname=?", nickName)

		if errU == nil {
			defer user.Close()
			if user.Next() {
				user.Scan(&id_user)
			}

			_, err3 := sentence.Exec((id + 1), id_user, route_id, companies[0], prices[0], hoursDep[0], hoursArr[0], companies[1], prices[1], hoursDep[1], hoursArr[1], companies[2], prices[2], hoursDep[2], hoursArr[2], companies[3], prices[3], hoursDep[3], hoursArr[3], companies[4], prices[4], hoursDep[4], hoursArr[4])
			if err3 == nil {
				io.WriteString(w, "Saved conrrectly")
			} else {
				fmt.Printf("*%v\n", err3)
				io.WriteString(w, "Error: ask for help an administrator")
			}
		} else {
			fmt.Printf("%v", errU)
			io.WriteString(w, "Error: ask for help an administrator")
		}
	} else {
		fmt.Printf("%v", err1)
		io.WriteString(w, "Error: ask for help an administrator")
	}
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// Just put some headers to allow CORS...
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

			if r.Method == "OPTIONS" {
				return
			}

			// and call next handler!
			next.ServeHTTP(w, r)
		})
}

func main() {
	app := http.NewServeMux()

	db, err := databaseConection()
	if err != nil {
		fmt.Printf("Error getting database: %v", err)
		return
	}

	// Test connection
	err = db.Ping()
	if err != nil {
		fmt.Printf("Error in connection: %v", err)
		return
	} else {
		fmt.Printf("Connection right")
	}

	app.HandleFunc("/register", register)
	app.HandleFunc("/login", login)
	app.HandleFunc("/airports", airports)
	app.HandleFunc("/airports/selected", airportsWithoutOne)
	app.HandleFunc("/routes", getRoutes)
	app.HandleFunc("/favAirport", favourite)
	app.HandleFunc("/favAirportGet", getFavourite)
	app.HandleFunc("/save", saveFlies)

	http.ListenAndServe(":5152", middlewareCors(app))

	// close connection at the end of fucntion
	defer db.Close()
}
