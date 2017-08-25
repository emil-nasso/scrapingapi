package main

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/gin-gonic/gin"
	"github.com/PuerkitoBio/goquery"
	"log"
	"strings"
)

type Field struct {
	Name      string
	Selector  string
	Type      string
	SubFields []Field
}

type Endpoint struct {
	Path      string
	Source    string
	Variables []string
	Fields    []Field
}

func main() {
	r := gin.Default()
	initConfig()

	var endpoints []Endpoint
	viper.UnmarshalKey("endpoints", &endpoints)

	for _, endpoint := range endpoints {
		r.GET(endpoint.Path, endpointHandler(endpoint))
	}

	r.Run()
}

func endpointHandler(endpoint Endpoint) gin.HandlerFunc {
	fn := func(c *gin.Context) {

		sourceUrl := endpoint.Source
		for _, variable := range endpoint.Variables {
			sourceUrl = strings.Replace(sourceUrl, "%"+variable+"%", c.Query(variable), -1)
		}
		doc, err := goquery.NewDocument(sourceUrl)

		if err != nil {
			log.Fatal(err)
		}

		response := gin.H{
			"source": sourceUrl,
		}

		for _, field := range endpoint.Fields {
			if len(field.SubFields) == 0 {
				response[field.Name] = getValues(doc.Selection, field)
			} else {
				response[field.Name] = getSubFields(doc.Selection, field)
			}
		}

		c.JSON(200, response)
	}
	return gin.HandlerFunc(fn)
}

func getValues(s *goquery.Selection, field Field) interface{} {
	elements := s.Find(field.Selector)

	if field.Type == "string" {
		return strings.TrimSpace(elements.First().Text());
	} else if field.Type == "count" {
		return elements.Length();
	} else {
		values := make([]string, 0)
		elements.Each(func(i int, s *goquery.Selection) {
			values = append(values, strings.TrimSpace(s.Contents().Text()))
		})
		return values
	}
}

func getSubFields(s *goquery.Selection, field Field) []gin.H {
	result := []gin.H{}
	s.Find(field.Selector).Each(func(i int, s *goquery.Selection) {
		subElement := gin.H{}
		for _, subField := range field.SubFields {
			subElement[subField.Name] = getValues(s, subField)
		}
		result = append(result, subElement)
	})
	return result
}

func initConfig() {

	viper.SetConfigName("config")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()

	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}

}
