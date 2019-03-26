package main

import (
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/tealeg/xlsx"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func main1() {
	envTelegramToken := "TELEGRAM_TOKEN"
	log.Println("TELEGRAM_TOKEN: ", os.Getenv(envTelegramToken))
	bot, err := tgbotapi.NewBotAPI("...")

	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	if err != nil {
		log.Panic(err)
	}
	// В канал updates будут приходить все новые сообщения.
	for update := range updates {
		log.Println("update: ", update)
		log.Println("Document: ", update.Message.Document)
		log.Println("FileID: ", update.Message.Document.FileID)
		log.Println("FileName: ", update.Message.Document.FileName)
		log.Println("FileSize: ", update.Message.Document.FileSize)

		fileWrapper, err := bot.GetFile(tgbotapi.FileConfig{FileID: update.Message.Document.FileID})
		if (err != nil) {
			panic(err);
		}
		fileLink := fileWrapper.Link(os.Getenv(envTelegramToken))
		log.Println("link: ", fileLink)

		resp, err := http.Get(fileLink)
		if (err != nil) {
			panic(resp)
		}
		log.Println("Response: ", resp)
		log.Println("Body: ", resp.Body)
		log.Println("Content length: ", resp.ContentLength)
		arrBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		xlFile, err := xlsx.OpenBinary(arrBytes)

		if (err != nil) {
			panic(err)
		}

		log.Println(xlFile)
		log.Println(xlFile.Sheets)
		log.Println(xlFile.Sheets[0])
		log.Println(xlFile.Sheets[0].Rows)
		log.Println(xlFile.Sheets[0].Rows[0])
		log.Println(xlFile.Sheets[0].Rows[1].Cells)
		log.Println(xlFile.Sheets[0].Rows[1].Cells[1])
		log.Println("content xlsx: ", xlFile.Sheets[0].Rows[1].Cells[1].Value)

		//tgbotapi.

		// Создав структуру - можно её отправить обратно боту
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID
		bot.Send(msg)
	}

}

func DownloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

type RowIn struct {
	Pair     string
	Type     string
	AvgPrice *big.Float
	Amount   *big.Float
	Total    *big.Float
	Status   string
}

type RowOut struct {
	Pair      string
	AmountIn  []*big.Float
	PriceIn   []*big.Float
	AmountOut []*big.Float
	PriceOut  []*big.Float
	Profit    *big.Float
}

func main() {
	excelFileName := "/home/gg/Downloads/OrderHistory.xlsx"
	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		log.Fatal("Can't open file")
	}

	rowsIn := parse(xlFile)
	rowsOut := format(rowsIn)

	var file *xlsx.File
	var sheet *xlsx.Sheet
	file = xlsx.NewFile()
	sheet, err = file.AddSheet("Sheet1")
	if err != nil {
		log.Printf(err.Error())
	}

	maxNumIn, maxNumOut := maxNumAmountIn(rowsOut), maxNumAmountOut(rowsOut)
	head := sheet.AddRow()
	header(head, maxNumIn, maxNumOut)

	for k, v := range rowsOut {
		//log.Println(v)
		r := sheet.AddRow()
		if k == 0 {
			r.AddCell().SetFloat(258000)
		} else {
			r.AddCell()
		}
		setRow(r, v, maxNumIn, maxNumOut)
	}

	err = file.Save("/home/gg/Downloads/OrderHistoryReadable.xlsx")
	if err != nil {
		log.Printf(err.Error())
	}
}

func maxNumAmountIn(rows []RowOut) int {
	max := 0
	for _, v := range rows {
		if max < len(v.AmountIn) {
			max = len(v.AmountIn)
		}
	}
	return max
}

func maxNumAmountOut(rows []RowOut) int {
	max := 0
	for _, v := range rows {
		if max < len(v.AmountOut) {
			max = len(v.AmountOut)
		}
	}
	return max
}

func header(row *xlsx.Row, maxNumAmountIn int, maxNumAmountOut int) {
	createCell(row, "Курс биткоина")
	cell := row.AddCell()
	cell.Value = ""
	createCell(row, "Заметки")
	for i := 0; i < maxNumAmountIn; i++ {
		createCell(row, "Кол-во покупки")
		createCell(row, "Цена покупки")
	}
	for i := 0; i < maxNumAmountOut; i++ {
		createCell(row, "Кол-во продажи")
		createCell(row, "Цена продажи")
	}
	createCell(row, "Продажа - покупка")
	createCell(row, "Прибыль")
}

func createCell(r *xlsx.Row, value string) *xlsx.Cell {
	cell := r.AddCell()
	cell.GetStyle().Alignment.WrapText = true
	cell.Value = value
	return cell
}

func setRow(row *xlsx.Row, rOut RowOut, maxNumAmountIn int, maxNumAmountOut int) {
	row.AddCell().Value = rOut.Pair
	row.AddCell().Value = ""
	for k, v := range rOut.AmountIn {
		row.AddCell().Value = v.Text('f', 8)
		row.AddCell().Value = rOut.PriceIn[k].Text('f', 8)
	}
	if len(rOut.AmountIn) < maxNumAmountIn {
		for i := len(rOut.AmountIn); i < maxNumAmountIn; i++ {
			row.AddCell().Value = ""
			row.AddCell().Value = ""
		}
	}
	for k, v := range rOut.AmountOut {
		row.AddCell().Value = v.Text('f', 8)
		row.AddCell().Value = rOut.PriceOut[k].Text('f', 8)
	}
	if len(rOut.AmountOut) < maxNumAmountOut {
		for i := len(rOut.AmountOut); i < maxNumAmountOut; i++ {
			row.AddCell().Value = ""
			row.AddCell().Value = ""
		}
	}
	row.AddCell().SetFormula(rOut.Profit.Text('f', 8) + "*A2")
	if len(rOut.AmountOut) != 0 {
		row.AddCell().SetFormula(rOut.Profit.Text('f', 8) + "*A2")
	}
	//log.Println(rOut.Profit)
}

func parse(xlFile *xlsx.File) []RowIn {
	result := make([]RowIn, 0, 300)
	rows := xlFile.Sheets[0].Rows
	for i := len(rows) - 1; i >= 0; i-- {
		if strings.Contains(rows[i].Cells[1].Value, "BTC") && (rows[i].Cells[8].Value == "Filled" || rows[i].Cells[8].Value == "Partial Fill") {
			result = append(result, parseRow(rows[i]))
		}
	}
	return result
}

func parseRow(row *xlsx.Row) RowIn {
	return RowIn{
		Pair:     row.Cells[1].Value,
		Type:     row.Cells[2].Value,
		Amount:   StringToBigFloat(row.Cells[6].Value),
		AvgPrice: StringToBigFloat(row.Cells[5].Value),
		Total:    StringToBigFloat(row.Cells[7].Value),
		Status:   row.Cells[8].Value,
	}
}

func format(rowsIn []RowIn) []RowOut {
	result := make([]RowOut, 0, 300)

	setOfPairs := setOfPairs(rowsIn)
	for _, v := range setOfPairs {
		result = append(result, parseByPair(v, rowsIn)...)
	}
	return result
}

func parseByPair(pair string, rowsIn []RowIn) []RowOut {
	r, _ := parseByPairRecursive(pair, rowsIn, 0)
	return r
}

func parseByPairRecursive(pair string, rowsIn []RowIn, indexOffset int) ([]RowOut, int) {
	result := make([]RowOut, 0, 300)
	i := indexOffset
	for ; i < len(rowsIn); i++ {
		rOut, jOffset := createRowOut(rowsIn, pair, i)
		setProfit(rOut)
		if len(rOut.AmountIn) == 0 && len(rOut.AmountOut) == 0 {
			break
		}
		result = append(result, *rOut)

		i = jOffset + 1

		if i < len(rowsIn) {
			r, kOffset := parseByPairRecursive(pair, rowsIn, i)
			result = append(result, r...)
			i = kOffset
		}
	}
	return result, i + 1
}

func setOfPairs(rowsIn []RowIn) []string {
	result := make([]string, 0, 300)
	for _, v := range rowsIn {
		if !contains(result, v.Pair) {
			result = append(result, v.Pair)
		}
	}
	return result
}

func createRowOut(rowsIn []RowIn, pair string, indexOffset int) (*RowOut, int) {
	result := new(RowOut)
	result.Pair = pair
	i := indexOffset
	for ; i < len(rowsIn); i++ {
		r := rowsIn[i]
		if pair == r.Pair && r.Type == "BUY" {
			//if pair == "EOSBTC" {
			//	log.Println(r)
			//}
			result.AmountIn = append(result.AmountIn, r.Amount)
			result.PriceIn = append(result.PriceIn, r.AvgPrice)
			continue
		} else if pair == r.Pair && r.Type == "SELL" {
			//if pair == "EOSBTC" {
			//	log.Println(r)
			//}
			result.AmountOut = append(result.AmountOut, r.Amount)
			result.PriceOut = append(result.PriceOut, r.AvgPrice)
			if len(result.AmountIn) == len(result.AmountOut) || amountIsEquals(result.AmountIn, result.AmountOut) {
				return result, i
			}
		}
	}
	i--
	return result, i
}

func amountIsEquals(am1, am2 []*big.Float) bool {
	dif := new(big.Float)
	f, _ := dif.Sub(sum(am2), sum(am1)).Float64()
	return f > -2
	//return sum(am1).Cmp(sum(am2)) == 0
}

func sum(am []*big.Float) *big.Float {
	result := new(big.Float)
	for _, v := range am {
		result.Add(result, v)
	}
	return result
}

func setProfit(r *RowOut) {
	purchase, sell := createPrice(), createPrice()
	for i := 0; i < len(r.AmountIn); i++ {
		pricePurchase := createPrice()
		pricePurchase.Mul(r.AmountIn[i], r.PriceIn[i])
		purchase.Add(purchase, pricePurchase)

		//log.Println(pricePurchase.Text('f', 8), purchase.Text('f', 8))
	}

	//log.Println("sdfdf")

	for i := 0; i < len(r.AmountOut); i++ {
		priceSell := createPrice()
		priceSell.Mul(r.AmountOut[i], r.PriceOut[i])
		sell.Add(sell, priceSell)

		//log.Println(priceSell.Text('f', 8), sell.Text('f', 8))
	}
	//log.Println(purchase.Text('f', 8), sell.Text('f', 8))

	result := createPrice()
	result.Sub(sell, purchase)
	r.Profit = result
}

func createPrice() *big.Float {
	p := new(big.Float)
	p.SetPrec(16)
	p.SetMode(big.ToNearestEven)
	return p
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Float64toBigFloat(num float64) *big.Float {
	return new(big.Float).SetFloat64(num)
}

func IntToBigFloat(num int) *big.Float {
	return new(big.Float).SetInt64(int64(num))
}

func StringToBigFloat(str string) *big.Float {
	res, _ := new(big.Float).SetString(str)
	return res
}

func FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 6, 64)
}
