package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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
func ShapedTrainInfo(url string) string {
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
	titleSelector := document.Find("div.labelLarge")
	areaText := titleSelector.Find("h1.title").Text()
	timeText := titleSelector.Find("span.subText").Text()

	// 京成本線、都営浅草線の運行情報を取得
	keiseiMainLineInfo := ScrapingTrainInfo("https://transit.yahoo.co.jp/traininfo/detail/96/0/")
	asakusaLineInfo := ScrapingTrainInfo("https://transit.yahoo.co.jp/traininfo/detail/128/0/")

	// 送信するメッセージを形成
	outgoingMessage :=
		"【" + areaText + "】\n" +
			timeText + "\n" +
			"・京成本線\n" +
			keiseiMainLineInfo + "\n" +
			"・都営浅草線\n" +
			asakusaLineInfo

	return outgoingMessage
}

// ConvertToWeatherEmoji 天気文字列から絵文字文字列を作成
func ConvertToWeatherEmoji(weatherText string) string {
	emojis := ""
	// runeに変換する
	runed := []rune(weatherText)

	// 1文字目: ベースの天気
	switch ( string(runed[0]) ) {
	case "晴":
		emojis += "☀️"
	case "曇":
		emojis += "☁️"
	case "雨":
		emojis += "☂️"
	case "雪":
		emojis += "⛄️"
	// イレギュラーな天気は未実装で
	default:
		return "❓"
	}

	// サブ天気がなければ終了
	if len(runed) < 4 { return emojis }

	// 2-3文字目: サブ天気の頻度 ex. のち 時々 一時
	switch ( string(runed[1:3]) ) {
	case "のち":
		emojis += "=>"
	case "時々":
		emojis += "/"
	case "一時":
		emojis += ":"
	// イレギュラーなパターンは未実装で
	default:
		emojis += "-"
	}

	// 4文字目: サブ天気
	switch ( string(runed[3]) ) {
	case "晴":
		emojis += "☀️"
	case "曇":
		emojis += "☁️"
	case "雨":
		emojis += "☂️"
	case "雪":
		emojis += "⛄️"
	// イレギュラーな天気は未実装で
	default:
		return "❓"
	}

	return emojis
}

// ShapedWeatherInfo 千葉県の天気予報を教えるLINEメッセージを形成し出力
func ShapedWeatherInfo(url string) string {
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

	outgoingMessage := ""

	forecastSelector := document.Find("div.forecastCity")
	// innerForecastSelector ∋ 今日と明日の天気情報が記述されたdiv要素
	innerForecastSelector := forecastSelector.Find("div")
	// 今日と明日分の天気情報をスクレイピングして、メッセージ化
	innerForecastSelector.Each(func(index int, s *goquery.Selection) {
		date 			:= s.Find("p.date").First().Text()
		weather 		:= s.Find("p.pict").First().Text()
		tempSelector 	:= s.Find("ul.temp").First()
		high 			:= tempSelector.Find("li.high > em").Text()
		low 			:= tempSelector.Find("li.low > em").Text()

		outgoingMessage +=
		"【" + date + " の天気】\n" +
		"予報: " + weather + " " + ConvertToWeatherEmoji(weather) + "\n" +
		"最高気温: " + high + "度\n" +
		"最低気温: " + low + "度" +
		"\n"
	})
	
	return outgoingMessage
}

// GetJstTime time.Time型の変数tを日本時間に変換して返します
func GetJstTime(t time.Time) time.Time {
	// convert to UTC time
	tUTC := t.UTC()

	// get Jst timeZone
	jstTimezone := time.FixedZone("Asia/Tokyo", 9*60*60)
	return tUTC.In(jstTimezone)
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
					outgoingMessage := ShapedTrainInfo("https://transit.yahoo.co.jp/traininfo/area/4/")
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(outgoingMessage)).Do(); err != nil {
						log.Print(err)
					}
				case "天気":
					// 千葉県の天気情報をスクレイピングし、テキストメッセージを送信
					outgoingMessage := ShapedWeatherInfo("https://weather.yahoo.co.jp/weather/jp/12/4510.html")
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(outgoingMessage)).Do(); err != nil {
						log.Print(err)
					}
				case "PSY":
					outgoingMessage :=
						"オッパン カンナムスタイル\n" +
						"Eh sexy lady\n" +
						"오-오-오-오 오빤 강남스타일"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(outgoingMessage)).Do(); err != nil {
						log.Print(err)
					}
				default:
					outgoingMessage := ""
					// 日本時間の現在時刻を取得
					jstNowTime := GetJstTime(time.Now())
					// .Hour()で分岐
					switch {
					case jstNowTime.Hour() >= 5 && jstNowTime.Hour() < 12:
						outgoingMessage += "おはようございます。\n"
					case jstNowTime.Hour() >= 12 && jstNowTime.Hour() < 18:
						outgoingMessage += "こんにちは。\n"
					case jstNowTime.Hour() >= 18 && jstNowTime.Hour() < 24:
						outgoingMessage += "こんばんは。\n"
					default:
						outgoingMessage += "😪💤\n"
					}
					// 運行情報と天気を両方表示する。
					outgoingMessage += ShapedTrainInfo("https://transit.yahoo.co.jp/traininfo/area/4/")
					outgoingMessage += "\n\n"
					outgoingMessage += ShapedWeatherInfo("https://weather.yahoo.co.jp/weather/jp/12/4510.html")
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(outgoingMessage)).Do(); err != nil {
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
