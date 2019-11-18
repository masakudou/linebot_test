package main

import (
	"./scraping"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/line/line-bot-sdk-go/linebot/httphandler"
)

func main() {
	// LINE bot SDKに含まれるhttpHandlerの初期化
	handler, err := httphandler.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	// 初期化に失敗した場合
	if err != nil {
		log.Fatal(err)
	}

	// ポート番号 環境変数がセットされていない場合は8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("ポート番号にデフォルト値 %s をセットします.", port)
	}

	// リクエストを受け取った時に実行する関数を定義し、ハンドラに登録
	handler.HandleEvents(func(events []*linebot.Event, r *http.Request) {
		bot, err := handler.NewClient()
		// クライアント初期化に失敗
		if err != nil {
			log.Print(err)
			return
		}
		for _, event := range events {
			// Event種別がメッセージでなければ終了
			if event.Type != linebot.EventTypeMessage {
				return
			}

			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				switch message.Text {
				case "運行情報":
					// 京成本線の運行情報をスクレイピング
					trainInfoText := scraping.ScrapingTrainInfo("https://transit.yahoo.co.jp/traininfo/detail/96/0/")
					// trainInfoTextの内容を送信
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(trainInfoText)).Do(); err != nil {
						log.Print(err)
					}
				}
			}
		}
	})

	// /callback にエンドポイントの定義
	http.Handle("/callback", handler)
	// HTTPサーバの起動
	log.Printf("ポート番号: %s をリッスン...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
