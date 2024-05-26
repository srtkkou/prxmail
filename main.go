package prxmail

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func AppMain(gitRevision string) (code int) {
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error=%+v\n", err)
		return -1
	}
	fmt.Printf("host=%s,port=%s,from=%s,password=%s\n",
		os.Getenv("HOST"), os.Getenv("PORT"),
		os.Getenv("FROM"), os.Getenv("PASSWORD"))
	return 0
}
