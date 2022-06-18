package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"flag"
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
	var login string
	var pat string
	flag.StringVar(&login, "u", "", "GitHub ID")
	flag.StringVar(&pat, "p", "", "Personal Access Token")
	flag.Parse()

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
			"Login": login,
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

		req.Header.Add("Authorization", "Bearer "+pat)

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

}
