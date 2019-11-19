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

const errorMessage = "大変申し訳ございません。エラーが発生しました。時間をおいて試してみて下さい。"

// ScrapingTrainInfo Yahoo!路線情報の各路線のページへアクセスし、運行情報のテキストをスクレイピングして返す.
func ScrapingTrainInfo(url string) string {
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

// ShapedTrainInfo Yahoo!路線情報のページをスクレイピングして、運行情報を教えるLINEメッセージを形成し出力
func ShapedTrainInfo(url string)
{
	response, err := http.Get()
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
	titleSelector := document.Find("div.labelLarge")
	areaText := titleSelector.Find("h1.title").Text()
	timeText := titleSelector.Find("span.subText").Text()

	// 京成本線、都営浅草線の運行情報を取得
	keiseiMainLineInfo := ScrapingTrainInfo("https://transit.yahoo.co.jp/traininfo/detail/96/0/")
	asakusaLineInfo := ScrapingTrainInfo("https://transit.yahoo.co.jp/traininfo/detail/128/0/")
	
	// 送信するメッセージを形成
	outgoingMessage := 
		"【" + areaText + "】\n" +
		timeText + "\n"
		"・京成本線\n" +
		keiseiMainLineInfo + "\n"
		"・都営浅草線\n" +
		asakusaLineInfo

	if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(outgoingMessage)).Do(); err != nil {
		log.Print(err)
	}
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
				case "運行情報":
					// 運行情報をスクレイピングし、テキストメッセージを送信
					ShapedTrainInfo("https://transit.yahoo.co.jp/traininfo/area/4/")
				case "天気":
					// 都営浅草線の運行情報をスクレイピング
					trainInfoText := ScrapingTrainInfo("https://transit.yahoo.co.jp/traininfo/detail/128/0/")
					// trainInfoTextの内容を送信
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(trainInfoText)).Do(); err != nil {
						log.Print(err)
					}
				case "PSY":
					trainInfoText := 
					"オッパン カンナムスタイル\n" +
					"Eh sexy lady\n" +
					"오-오-오-오 오빤 강남스타일"
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
