package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/PuerkitoBio/goquery"
	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/line/line-bot-sdk-go/linebot/httphandler"
)

// ScrapingTrainInfo Yahoo!路線情報の各路線のページへアクセスし、運行情報のテキストをスクレイピングして返す.
func ScrapingTrainInfo(url string) string {
	const errorMessage = "大変申し訳ございません。エラーが発生しました。時間をおいて試してみて下さい。"

	response, err := http.Get(url)
	if err != nil {
		log.Print(err)
		return errorMessage
	}

	// HTTP Response Bodyをクローズ
	defer response.Body.Close()

	if response.StatusCode != 200 {
		log.Printf("status code error: %d %s", response.StatusCode, response.Status)
		return errorMessage
	}

	// HTMLドキュメントの取得
	document, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		log.Print(err)
		return errorMessage
	}

	// セレクタの取得
	selection := document.Find("div#mdServiceStatus")
	innerSelection := selection.Find("p")

	trainInfoText := innerSelection.Text()

	return trainInfoText
}

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
				case "京成本線":
					// 京成本線の運行情報をスクレイピング
					trainInfoText := ScrapingTrainInfo("https://transit.yahoo.co.jp/traininfo/detail/96/0/")
					// trainInfoTextの内容を送信
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(trainInfoText)).Do(); err != nil {
						log.Print(err)
					}
				case "都営浅草線":
					// 都営浅草線の運行情報をスクレイピング
					trainInfoText := ScrapingTrainInfo("https://transit.yahoo.co.jp/traininfo/detail/128/0/")
					// trainInfoTextの内容を送信
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(trainInfoText)).Do(); err != nil {
						log.Print(err)
					}
				case "PSY":
					trainInfoText := `
					オッパン カンナムスタイル
					Eh sexy lady
					오-오-오-오 오빤 강남스타일`
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
