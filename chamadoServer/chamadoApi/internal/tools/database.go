package tools

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type UserData struct {
	Name        string
	Password    string
	Team        string
	Role        string
	TimeCreated string
}

type ChamadoData struct {
	Number      int
	Name        string
	Section     string
	Description string
	Response    string
	Status      string
	CreatedAt   string
}

var Database *sql.DB

func connString() string {
	host := getEnvDefault("POSTGRES_HOST", "localhost")
	port := getEnvDefault("POSTGRES_PORT", "5432")
	user := getEnvDefault("POSTGRES_USER", "postgres")
	password := getEnvDefault("POSTGRES_PASSWORD", "postgres")
	dbName := getEnvDefault("POSTGRES_DB", "chamado-db")
	sslMode := getEnvDefault("POSTGRES_SSLMODE", "disable")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbName, sslMode)
}

func getEnvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func Setup() error {
	var err error
	Database, err = sql.Open("postgres", connString())
	if err != nil {
		return fmt.Errorf("failed to open postgres: %w", err)
	}

	if err = waitForDatabase(Database, 10, 2*time.Second); err != nil {
		return err
	}

	if err = CreateUserTable(Database); err != nil {
		return fmt.Errorf("failed to create users table: %w", err)
	}

	if err = CreateChamadoTable(Database); err != nil {
		return fmt.Errorf("failed to create chamados table: %w", err)
	}

	return nil
}

func waitForDatabase(database *sql.DB, attempts int, delay time.Duration) error {
	var err error
	for i := 0; i < attempts; i++ {
		if err = database.Ping(); err == nil {
			return nil
		}
		log.Warnf("database not ready yet (attempt %d/%d): %v", i+1, attempts, err)
		time.Sleep(delay)
	}
	return fmt.Errorf("failed to ping postgres after %d attempts: %w", attempts, err)
}

func CreateUserTable(database *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS usuarios (
	name VARCHAR(100) PRIMARY KEY NOT NULL,
	password VARCHAR(30) NOT NULL,
	team VARCHAR(50) NOT NULL,
	role VARCHAR(15) NOT NULL DEFAULT 'technician',
	created TIMESTAMP DEFAULT NOW()
	)`
	_, err := database.Exec(query)
	if err != nil {
		log.Error("error on create usuarios table query: ", err)
		return err
	}
	return nil
}

func CreateChamadoTable(database *sql.DB) error {
	query := `CREATE TABLE IF NOT EXISTS chamados (
	number SERIAL PRIMARY KEY,
	name VARCHAR(100) NOT NULL,
	section VARCHAR(50) NOT NULL,
	description TEXT NOT NULL,
	response TEXT NOT NULL DEFAULT '',
	status VARCHAR(20) NOT NULL DEFAULT 'aberto',
	created_at TIMESTAMP DEFAULT NOW()
	)`
	_, err := database.Exec(query)
	if err != nil {
		log.Error("error on create chamados table query: ", err)
		return err
	}
	return nil
}

// USERS

func CreateUser(database *sql.DB, user UserData) (string, error) {
	query := `INSERT INTO usuarios (name, password, team, role)
VALUES ($1, $2, $3, $4) RETURNING name`
	var nameCreated string
	err := database.QueryRow(query, user.Name, user.Password, user.Team, user.Role).Scan(&nameCreated)
	if err != nil {
		log.Error("error on create user query: ", err)
		return nameCreated, err
	}
	return nameCreated, nil
}

func DeleteUser(database *sql.DB, name string) error {
	query := "DELETE FROM usuarios WHERE name = ($1) RETURNING name, password, team, role, created"
	var deletedRow UserData
	err := database.QueryRow(query, name).Scan(&deletedRow.Name, &deletedRow.Password, &deletedRow.Team, &deletedRow.Role, &deletedRow.TimeCreated)
	if err != nil {
		return errors.New("No matching rows")
	}
	return nil
}

func GetUserData(database *sql.DB, name string) (UserData, error) {
	query := "SELECT name, password, team, role, created FROM usuarios WHERE name = ($1)"
	var user UserData
	err := database.QueryRow(query, name).Scan(&user.Name, &user.Password, &user.Team, &user.Role, &user.TimeCreated)
	if err != nil {
		return user, errors.New("No matching rows")
	}
	return user, nil
}

// CHAMADOS

func CreateChamado(database *sql.DB, chamado ChamadoData) (ChamadoData, error) {
	query := `INSERT INTO chamados (name, section, description, status)
VALUES ($1, $2, $3, 'aberto') RETURNING number, name, section, description, response, status, created_at`
	var created ChamadoData
	err := database.QueryRow(query, chamado.Name, chamado.Section, chamado.Description).
		Scan(&created.Number, &created.Name, &created.Section, &created.Description, &created.Response, &created.Status, &created.CreatedAt)
	if err != nil {
		log.Error("error on create chamado query: ", err)
		return created, err
	}
	return created, nil
}

func GetChamadoByNumber(database *sql.DB, number int) (ChamadoData, error) {
	query := "SELECT number, name, section, description, response, status, created_at FROM chamados WHERE number = ($1)"
	var chamado ChamadoData
	err := database.QueryRow(query, number).
		Scan(&chamado.Number, &chamado.Name, &chamado.Section, &chamado.Description, &chamado.Response, &chamado.Status, &chamado.CreatedAt)
	if err != nil {
		return chamado, errors.New("No matching rows")
	}
	return chamado, nil
}

func GetChamadoBySection(database *sql.DB, section string, status string) ([]ChamadoData, error) {
	var rows *sql.Rows
	var err error
	if status == "" {
		query := "SELECT number, name, section, description, response, status, created_at FROM chamados WHERE section = ($1) ORDER BY created_at DESC"
		rows, err = database.Query(query, section)
	} else {
		query := "SELECT number, name, section, description, response, status, created_at FROM chamados WHERE section = ($1) AND status = ($2) ORDER BY created_at DESC"
		rows, err = database.Query(query, section, status)
	}
	var chamados []ChamadoData
	if err != nil {
		return chamados, errors.New("No matching rows")
	}
	defer rows.Close()
	for rows.Next() {
		var chamado ChamadoData
		if err := rows.Scan(&chamado.Number, &chamado.Name, &chamado.Section, &chamado.Description, &chamado.Response, &chamado.Status, &chamado.CreatedAt); err != nil {
			return chamados, errors.New("scan failed")
		}
		chamados = append(chamados, chamado)
	}
	return chamados, nil
}

func GetAllChamados(database *sql.DB, status string) ([]ChamadoData, error) {
	var rows *sql.Rows
	var err error
	if status == "" {
		rows, err = database.Query("SELECT number, name, section, description, response, status, created_at FROM chamados ORDER BY created_at DESC")
	} else {
		rows, err = database.Query("SELECT number, name, section, description, response, status, created_at FROM chamados WHERE status = ($1) ORDER BY created_at DESC", status)
	}
	var chamados []ChamadoData
	if err != nil {
		return chamados, errors.New("No matching rows")
	}
	defer rows.Close()
	for rows.Next() {
		var chamado ChamadoData
		if err := rows.Scan(&chamado.Number, &chamado.Name, &chamado.Section, &chamado.Description, &chamado.Response, &chamado.Status, &chamado.CreatedAt); err != nil {
			return chamados, errors.New("scan failed")
		}
		chamados = append(chamados, chamado)
	}
	return chamados, nil
}

func UpdateChamado(database *sql.DB, number int, status string, response string) (ChamadoData, error) {
	query := `UPDATE chamados SET status = ($1), response = ($2) WHERE number = ($3)
RETURNING number, name, section, description, response, status, created_at`
	var updated ChamadoData
	err := database.QueryRow(query, status, response, number).
		Scan(&updated.Number, &updated.Name, &updated.Section, &updated.Description, &updated.Response, &updated.Status, &updated.CreatedAt)
	if err != nil {
		return updated, errors.New("No matching rows")
	}
	return updated, nil
}

func DeleteChamado(database *sql.DB, number int) error {
	query := "DELETE FROM chamados WHERE number = ($1) RETURNING number"
	var deletedNumber int
	err := database.QueryRow(query, number).Scan(&deletedNumber)
	if err != nil {
		return errors.New("No matching rows")
	}
	return nil
}
