package main

import (
	"cis-telegram/database"
	"cis-telegram/database/bot"
	"cis-telegram/database/push"
	"cis-telegram/database/settings"
	"cis-telegram/serv"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("START")

	log.Println("runtime.GOMAXPROCS:", runtime.GOMAXPROCS(0))

	if err := godotenv.Load("../.env"); err != nil {
		log.Println("Cant load .env: ", err)
	}

	// telegramReportGroupID, _ := strconv.ParseInt(os.Getenv("TELEGRAM_GROUP_ID"), 10, 64)

	mysqlURL := os.Getenv("MYSQL_URL")

	if os.Getenv("MYSQL_URL_FILE") != "" {
		mysqlURL_, err := os.ReadFile(os.Getenv("MYSQL_URL_FILE"))
		if err != nil {
			log.Fatal(err)
		}
		mysqlURL = strings.TrimSpace(string(mysqlURL_))
	}

	dbService, err := database.NewService(mysqlURL)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println("dbService OK")
	}

	/*telegramApiToken := os.Getenv("TELEGRAM_STATUS_APITOKEN")
	if os.Getenv("TELEGRAM_STATUS_APITOKEN_FILE") != "" {
		telegramApiToken_, err := os.ReadFile(os.Getenv("TELEGRAM_STATUS_APITOKEN_FILE"))
		if err != nil {
			log.Fatal(err)
		}
		telegramApiToken = strings.TrimSpace(string(telegramApiToken_))
	}*/

	settingsService, err := settings.NewService(dbService, 60)
	if err != nil {
		log.Fatal(err)
	}

	botService, err := bot.NewService(dbService, settingsService, 60)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	_ = botService

	/*telegramService, err := telegram.NewService(telegramApiToken, dbService, settingsService)
	if err != nil {
		log.Fatal(err)
	}*/

	_, err = push.NewService(dbService, botService)
	if err != nil {
		log.Fatal(err)
	}

	// telegramService.Send(telegramReportGroupID, fmt.Sprintf("dmca started"))

	httpService, err := serv.NewService(os.Getenv("PORT"))
	if err != nil {
		log.Fatal(err)
	}
	httpService.Run()
}
