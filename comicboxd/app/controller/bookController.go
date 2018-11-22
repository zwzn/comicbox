package controller

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/abibby/comicbox/comicboxd/j"

	sq "github.com/Masterminds/squirrel"
	"github.com/abibby/comicbox/comicboxd/app"
	"github.com/abibby/comicbox/comicboxd/app/database"
	"github.com/abibby/comicbox/comicboxd/app/gql"
	"github.com/abibby/comicbox/comicboxd/app/model"
	"github.com/abibby/comicbox/comicboxd/errors"
	"github.com/chai2010/webp"
	"github.com/nfnt/resize"
)

type book struct{}

var Book = &book{}

func (b *book) Page(w http.ResponseWriter, r *http.Request) {
	c := app.Ctx(r)
	bookID := c.Var("id")
	pageNum := int(c.VarInt64("page"))
	book := model.BookUserBook{}
	err := gql.Query(r, `query getBook($id:ID!) {
		book(id: $id){
			file
			pages {
				file_number
				type
			}
		}
	}`, map[string]interface{}{"id": bookID}, &book)
	errors.Check(err)

	pages := book.Pages
	if pageNum < 0 || pageNum >= len(pages) {
		c.Response = errors.HTTP(404)
		return
	}
	if _, err := os.Stat(book.File); os.IsNotExist(err) {
		c.Response = fmt.Errorf("can't find file '%s'", book.File)
		return
	}

	imageFiles, err := ZippedImages(book.File)
	errors.Check(err)

	page := book.Pages[pageNum]

	rc, err := imageFiles[page.FileNumber].Open()
	if err != nil {
		c.Response = err
		return
	}
	defer rc.Close()
	imgBytes, err := ioutil.ReadAll(rc)
	errors.Check(err)

	c.Response = imgBytes
	mime := http.DetectContentType(imgBytes)
	var img image.Image
	if mime == "image/jpeg" {
		img, err = jpeg.Decode(bytes.NewReader(imgBytes))
	} else if mime == "image/png" {
		img, err = png.Decode(bytes.NewReader(imgBytes))
	} else {
		err = fmt.Errorf("Unsupported file type: %s", mime)
	}
	if err != nil {
		c.Response = err
		return
	}

	if height, ok := c.QGetInt64("height"); ok {
		img = resize.Resize(uint(height), 0, img, resize.Lanczos3)
	}
	quality, ok := c.QGetInt64("quality")
	if !ok {
		quality = 30
	}

	switch c.Var("ext") {
	case "jpg":
		err = jpeg.Encode(w, img, &jpeg.Options{
			Quality: int(quality),
		})
		errors.Check(err)
		// w.Header().Set("Content-Type", "image/jpeg")
	case "png":
		err = png.Encode(w, img)
		errors.Check(err)
		// w.Header().Set("Content-Type", "image/png")
	case "webp":
		err = webp.Encode(w, img, &webp.Options{
			Lossless: false,
			Quality:  float32(quality),
		})
		errors.Check(err)
		// w.Header().Set("Content-Type", "image/png")
	}
}

func ZippedImages(file string) ([]*zip.File, error) {
	// Open a zip archive for reading.
	reader, err := zip.OpenReader(file)
	if err != nil {
		return nil, fmt.Errorf("error opening zip: %v", err)
	}

	sort.Slice(reader.File, func(i, j int) bool {
		return strings.Compare(reader.File[i].Name, reader.File[j].Name) < 0
	})

	imageFiles := reader.File[:0]
	for _, x := range reader.File {
		lowerName := strings.ToLower(x.Name)
		if strings.HasSuffix(lowerName, ".jpg") ||
			strings.HasSuffix(lowerName, ".jpeg") ||
			strings.HasSuffix(lowerName, ".png") ||
			strings.HasSuffix(lowerName, ".bmp") ||
			strings.HasSuffix(lowerName, ".tiff") {
			imageFiles = append(imageFiles, x)
		}
	}
	return imageFiles, nil
}

func (b *book) Scan(w http.ResponseWriter, r *http.Request) {
	go scan(r)
}

func scan(r *http.Request) {
	defer (func() {
		if err := recover(); err != nil {
			j.Errorf("error scanning books: %s", err)
			Push.Error("error scanning books: %s", err)
		}
	})()

	Push.Message("Starting book scan")
	dbFiles := []string{}
	// addFiles := []string{}
	// removeFiles := []string{}
	sql, args, err := sq.Select("file").From("book").OrderBy("file").ToSql()
	errors.Check(err)

	err = database.Select(&dbFiles, sql, args...)
	errors.Check(err)

	realFiles := []string{}
	// err = filepath.Walk("/home/adam/manga", func(path string, info os.FileInfo, err error) error {
	err = filepath.Walk("/mnt/public/old_comics", func(path string, info os.FileInfo, err error) error {
		ext := filepath.Ext(path)
		if info.IsDir() || (ext != ".cbz" && ext != ".zip") {
			return nil
		}
		realFiles = append(realFiles, path)
		return nil
	})
	errors.Check(err)

	addFiles, removeFiles := DiffSlice(realFiles, dbFiles)

	Push.Message("Started Add")

	addFilesLen := len(addFiles)
	for i, path := range addFiles {
		if i%100 == 0 {
			Push.Message("Done %.0f%%", float64(i)/float64(addFilesLen)*100.0)
		}
		err = gql.Query(r, `mutation addBook($file: String) {
				book(book: { file: $file }) {
				  	id
				}
			}`, map[string]interface{}{
			"file": path,
		}, nil)
		if err != nil {
			j.Warningf("error adding file '%s': %v", path, err)
		}
	}

	for _, path := range removeFiles {
		fmt.Printf("%#v\n", path)
	}

	Push.Message("Finished book scan")
}

func DiffSlice(a, b []string) ([]string, []string) {
	aMap := map[string]struct{}{}

	onlyA := make([]string, 0, len(a))
	onlyB := make([]string, 0, len(b))

	for _, str := range a {
		aMap[str] = struct{}{}
	}

	for _, str := range b {
		if _, ok := aMap[str]; ok {
			delete(aMap, str)
		} else {
			onlyB = append(onlyB, str)
		}
	}

	for str := range aMap {
		onlyA = append(onlyA, str)
	}

	return onlyA, onlyB
}