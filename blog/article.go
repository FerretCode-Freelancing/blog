package blog

import (
	"html/template"
	"net/http"

	"github.com/ferretcode-freelancing/blog/admin"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type article struct {
	Article htmlArticle
}

func Article(w http.ResponseWriter, r *http.Request, db *gorm.DB) error {
	var a admin.Post

	err := db.Model(&admin.Post{}).Where("id = ?", chi.URLParam(r, "id")).First(&a).Error

	if err != nil {
		return err
	}

	tmpl, err := template.ParseFiles("templates/blog/article.html")

	if err != nil {
		return err
	}

	tmpl.Execute(w, article{
		Article: htmlArticle{
			Id:          a.Id,
			Image:       a.Image,
			Title:       a.Title,
			Content:     template.HTML(a.Content),
			Description: a.Description,
			Tags:        a.Tags,
		},
	})

	return nil
}
