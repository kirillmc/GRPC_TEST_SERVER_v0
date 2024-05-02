package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"

	_ "github.com/lib/pq"

	"github.com/joho/godotenv"

	desc "github.com/kirillmc/GRPC_TEST_SERVER/pkg/user_v1"
)

const (
	idColumn = "id"
	name     = "name"
	surname  = "surname"
	email    = "email"
	avatar   = "avatar"
	login    = "login"
	password = "password"
	role     = "role"
	weight   = "weight"
	height   = "height"
	locked   = "locked"
)

type User struct {
	Id       int64
	Name     string
	Surname  string
	Email    string
	Avatar   string
	Login    string
	Password string
	Role     int32
	Weight   float64
	Height   float64
	Locked   bool
}

type server struct {
	desc.UnimplementedUserV1Server
	db *sql.DB
}

func main() {
	// Загрузка значений из файла .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Подключение к базе данных PostgreSQL
	db, err := sql.Open("postgres", os.Getenv("PG_DSN"))
	if err != nil {
		log.Printf("ERROROR IS: %v\n", err)
		log.Fatal(err)
	}
	defer db.Close()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 50051))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	reflection.Register(s)
	desc.RegisterUserV1Server(s, &server{db: db})

	log.Printf("server listening at %v", lis.Addr())

	if err = s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

func (s *server) Create(ctx context.Context, req *desc.CreateRequest) (*desc.CreateResponse, error) {
	// Делаем запрос на вставку записи в таблицу note
	query := fmt.Sprintf("INSERT INTO users (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s) VALUES ($1, $2, $3, $4, $5,$6, $7, $8, $9, $10 ) RETURNING id ", name, surname, email, avatar, login, password, role, weight, height, locked)

	var id int64
	err := s.db.QueryRow(query,
		req.User.Name,
		req.User.Surname,
		req.User.Email,
		req.User.Avatar,
		req.User.Login,
		req.User.Password,
		req.User.Role,
		req.User.Weight,
		req.User.Height,
		req.User.Locked,
	).
		Scan(&id)

	if err != nil {
		return nil, err
	}
	return &desc.CreateResponse{
		Id: id,
	}, nil
}

func (s *server) Get(ctx context.Context, req *desc.GetRequest) (*desc.GetResponse, error) {
	var user User

	query := fmt.Sprintf("SELECT %s, %s, %s, %s, %s, %s, %s, %s, %s, %s FROM users WHERE id = $1", name, surname, email, avatar, login, password, role, weight, height, locked)

	err := s.db.QueryRowContext(ctx, query, req.GetId()).
		Scan(
			&user.Name,
			&user.Surname,
			&user.Email,
			&user.Avatar,
			&user.Login,
			&user.Password,
			&user.Role,
			&user.Weight,
			&user.Height,
			&user.Locked,
		)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &desc.GetResponse{
		Id: req.Id,
		User: &desc.User{
			Name:     user.Name,
			Surname:  user.Surname,
			Email:    user.Email,
			Avatar:   user.Avatar,
			Login:    user.Login,
			Password: user.Password,
			Role:     desc.Role(user.Role),
			Weight:   user.Weight,
			Height:   user.Height,
			Locked:   user.Locked,
		},
	}, nil
}

func (s *server) GetUsers(ctx context.Context, _ *emptypb.Empty) (*desc.GetUsersResponse, error) {
	query := fmt.Sprintf("SELECT %s ,%s, %s, %s, %s, %s, %s, %s, %s, %s, %s FROM users", idColumn, name, surname, email, avatar, login, password, role, weight, height, locked)
	rows, err := s.db.Query(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	users := []*desc.GetResponse{}
	for rows.Next() {
		var user User
		err := rows.Scan(&user.Id, &user.Name, &user.Surname, &user.Email, &user.Avatar, &user.Login, &user.Password,
			&user.Role, &user.Weight, &user.Height, &user.Locked)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		users = append(users, &desc.GetResponse{
			Id: user.Id,
			User: &desc.User{
				Name:     user.Name,
				Surname:  user.Surname,
				Email:    user.Email,
				Avatar:   user.Avatar,
				Login:    user.Login,
				Password: user.Password,
				Role:     desc.Role(user.Role),
				Weight:   user.Weight,
				Height:   user.Height,
				Locked:   user.Locked,
			},
		})
	}

	return &desc.GetUsersResponse{Users: users}, nil
}

func (s *server) Update(ctx context.Context, req *desc.UpdateRequest) (*emptypb.Empty, error) {
	// Формируем запрос к базе данных для обновления пользователя
	query := fmt.Sprintf("UPDATE users SET %s=$1, %s=$2, %s=$3, %s=$4, %s=$5, %s=$6, %s=$7, %s=$8, %s=$9, %s=$10 WHERE id=$11", name, surname, email, avatar, login, password, role, weight, height, locked)
	_, err := s.db.Exec(query,
		req.User.Name,
		req.User.Surname,
		req.User.Email,
		req.User.Avatar,
		req.User.Login,
		req.User.Password,
		req.User.Role,
		req.User.Weight,
		req.User.Height,
		req.User.Locked,
		req.GetId())
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return nil, nil
}

func (s *server) Delete(ctx context.Context, req *desc.DeleteRequest) (*emptypb.Empty, error) {
	// Удаляем пользователя из базы данных
	result, err := s.db.Exec("DELETE FROM users WHERE id = $1", req.Id)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Проверяем, сколько строк было затронуто удалением
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Если ни одна строка не была затронута удалением, значит пользователь с указанным id не найден
	if rowsAffected == 0 {
		return nil, err
	}

	return nil, nil
}
