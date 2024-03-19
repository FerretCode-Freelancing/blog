package admin

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Post struct {
	Id      string `json:"id"`
	Image   string `json:"image"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func NewPost(w http.ResponseWriter, r *http.Request, db *gorm.DB) error {
	tmpl, err := template.ParseFiles("templates/admin/new_post.html")

	if err != nil {
		return err
	}

	tmpl.Execute(w, nil)

	return nil
}

func CreateNewPost(w http.ResponseWriter, r *http.Request, db *gorm.DB) error {
	err := r.ParseMultipartForm(10 << 20)

	if err != nil {
		return err
	}

	fmt.Println(r.PostForm)

	post := Post{
		Id:      uuid.NewString(),
		Title:   r.PostForm.Get("title"),
		Content: r.PostForm.Get("content"),
	}

	file, fileHeader, err := r.FormFile("image")

	if err != nil {
		return err
	}

	dst, err := os.Create(fmt.Sprintf("./%s%s", post.Id, filepath.Ext(fileHeader.Filename)))

	if err != nil {
		return err
	}

	_, err = io.Copy(dst, file)

	if err != nil {
		return err
	}

	dst.Close()
	file.Close()

	err = uploadImage(post.Id)

	if err != nil {
		return err
	}

	post.Image = fmt.Sprintf("%s/%s", os.Getenv("STORAGE_BUCKET"), post.Id+".jpg")

	err = db.Create(&post).Error

	if err != nil {
		return err
	}

	w.WriteHeader(200)
	w.Write([]byte("The post was created."))

	return nil
}

func uploadImage(id string) error {
	ctx := context.Background()

	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               fmt.Sprintf("https://%s.r2.cloudflarestorage.com", os.Getenv("ACCOUNT_ID")),
			HostnameImmutable: true,
			Source:            aws.EndpointSourceCustom,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			os.Getenv("STORAGE_BUCKET_KEY_ID"),
			os.Getenv("STORAGE_BUCKET_KEY_SECRET"),
			"",
		)),
		config.WithRegion("auto"),
	)

	if err != nil {
		return err
	}

	client := s3.NewFromConfig(cfg)

	f, err := os.Open(fmt.Sprintf("./%s.jpg", id))

	if err != nil {
		return err
	}

	image, err := jpeg.Decode(f)
	if err != nil {
		return err
	}
	var imageBuffer bytes.Buffer
	if err := jpeg.Encode(&imageBuffer, image, nil); err != nil {
		return err
	}

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(os.Getenv("STORAGE_BUCKET_NAME")),
		Key:    aws.String(id + ".jpg"),
		Body:   bytes.NewReader(imageBuffer.Bytes()),
	})

	if err != nil {
		return err
	}

	f.Close()

	err = os.Remove(fmt.Sprintf("./%s.jpg", id))

	if err != nil {
		return err
	}

	return nil
}
