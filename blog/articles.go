package blog

import (
	"html/template"
	"net/http"

	"github.com/ferretcode-freelancing/blog/admin"
	"gorm.io/gorm"
)

type articleList struct {
	Articles []admin.Post
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

	tmpl.Execute(w, articleList{
		Articles: articles,
	})

	return nil
}
