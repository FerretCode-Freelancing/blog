package blog

import (
	"html/template"
	"net/http"

	"github.com/ferretcode-freelancing/blog/admin"
	"gorm.io/gorm"
)

type articleList struct {
	Articles []htmlArticle
}

type htmlArticle struct {
	Id          string        `json:"id"`
	Image       string        `json:"image"`
	Title       string        `json:"title"`
	Content     template.HTML `json:"content"`
	Description string        `json:"description"`
	Tags        []string      `json:"tags"`
}

func Articles(w http.ResponseWriter, r *http.Request, db *gorm.DB) error {
	var articles []admin.Post

	err := db.Model(&admin.Post{}).Find(&articles).Error

	if err != nil {
		return err
	}

	tmpl, err := template.ParseFiles("templates/blog/articles.html")

	if err != nil {
		return err
	}

	var htmlArticles []htmlArticle

	for _, article := range articles {
		htmlArticles = append(htmlArticles, htmlArticle{
			Id:          article.Id,
			Image:       article.Image,
			Title:       article.Title,
			Content:     template.HTML(article.Content),
			Description: article.Description,
			Tags:        article.Tags,
		})
	}

	tmpl.Execute(w, articleList{
		Articles: htmlArticles,
	})

	return nil
}
