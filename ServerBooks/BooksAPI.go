package ServerBooks

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
)

type Book struct {
	Id       string `json:"id"`
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
}

var Books = make([]Book, 0)

func OperFile(nameFile string) *os.File {
	f, err1 := os.Open(nameFile)
	if err1 != nil {
		log.Fatal(err1)
	}
	return f
}

func ReadFile(f *os.File) {
	reader := csv.NewReader(f)
	CSVfile, err := reader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}
	ConvertCSVtoBook(CSVfile)
}

func UpdateFile() {
	tmpSlice := [][]string{}

	fmt.Println(len(Books))

	for _, book := range Books {
		row := []string{}
		row = append(row, book.Id)
		row = append(row, book.Title)
		row = append(row, book.Subtitle)

		tmpSlice = append(tmpSlice, row)
	}

	nf, err := os.Create("listaLibri.csv")
	if err != nil {
		log.Fatal(err)
	}
	err = csv.NewWriter(nf).WriteAll(tmpSlice)
	nf.Close()
	if err != nil {
		log.Fatal(err)
	}
}

func ConvertCSVtoBook(filecsv [][]string) {
	svuotaBooks()
	for _, line := range filecsv {
		var rec Book
		rec.Id = line[0]
		rec.Title = line[1]
		rec.Subtitle = line[2]

		Books = append(Books, rec)
	}
}

func svuotaBooks() {
	Books = Books[:0]
}

func emptyInputDecoder(*http.Request) (interface{}, error) {
	return nil, nil
}

func getIdDecoder(r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	key := vars["id"]
	//todo controllare esisteza id altrimenti tornare errore
	return key, nil
}

func getBookDecoder(r *http.Request) (interface{}, error) {
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Risulta qualche errore", err)
		return nil, err
	}
	var book Book
	err = json.Unmarshal(reqBody, &book)
	if err != nil {
		fmt.Println("Risulta qualche errore", err)
		return nil, err
	}
	return book, nil
}

type Msg struct {
	Msg string
}

func Wrapper(fn func(interface{}) (interface{}, error), dec func(*http.Request) (interface{}, error)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		payload, err := dec(r)
		if err != nil {
			json.NewEncoder(w).Encode("400")
			return
		}

		resp, err := fn(payload)
		if err != nil {
			json.NewEncoder(w).Encode(Msg{
				Msg: err.Error(),
			})
			return
		}

		jsonData, err := json.Marshal(resp)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Scrive il JSON come risposta
		w.Write(jsonData)
	}
}

func homePage(req interface{}) (interface{}, error) {
	fmt.Println("GET /books to view all books list;\n" +
		"GET /book/{id} to view the book identified by the id;\n" +
		"POST /book/create -d '{}' to save one book in the slice;\n" +
		"DELETE /books/delete/{id} to delete the book identified by the id;\n" +
		"/books/sort to sort the slice in order of id;")
	return nil, nil
}

func allBooks(_ interface{}) (interface{}, error) {
	return Books, nil
}

func getBookByID(key interface{}) (interface{}, error) {
	for _, book := range Books {
		if book.Id == key {
			return book, nil
		}
	}
	return nil, fmt.Errorf("Non trovato id")
}

func createNewBook(req interface{}) (interface{}, error) {
	book, ok := req.(Book)
	if !ok {
		return nil, fmt.Errorf("Book non riconoscibile")
	}

	Books = append(Books, book)

	UpdateFile()

	return book, nil

}

func deleteBooks(key interface{}) (interface{}, error) {

	flag := true

	for i, book := range Books {
		if book.Id == key {
			flag = false
			Books = append(Books[:i], Books[i+1:]...)
			UpdateFile()
		}
	}
	if flag {
		return nil, fmt.Errorf("Non trovato id")
	} else {
		for i, book := range Books {
			if book.Id == key {
				flag = false
				Books = append(Books[:i], Books[i+1:]...)
				UpdateFile()
				break
			}
		}
	}

	return Books, nil
}

func sortBooks(_ interface{}) (interface{}, error) {

	sort.Slice(Books, func(i, j int) bool {
		if Books[i].Id < Books[j].Id {
			return true
		} else {
			return false
		}
	})

	UpdateFile()

	return Books, nil
}

func HandleRequests() {

	//apertura file
	f := OperFile("listaLibri.csv")
	defer f.Close()

	//lettura file
	ReadFile(f)

	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", Wrapper(homePage, emptyInputDecoder))
	myRouter.HandleFunc("/books", Wrapper(allBooks, emptyInputDecoder)).Methods("GET")
	myRouter.HandleFunc("/books/{id}", Wrapper(getBookByID, getIdDecoder)).Methods("GET")
	myRouter.HandleFunc("/books", Wrapper(createNewBook, getBookDecoder)).Methods("POST")
	myRouter.HandleFunc("/books/{id}", Wrapper(deleteBooks, getIdDecoder)).Methods("DELETE")
	myRouter.HandleFunc("/sorts/books", Wrapper(sortBooks, emptyInputDecoder)).Methods("POST")

	http.Handle("/", myRouter)
}
