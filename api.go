package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "modernc.org/sqlite"
)

var db *sql.DB

type Usuario struct{
	ID 		int		`json:"id"`
	Nome 	string	`json:"nome"`
	Email 	string	`json:"email"`
}

func main() {

	var err error

	db, err = sql.Open("sqlite", "banco.db")
	if err != nil{
		log.Fatal(err)
	}

	defer db.Close()

	err = criarTabela(db)
	if err != nil{
		log.Fatalf("erro ao criar a tabela: %v", err )
	}

	r := mux.NewRouter()

	r.HandleFunc("/usuarios", GetUsers).Methods("GET")
	r.HandleFunc("/usuarios", AddUser).Methods("POST")
	r.HandleFunc("/usuarios/{id}", DeleteUser).Methods("DELETE")
	r.HandleFunc("/usuarios/{id}", UpdateUser).Methods("PUT")


	fmt.Printf("Servidor rodando na porta 3000")
	http.ListenAndServe(":3000", r)
}


func criarTabela(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS usuarios (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		nome TEXT,
		email TEXT
	);`

	_, err := db.Exec(query)
	return err
}


func GetUsers(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-type", "application/json")	//Aqui estamos definindo o cabeçalho da resposta
	rows, err := db.Query("SELECT id, nome, email FROM usuarios")	//Faz a consulta e retorna "rows" com a resposa da requisição
																	//rows é um ponteiro para sql.Rows
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)	//tratamento de erro 
		return
	}

	defer rows.Close()	//Fecha o recurso rows após a execução da função

	var usuarios []Usuario	//Cria uma slice de structs Usuario, onde cada elemento representará um usuário

	for rows.Next() {
		var u Usuario	//Variavel para armazenar um usuario a cada iteração do loop

		if err := rows.Scan(&u.ID, &u.Nome, &u.Email); err != nil {
			http.Error(w, "Erro ao ler os dados", http.StatusInternalServerError)	//tratamento de erro 
			return
		}

		usuarios = append(usuarios, u)	// adiciona um usuário (u) ao Slice usuarios
	}

	json.NewEncoder(w).Encode(usuarios) // transforma a resposta em JSON 
}

func AddUser(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-type", "application/json")

	var u Usuario

	if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	if u.Nome == ""{
		http.Error(w, "O campo NOME não pode estar vazio", http.StatusBadRequest)
		return
	}

	if u.Email == ""{
		http.Error(w, "O campo EMAIL não pode estar vazio", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("INSERT INTO usuarios (nome, email) VALUES (?,?)", u.Nome, u.Email)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	id, _ := result.LastInsertId()
	u.ID = int(id)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(u)
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "applicatio/json")

	vars := mux.Vars(r)
	id := vars["id"]

	result, err := db.Exec("DELETE FROM usuarios WHERE id = ?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		http.Error(w, "Usuário não encontrado", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Usuário deletado com sucesso"})
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id := params["id"]
	idInt, err := strconv.Atoi(id)

	if err != nil{
		http.Error(w, "ID inválido", http.StatusBadRequest)
		return
	}

	var u Usuario
	err = json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	result, err := db.Exec("UPDATE usuarios SET nome = ?, email = ? WHERE id = ?", u.Nome, u.Email, idInt )
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, _ := result.RowsAffected() 
	if rowsAffected == 0{
		http.Error(w, "Usuário não encontrado", http.StatusNotFound)
		return
	}

	u.ID = idInt
	json.NewEncoder(w).Encode(u)
}








