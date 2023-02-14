package service

import (
	"github.com/jackc/pgx"
	"golang.org/x/crypto/bcrypt"

)

func isUniqueViolation(err error) bool {
	pgerr, ok := err.(pgx.PgError)
	return ok && pgerr.Code == "23505"
}

func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}

func checkPasswordHash(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

// func searchAndAppend(arr *[]int, val int) {
// 	found := false
// 	for _, v := range *arr {
// 		if v == val {
// 			found = true
// 			break
// 		}
// 	}

// 	if !found {
// 		*arr = append(*arr, val)
// 	}
// } 

// func searchAndDelete(arr *[]int, val int) {
// 	for i, v := range *arr {
// 		if v == val {
// 			*arr = append((*arr)[:i], (*arr)[i+1:]...)
// 			return
// 		}
// 	}

// 	*arr = append(*arr, val)
// }
