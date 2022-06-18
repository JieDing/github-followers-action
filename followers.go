package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
)

//go:embed templates
var fs embed.FS

type QL struct {
	Query string `json:"query"`
}

func main() {
	client := &http.Client{}
	url := "https://api.github.com/graphql"

	tpl, err := template.ParseFS(fs, "templates/followers.ql")
	if err != nil {
		log.Fatal(err)
	}

	var hasNextPage bool
	after := ""
	for {
		var b bytes.Buffer
		err = tpl.Execute(&b, map[string]interface{}{
			"Login": "JieDing",
			"After": after,
		})
		if err != nil {
			log.Fatal(err)
		}
		ql := QL{
			Query: b.String(),
		}
		byteArr, err := json.Marshal(&ql)
		//fmt.Println(string(byteArr))
		if err != nil {
			log.Fatal(err)
		}
		reader := bytes.NewReader(byteArr)
		req, err := http.NewRequest("POST", url, reader)
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Add("Authorization", "Bearer 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")

		rep, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		data, err := ioutil.ReadAll(rep.Body)
		rep.Body.Close()
		if err != nil {
			log.Fatal(err)
		}
		ret := gjson.GetBytes(data, "data.user.followers")
		pageInfo := ret.Get("pageInfo")
		hasNextPage = pageInfo.Get("hasNextPage").Bool()
		after = "after: \"" + pageInfo.Get("endCursor").String() + "\""

		fmt.Println(ret.Get("pageInfo.hasNextPage"))
		if !hasNextPage {
			break
		}
	}
	//after: "Y3Vyc29yOnYyOpK5MjAyMi0wMi0wOVQxMDoyOToxNiswODowMM4FVwgv"
	/*err = tpl.Execute(&b, map[string]interface{}{
		"Login": "JieDing",
		"After": "",
	})
	if err != nil {
		log.Fatal(err)
	}

	ql := QL{
		Query: b.String(),
	}
	//bytes.new
	byteArr, err := json.Marshal(&ql)
	//fmt.Println(string(byteArr))
	if err != nil {
		log.Fatal(err)
	}
	reader := bytes.NewReader(byteArr)
	req, err := http.NewRequest("POST", url, reader)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Add("Authorization", "Bearer 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")

	rep, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	data, err := ioutil.ReadAll(rep.Body)
	rep.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	ret := gjson.GetBytes(data, "data.user.followers")
	fmt.Println(ret.Get("pageInfo.hasNextPage"))*/
	//log.Printf("%s", data)
}
