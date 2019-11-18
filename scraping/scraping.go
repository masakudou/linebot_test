package scraping

import (
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
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
