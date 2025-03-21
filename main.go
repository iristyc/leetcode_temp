package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var userName = "iristyc"


func main() {
	easy, medium, hard := getQuestionProgressInfo()
	mdContent := readFile()
	mdContent = strings.ReplaceAll(mdContent, `[[1]]`, strconv.Itoa(easy+medium+hard))
	mdContent = strings.ReplaceAll(mdContent, `[[2]]`, strconv.Itoa(easy))
	mdContent = strings.ReplaceAll(mdContent, `[[3]]`, strconv.Itoa(medium))
	mdContent = strings.ReplaceAll(mdContent, `[[4]]`, strconv.Itoa(hard))
	
	fmt.Println(mdContent)
	createWriteFile(mdContent)
	updateGithub()
}

/**
 * 判断文件是否存在  存在返回 true 不存在返回false
 */
func checkFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

// push到github仓库
func updateGithub() {
	cmd := exec.Command("sh", "./auto.sh")
	stdout, err := cmd.StdoutPipe()
	cmd.Start()
	reader := bufio.NewReader(stdout)
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		fmt.Println(line)
	}
	checkErr(err)
}

// 生成README.md文件
func createWriteFile(mdContent string) {
	if !checkFileIsExist("README.MD") {
		f, err := os.Create("README.MD") //创建文件
		checkErr(err)
		f.Close()
	}
	f, err := os.OpenFile("README.MD", os.O_WRONLY|os.O_TRUNC, 0600)
	checkErr(err)
	defer f.Close()
	_, err = f.Write([]byte(mdContent))
	checkErr(err)
}

// 请求接口获取做题进度
func getQuestionProgressInfo() (easy, medium, hard int) {
	client := &http.Client{}
	// jsonStr := `{"operationName":"userQuestionProgress","variables":{"userSlug":"` + userName + `"},"query":"query userQuestionProgress($userSlug: String!) {\n  userProfileUserQuestionProgress(userSlug: $userSlug) {\n    numAcceptedQuestions {\n      difficulty\n      count\n}\n}\n}\n"}`
	jsonStr := `{
	   "operationName": "userQuestionProgress",
	   "variables": {"userSlug": "` + userName + `"},
	   "query": "query userQuestionProgress($userSlug: String!) {\n  userProfileUserQuestionProgressV2(userSlug: $userSlug) {\n    numAcceptedQuestions {\n      difficulty\n      count\n    }\n  }\n}"
	}`

	req, err := http.NewRequest("POST", "https://leetcode.com/graphql/", strings.NewReader(jsonStr))
	req.Header.Add("content-type", "application/json")
	checkErr(err)
	resp, err := client.Do(req)
	checkErr(err)
	defer resp.Body.Close()
	
	body, _ := ioutil.ReadAll(resp.Body)
	// 新增
	fmt.Println("Response Body: ", string(body))  // 打印出 API 的回應
	
	var mapResult map[string]interface{}
	err = json.Unmarshal(body, &mapResult)
	checkErr(err)
	// f1(&mapResult)
	easy, medium, hard = analysisProgressInfo(&mapResult)
	return
}

// 解析做题详情
// func analysisProgressInfo(mapResult *map[string]interface{}) (easy, medium, hard int) {
// 	data := (*mapResult)["data"]
// 	userProfileUserQuestionProgress := data.(map[string]interface{})["userProfileUserQuestionProgress"]
// 	numAcceptedQuestions := userProfileUserQuestionProgress.(map[string]interface{})["numAcceptedQuestions"]
// 	for _, v := range numAcceptedQuestions.([]interface{}) {
// 		m := v.(map[string]interface{})
// 		if m["difficulty"] == "EASY" {
// 			easy += int(m["count"].(float64))
// 		}
// 		if m["difficulty"] == "MEDIUM" {
// 			medium += int(m["count"].(float64))
// 		}
// 		if m["difficulty"] == "HARD" {
// 			hard += int(m["count"].(float64))
// 		}
// 	}
// 	// fmt.Println(easy, medium, hard)
// 	return
// }
func analysisProgressInfo(mapResult *map[string]interface{}) (easy, medium, hard int) {
    // 確認 mapResult 中的 "data" 欄位是否存在且為正確型別
    data, ok := (*mapResult)["data"].(map[string]interface{})
    if !ok {
        fmt.Println("Error: Unable to parse data field")
        return
    }

    // 確認 "userProfileUserQuestionProgressV2" 是否存在且為正確型別
    userProfileUserQuestionProgressV2, ok := data["userProfileUserQuestionProgressV2"].(map[string]interface{})
    if !ok {
        fmt.Println("Error: Unable to parse userProfileUserQuestionProgressV2 field")
        return
    }

    // 確認 "numAcceptedQuestions" 是否存在且為正確型別
    numAcceptedQuestions, ok := userProfileUserQuestionProgressV2["numAcceptedQuestions"].([]interface{})
    if !ok {
        fmt.Println("Error: Unable to parse numAcceptedQuestions field")
        return
    }

    // 遍歷 numAcceptedQuestions
    for _, v := range numAcceptedQuestions {
        m, ok := v.(map[string]interface{})
        if !ok {
            fmt.Println("Error: Invalid data structure")
            continue
        }

        // 確認 difficulty 和 count 欄位存在並進行計算
        if diff, ok := m["difficulty"].(string); ok {
            if count, ok := m["count"].(float64); ok {
                switch diff {
                case "EASY":
                    easy += int(count)
                case "MEDIUM":
                    medium += int(count)
                case "HARD":
                    hard += int(count)
                }
            } else {
                fmt.Println("Error: Invalid count value")
            }
        } else {
            fmt.Println("Error: Invalid difficulty value")
        }
    }

    return
}


// 读取模版文件
func readFile() string {
	data, err := ioutil.ReadFile("README-TEMP.md")
	checkErr(err)
	return string(data)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
