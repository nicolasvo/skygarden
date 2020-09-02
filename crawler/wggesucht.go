package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gocolly/colly/v2"
)

// BodyRequest is our self-made struct to process JSON request from Client
type BodyRequest struct {
	Url string `json:"url"`
}

// BodyResponse is our self-made struct to build response for Client
type BodyResponse struct {
	// ResponseName string `json:"name"`
	Offers []Offer
}

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

// Handler function Using AWS Lambda Proxy Request
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// BodyRequest will be used to take the json response from client and build it
	bodyRequest := BodyRequest{
		Url: "",
	}

	// Unmarshal the json, return 404 if error
	err := json.Unmarshal([]byte(request.Body), &bodyRequest)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 404}, nil
	}

	// We will build the BodyResponse and send it back in json form
	bodyResponse := BodyResponse{
		Offers: crawl(bodyRequest.Url),
	}

	// Marshal the response into json bytes, if error return 404
	response, err := json.Marshal(&bodyResponse)
	if err != nil {
		return events.APIGatewayProxyResponse{Body: err.Error(), StatusCode: 404}, nil
	}

	//Returning response with AWS Lambda Proxy Response
	return events.APIGatewayProxyResponse{Body: string(response), StatusCode: 200}, nil
}

func main() {
	lambda.Start(Handler)
}
