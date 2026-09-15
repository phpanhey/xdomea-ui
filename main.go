package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Node struct {
	Type     string `json:"type"`
	Label    string `json:"label"`
	Children []Node `json:"children,omitempty"`
}

func main() {
	http.HandleFunc("/", ServeIndex)
	http.HandleFunc("/generatexdomea", Generatexdomea)
	fmt.Println("Server running on :8080")
	fmt.Println("/generatexdomea endpoint is exposed")
	fmt.Println("ui is on http://127.0.0.1:8080")
	http.ListenAndServe(":8080", nil)

	/*
	   		jsondata := `[
	   	   {
	   	        "type": "akte",
	   	        "label": "Akte 1",
	   	        "children": [
	   	            {
	   	                "type": "vorgang",
	   	                "label": "Vorgang 1",
	   	                "children": [
	   	                    {
	   	                        "type": "untervorgang",
	   	                        "label": "Untervorgang 1",
	   	                        "children": [
	   	                            {
	   	                                "type": "dokument",
	   	                                "label": "Dokument 1"
	   	                            }
	   	                        ]
	   	                    }
	   	                ]
	   	            }
	   	        ]
	   	    }

	   ]`

	   	var nodes []Node
	   	if err := json.Unmarshal([]byte(jsondata), &nodes); err != nil {
	   		panic("json could not be parsed")
	   	}

	   	xml := filesMap["base_prefix"]

	   	for _, node := range nodes {
	   		xml += nodeToXML(node, filesMap)
	   	}

	   	xml += filesMap["base_postfix"]

	   	exportFile := "output.xml"
	   	stringToFile(exportFile, xml)
	   	format(exportFile)
	*/
}

func ServeIndex(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintln(w, readFile("template/index.html"))
	if err != nil {
		return
	}
}

func nodeToXML(n Node, filesMap map[string]string) string {

	nodeType := n.Type

	prefixContent := filesMap[nodeType+"_prefix"]
	postfixContent := filesMap[nodeType+"_postfix"]

	xml := prefixContent

	for _, child := range n.Children {
		xml += nodeToXML(child, filesMap)
	}

	xml += postfixContent

	return xml
}

func Generatexdomea(w http.ResponseWriter, r *http.Request) {

	filesMap := readFiles()
	fmt.Println("Generating xdomea endpoint")

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var nodes []Node
	if err := json.Unmarshal([]byte(body), &nodes); err != nil {
		panic("json could not be parsed")
	}

	xml := filesMap["base_prefix"]

	for _, node := range nodes {
		xml += nodeToXML(node, filesMap)
	}

	xml += filesMap["base_postfix"]

	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(format(xml)))
}

func readFiles() map[string]string {
	filesMap := make(map[string]string)
	entries, err := os.ReadDir("./template")
	if err != nil {
		panic(err)
	}
	for _, e := range entries {
		path := "template/"
		fileName := e.Name()

		fileContent := readFile(path + fileName)

		filesMap[getKeyFromFileName(fileName)] = fileContent
	}
	return filesMap
}

func getKeyFromFileName(fileName string) string {
	return strings.Split(fileName, ".")[0]
}

func format(xmlData string) string {
	var buf bytes.Buffer

	decoder := xml.NewDecoder(strings.NewReader(xmlData))
	encoder := xml.NewEncoder(&buf)
	encoder.Indent("", "  ")

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		if err := encoder.EncodeToken(token); err != nil {
			panic(err)
		}
	}

	if err := encoder.Flush(); err != nil {
		panic(err)
	}

	return buf.String()
}

func readFile(path string) string {
	fileContent, error := os.ReadFile(path)
	if error != nil {
		panic("file not found")
	}
	return string(fileContent)
}
