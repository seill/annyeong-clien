package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tebeka/selenium"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const SLEEP_FETCH_ARTICLE = 1 * time.Second

type CrawlerService struct {
	BaseUrl string
	Driver  selenium.WebDriver
}

type Article struct {
	Link        string
	BoardName   string
	Title       string
	Hits        int
	CreatedTime string
	Ip          string
	Content     string
	Images      []Image
	Source      string
	Comments    []Comment
}

type Comment struct {
	Nickname    string
	Like        int
	Content     string
	Link        string `json:",omitempty"`
	Ip          string
	CreatedTime string
	Children    []Comment
}

type Image struct {
	Filename string
	Url      string
}

func NewCrawlerService(baseUrl string, driver selenium.WebDriver) *CrawlerService {
	return &CrawlerService{
		BaseUrl: baseUrl,
		Driver:  driver,
	}
}

func (s *CrawlerService) Login(id *string, password *string) (err error) {
	var accountSectionElement selenium.WebElement
	var wg sync.WaitGroup

	// open the Base URL
	err = s.Driver.Get(fmt.Sprintf("%s%s", s.BaseUrl, "/"))
	if nil != err {
		return
	}

	// for login
	accountSectionElement, err = s.Driver.FindElement(selenium.ByClassName, "account_section")
	if nil != err {
		return
	}

	if nil != id {
		var userIdElement selenium.WebElement

		userIdElement, err = accountSectionElement.FindElement(selenium.ByName, "userId")
		if nil != err {
			return
		}
		_ = userIdElement.SendKeys(*id)
	}

	if nil != password {
		var userPasswordElement selenium.WebElement
		userPasswordElement, err = accountSectionElement.FindElement(selenium.ByName, "userPassword")
		if nil != err {
			return
		}
		_ = userPasswordElement.SendKeys(*password)
	}

	if nil != id && nil != password {
		var loginButtonElement selenium.WebElement

		loginButtonElement, err = accountSectionElement.FindElement(selenium.ByName, "로그인하기")
		if nil != err {
			return
		}
		_ = loginButtonElement.Click()
	}

	wg.Add(1)

	// wait for login. Wait, WaitWithTimeout, WaitWithTimeoutAndInterval not working(?)
	go func(wg *sync.WaitGroup) {
		var err error

		for {
			_, err = s.Driver.FindElement(selenium.ByClassName, "account_name_group")
			if nil != err {
				log.Println("not found account_name_group")
				time.Sleep(1 * time.Second)
				continue
			}
			break
		}

		wg.Done()
	}(&wg)

	wg.Wait()
	log.Println("Login Success")

	return
}

func (s *CrawlerService) Archive(id *string, password *string) (err error) {
	var listMyArticleLink []string
	var listMyArticle []Article
	var bytesMyArticle []byte

	err = s.Login(id, password)
	if nil != err {
		return
	}

	log.Println("Archive: My Article")

	// for my article
	listMyArticleLink = []string{}
	listMyArticle = []Article{}

	for pageIndex := 0; pageIndex < 1; pageIndex++ {
		var urlMyArticle string
		var myArticleContainerElement selenium.WebElement
		var listMyArticleElement []selenium.WebElement

		urlMyArticle = fmt.Sprintf("%s/mypage/myArticle?&type=articles&sk=title&sv=&po=%d", s.BaseUrl, pageIndex)
		log.Println(urlMyArticle)

		err = s.Driver.Get(urlMyArticle)
		if nil != err {
			return
		}

		// my article container
		myArticleContainerElement, err = s.Driver.FindElement(selenium.ByClassName, "list_myArticle")
		if nil != err {
			return
		}

		// list of my article
		listMyArticleElement, err = myArticleContainerElement.FindElements(selenium.ByClassName, "list_item")
		if nil != err {
			return
		}

		if len(listMyArticleElement) == 0 {
			break
		}

		for _, articleElement := range listMyArticleElement {
			var titleElement, aTag selenium.WebElement

			// title
			titleElement, err = articleElement.FindElement(selenium.ByClassName, "list_title")
			if nil != err {
				return
			}

			// link
			aTag, err = titleElement.FindElement(selenium.ByTagName, "a")
			if err == nil {
				var link string

				link, err = aTag.GetAttribute("href")
				if nil != err {
					return
				}

				listMyArticleLink = append(listMyArticleLink, link)
			}
		}
	}

	for indexMyArticle, link := range listMyArticleLink {
		var article Article

		log.Printf("%d of %d \n", indexMyArticle+1, len(listMyArticleLink))

		article, err = s.FetchArticle(link)
		if nil != err {
			return
		}

		article.Link = link

		// download images
		for indexImage, image := range article.Images {
			if value, err := s.DownloadImage(image.Url); nil == err {
				article.Images[indexImage].Filename = value
			} else {
				log.Println(err)
				// continue even if there is an error
			}
		}

		listMyArticle = append(listMyArticle, article)
	}

	bytesMyArticle, err = json.Marshal(listMyArticle)
	if nil != err {
		return
	}

	// save to file
	err = os.WriteFile("annyeong-clien.json", bytesMyArticle, 0644)

	return
}

func (s *CrawlerService) Save(article Article) (err error) {
	return
}

func (s *CrawlerService) FetchArticle(link string) (article Article, err error) {
	var _err error
	var boardNameElement, titleElement, hitsElement, postAuthorElement, contentElement, sourceElement, postCommentContainerElement selenium.WebElement
	var imageElements []selenium.WebElement

	log.Println(fmt.Sprintf("Fetch Article: %s", link))

	err = s.Driver.Get(link)
	if nil != err {
		log.Fatal("Error:", err)
	}

	// board name
	boardNameElement, _err = s.Driver.FindElement(selenium.ByClassName, "board_name")
	if _err == nil {
		boardNameH2Element, _err := boardNameElement.FindElement(selenium.ByTagName, "h2")
		if _err == nil {
			article.BoardName, _ = boardNameH2Element.Text()
		}
	}

	// title
	titleElement, _err = s.Driver.FindElement(selenium.ByClassName, "post_title")
	if _err == nil {
		titleSpanElement, _ := titleElement.FindElement(selenium.ByTagName, "span")
		if _err == nil {
			article.Title, _ = titleSpanElement.Text()
		}
	}

	// hits
	hitsElement, _err = s.Driver.FindElement(selenium.ByClassName, "view_count")
	if _err == nil {
		if hits, _err := hitsElement.Text(); _err == nil {
			if value, _err := strconv.ParseInt(strings.ReplaceAll(hits, ",", ""), 10, 64); _err == nil {
				article.Hits = int(value)
			}
		}
	}

	postAuthorElement, _err = s.Driver.FindElement(selenium.ByClassName, "post_author")
	if _err == nil {
		postAuthorSpanElements, _err := postAuthorElement.FindElements(selenium.ByTagName, "span")
		if _err == nil {
			// created time
			article.CreatedTime, _ = postAuthorSpanElements[0].Text()

			if len(postAuthorSpanElements) > 4 {
				// ip
				article.Ip, _ = postAuthorSpanElements[3].Text()
			}
		}
	}

	// content
	contentElement, _err = s.Driver.FindElement(selenium.ByClassName, "post_content")
	if _err == nil {
		article.Content, _ = contentElement.Text()
	}

	// images
	imageElements, _err = contentElement.FindElements(selenium.ByTagName, "img")
	if _err == nil {
		for _, imageElement := range imageElements {
			src, _ := imageElement.GetAttribute("src")

			article.Images = append(article.Images, Image{
				Filename: "",
				Url:      src,
			})
		}
	}

	// source
	sourceElement, _err = s.Driver.FindElement(selenium.ByClassName, "attached_source")
	if _err == nil {
		article.Source, _ = sourceElement.Text()
		article.Source = strings.ReplaceAll(article.Source, "출처 :", "")
	}

	// comment
	postCommentContainerElement, _err = s.Driver.FindElement(selenium.ByClassName, "post_comment")
	if _err == nil {
		commentElements, _err := postCommentContainerElement.FindElements(selenium.ByClassName, "comment_row")

		if _err == nil {
			var commentPrevious *Comment

			for _, commentElement := range commentElements {
				var nicknameElement, likeElement, commentViewElement, commentDateElement, commentIpElement selenium.WebElement

				comment := Comment{}

				// nickname
				nicknameElement, _err = commentElement.FindElement(selenium.ByClassName, "contact_name")
				if _err == nil {
					comment.Nickname, _ = nicknameElement.Text()
				}

				if 0 == strings.Compare("", comment.Nickname) {
					imgNicknameElement, _err := commentElement.FindElement(selenium.ByTagName, "img")
					if _err == nil {
						comment.Nickname, _ = imgNicknameElement.GetAttribute("alt")
					}
				}

				// like
				likeElement, _err = commentElement.FindElement(selenium.ByClassName, "comment_symph")
				if _err == nil {
					if likeText, _err := likeElement.Text(); _err == nil {
						if value, _err := strconv.ParseInt(likeText, 10, 64); _err == nil {
							comment.Like = int(value)
						}
					}
				}

				// content
				commentViewElement, _err = commentElement.FindElement(selenium.ByClassName, "comment_view")
				if _err == nil {
					comment.Content, _ = commentViewElement.Text()
				}

				// created time
				commentDateElement, _err = commentElement.FindElement(selenium.ByClassName, "comment_time")
				if _err == nil {
					comment.CreatedTime, _ = commentDateElement.GetAttribute("textContent")
					comment.CreatedTime = strings.ReplaceAll(comment.CreatedTime, "\n", "")
					comment.CreatedTime = strings.ReplaceAll(comment.CreatedTime, "\t", "")
					comment.CreatedTime = comment.CreatedTime[8:]
				}

				// ip
				commentIpElement, _err = commentElement.FindElement(selenium.ByClassName, "ip_address")
				if _err == nil {
					comment.Ip, _ = commentIpElement.GetAttribute("textContent")
				}

				if value, _err := commentElement.GetAttribute("class"); _err == nil {
					if strings.Contains(value, "re") {
						if nil != commentPrevious {
							commentPrevious.Children = append(commentPrevious.Children, comment)
						}
					} else {
						article.Comments = append(article.Comments, comment)
						commentPrevious = &article.Comments[len(article.Comments)-1]
					}
				}
			}
		}
	}

	time.Sleep(SLEEP_FETCH_ARTICLE)

	return
}

func (s *CrawlerService) Delete(id *string, password *string) (err error) {
	return
}

func (s *CrawlerService) DownloadImage(imageUrl string) (filename string, err error) {
	var response *http.Response
	var file *os.File

	// Create a GET request to fetch the image
	response, err = http.Get(imageUrl)
	if nil != err {
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)

	// Check if the request was successful
	if response.StatusCode != http.StatusOK {
		err = errors.New(fmt.Sprintf("Failed to download image: %s", response.Status))
		return
	}

	// Create a new file to save the image
	filename = filepath.Base(imageUrl)
	filename = filename[:strings.Index(filename, "?")]
	filename = fmt.Sprintf("./images/%s", filename)
	file, err = os.Create(filename)
	if nil != err {
		return
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	// Copy the image data from the response body to the file
	_, err = io.Copy(file, response.Body)
	if nil != err {
		return
	}

	log.Println("Image downloaded successfully.")

	return
}
