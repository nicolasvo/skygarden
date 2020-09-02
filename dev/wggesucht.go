package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
)

// Offer stores offer data
type Offer struct {
	Url              string
	OfferID          string
	Title            string
	RentFull         int
	Rent             int
	RentAdditional   int
	Deposit          int
	Address          string
	AvailabilityDate string
	CreationDate     string
	Area             int
	Rooms            int
	Description      []string
	Details          []string
}

func standardizeSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func crawl(url string) []Offer {
	offers := make([]Offer, 0, 200)
	c := colly.NewCollector()
	d := c.Clone()

	c.OnRequest(func(r *colly.Request) {
		// fmt.Println("Visiting", r.URL.String())
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	// Parse result page
	c.OnXML("//div[@class='wgg_card offer_list_item ']//h3//a", func(e *colly.XMLElement) {
		urlOffer := "https://www.wg-gesucht.de/" + e.Attr("href")
		d.Visit(urlOffer)
	})

	d.OnXML("//div[@class='panel panel-default']", func(e *colly.XMLElement) {
		url := e.Request.URL.String()
		title := e.ChildText("//h1[@class='headline headline-detailed-view-title']")
		rentFull, _ := strconv.Atoi(strings.Replace(e.ChildText("(//div[@class='row'])[1]/div[3]/h2"), "€", "", -1))
		rentAdditional, _ := strconv.Atoi(strings.Replace(e.ChildText("(//div[@class='row'])[3]/div/table/tbody/tr[2]/td[2]/b"), "€", "", -1))
		deposit, _ := strconv.Atoi(strings.Replace(e.ChildText("(//div[@class='row'])[3]/div/table/tbody/tr[4]/td[2]/b"), "€", "", -1))
		address := standardizeSpaces(e.ChildText("(//div[@class='row'])[3]/div[2]/a"))
		availabilityDate := e.ChildText("(//div[@class='row'])[3]/div[3]/p[1]/b")
		creationDate := e.ChildText("(//div[@class='row'])[3]/div[3]/b")
		area, _ := strconv.Atoi(strings.Replace(e.ChildText("(//div[@class='row'])[1]/div[2]/h2"), "m²", "", -1))
		rooms, _ := strconv.Atoi(e.ChildText("(//div[@class='row'])[1]/div[4]/h2"))
		details := e.ChildTexts("(//div[@class='row'])[6]/div/div//div[@class='col-xs-6 col-sm-4 text-center print_text_left']")
		descriptionSection := e.ChildTexts("//*[@id='ad_description_text']//div[@class='wordWrap']//h3")
		descriptionContent := e.ChildTexts("//*[@id='ad_description_text']//div[@class='wordWrap']//p")
		description := make([]string, 0, 10)
		for i := range descriptionSection {
			description = append(description, descriptionSection[i]+": "+descriptionContent[i])
		}
		temp := strings.Fields(e.ChildText("//div[@class='row bottom_contact_box']/div/div/div[1]/div[2]/div[2]/div"))
		offerID := temp[len(temp)-1]

		offer := Offer{
			Url:              url,
			OfferID:          offerID,
			Title:            title,
			RentFull:         rentFull,
			RentAdditional:   rentAdditional,
			Deposit:          deposit,
			Address:          address,
			AvailabilityDate: availabilityDate,
			CreationDate:     creationDate,
			Area:             area,
			Rooms:            rooms,
			Description:      description,
			Details:          details,
		}
		offers = append(offers, offer)
	})

	c.Visit(url)

	return offers
}

func main() {
	url := "https://www.wg-gesucht.de/wohnungen-in-Karlsruhe.68.2.1.0.html?offer_filter=1&city_id=68&noDeact=1&categories%5B%5D=2&rent_types%5B%5D=2&sMin=40&rMax=600"
	// _ = crawl(url)
	offers := crawl(url)
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	offersEncoded := enc.Encode(offers)
	fmt.Printf("%+v\n", offersEncoded)
}
