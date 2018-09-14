package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"bitbucket.org/zwzn/comicbox/comicboxd/app/database"
	"bitbucket.org/zwzn/comicbox/comicboxd/app/gql"
	"bitbucket.org/zwzn/comicbox/comicboxd/app/model"
	"github.com/google/uuid"
	"github.com/graphql-go/graphql"
)

type BookQuery struct {
	PageInfo gql.Page              `json:"page_info"`
	Results  []*model.BookUserBook `json:"results"`
}

var BookType = graphql.NewObject(graphql.ObjectConfig{
	Name: "book",
	Fields: graphql.Fields{
		"id": &graphql.Field{
			Type:        graphql.ID,
			Description: "a unique id for the books",
			Resolve:     gql.ResolveVal("ID"),
		},
		"created_at": &graphql.Field{
			Type:        graphql.DateTime,
			Description: "the date a book was created",
			Resolve:     gql.ResolveVal("CreatedAt"),
		},
		"updated_at": &graphql.Field{
			Type:    graphql.DateTime,
			Resolve: gql.ResolveVal("UpdatedAt"),
		},
		"series": &graphql.Field{
			Type:    graphql.String,
			Resolve: gql.ResolveVal("Series"),
		},
		"summary": &graphql.Field{
			Type:    graphql.String,
			Resolve: gql.ResolveVal("Series"),
		},
		"story_arc": &graphql.Field{
			Type:    graphql.String,
			Resolve: gql.ResolveVal("StoryArc"),
		},
		"authors": &graphql.Field{
			Type: graphql.NewList(graphql.String),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				book := p.Source.(*model.BookUserBook)
				authors := []string{}

				if len(book.AuthorsJSON) == 0 {
					return authors, nil
				}

				err := json.Unmarshal(book.AuthorsJSON, authors)
				if err != nil {
					return nil, err
				}
				return authors, nil
			},
		},
		"web": &graphql.Field{
			Type:    graphql.String,
			Resolve: gql.ResolveVal("Web"),
		},
		"genres": &graphql.Field{
			Type: graphql.NewList(graphql.String),
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				book := p.Source.(*model.BookUserBook)
				genres := []string{}

				if len(book.GenresJSON) == 0 {
					return genres, nil
				}

				err := json.Unmarshal(book.GenresJSON, genres)
				if err != nil {
					return nil, err
				}
				return genres, nil
			},
		},
		"alternate_series": &graphql.Field{
			Type:    graphql.String,
			Resolve: gql.ResolveVal("AlternateSeries"),
		},
		"reading_direction": &graphql.Field{
			Type:    graphql.String,
			Resolve: gql.ResolveVal("ReadingDirection"),
		},
		"type": &graphql.Field{
			Type:    graphql.String,
			Resolve: gql.ResolveVal("MediaType"),
		},
		"file": &graphql.Field{
			Type:    graphql.NewNonNull(graphql.String),
			Resolve: gql.ResolveVal("File"),
		},
		"title": &graphql.Field{
			Type:    graphql.String,
			Resolve: gql.ResolveVal("Title"),
		},
		"volume": &graphql.Field{
			Type:    graphql.Int,
			Resolve: gql.ResolveVal("Volume"),
		},
		"community_rating": &graphql.Field{
			Type:    graphql.Float,
			Resolve: gql.ResolveVal("CommunityRating"),
		},
		"chapter": &graphql.Field{
			Type:    graphql.Float,
			Resolve: gql.ResolveVal("Chapter"),
		},
		"date_released": &graphql.Field{
			Type:    graphql.DateTime,
			Resolve: gql.ResolveVal("DateReleased"),
		},
		"current_page": &graphql.Field{
			Type:    graphql.Int,
			Resolve: gql.ResolveVal("CurrentPage"),
		},
		"last_page_read": &graphql.Field{
			Type:    graphql.Int,
			Resolve: gql.ResolveVal("LastPageRead"),
		},
		"rating": &graphql.Field{
			Type:    graphql.Float,
			Resolve: gql.ResolveVal("Rating"),
		},
		"read": &graphql.Field{
			Type:    graphql.Boolean,
			Resolve: gql.ResolveVal("Read"),
		},
		"pages": &graphql.Field{
			Type: graphql.NewList(PageType),
			Args: graphql.FieldConfigArgument{
				"type": &graphql.ArgumentConfig{
					Type: PageTypeEnum,
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				t, typeOk := p.Args["type"].(string)
				book := p.Source.(*model.BookUserBook)
				allPages := []*model.Page{}
				pages := []*model.Page{}

				if len(book.PagesJSON) == 0 {
					return pages, nil
				}

				err := json.Unmarshal(book.PagesJSON, &allPages)
				if err != nil {
					return nil, err
				}

				for _, page := range allPages {
					if page.Type == t || !typeOk {
						pages = append(pages, page)
					}
				}

				return pages, nil
			},
		},
	},
})

var PageTypeEnum = graphql.NewEnum(graphql.EnumConfig{
	Name: "page_type",
	Values: graphql.EnumValueConfigMap{
		"COVER": &graphql.EnumValueConfig{
			Value: "FrontCover",
		},
		"STORY": &graphql.EnumValueConfig{
			Value: "Story",
		},
		"DELETED": &graphql.EnumValueConfig{
			Value: "Deleted",
		},
	},
})

var PageType = graphql.NewObject(graphql.ObjectConfig{
	Name: "page",
	Fields: graphql.Fields{
		"file_number": &graphql.Field{
			Type: graphql.Int,
		},
		"type": &graphql.Field{
			Type: PageTypeEnum,
		},
		"url": &graphql.Field{
			Type: graphql.String,
		},
	},
})

var BookQueryType = graphql.NewObject(graphql.ObjectConfig{
	Name: "book_query",
	Fields: graphql.Fields{
		"page_info": &graphql.Field{
			Type: gql.PageInfoType,
		},
		"results": &graphql.Field{
			Type: graphql.NewList(BookType),
		},
	},
})
var BookQueries = graphql.Fields{
	"book": &graphql.Field{
		Type: BookType,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.ID),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			c := gql.Ctx(p)
			book := &model.BookUserBook{}
			err := database.DB.Get(book, `SELECT * FROM "book_user_book" where  user_id=? and id=? limit 1;`, c.User.ID, p.Args["id"])
			if err == sql.ErrNoRows {
				return nil, nil
			} else if err != nil {
				return nil, err
			}
			return book, nil
		},
	},
	"books": &graphql.Field{
		Args: graphql.FieldConfigArgument{
			"take": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(graphql.Int),
			},
			"skip": &graphql.ArgumentConfig{
				Type: graphql.Int,
			},
			"series": &graphql.ArgumentConfig{
				Type: graphql.String,
			},
		},
		Type: BookQueryType,
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			c := gql.Ctx(p)

			skip, _ := p.Args["skip"].(int)
			take, _ := p.Args["take"].(int)
			if 0 > take || take > 100 {
				return nil, fmt.Errorf("you must take between 0 and 100 items")
			}
			data := []interface{}{c.User.ID}
			wheres, whereData := gql.ToSQL(model.BookUserBook{}, p.Args)
			data = append(data, whereData...)

			var count int
			err := database.DB.Get(&count, fmt.Sprintf(`SELECT count(*) FROM "book_user_book" where user_id=? %s;`, wheres), data...)
			if err == sql.ErrNoRows {
				count = 0
			} else if err != nil {
				return nil, err
			}

			books := []*model.BookUserBook{}

			err = database.DB.Select(&books, fmt.Sprintf(`SELECT * FROM "book_user_book" where user_id=? %s limit ? offset ?;`, wheres), append(data, take, skip)...)
			if err == sql.ErrNoRows {
				return nil, nil
			} else if err != nil {
				return nil, err
			}

			bq := BookQuery{
				PageInfo: gql.Page{
					Skip:  skip,
					Take:  take,
					Total: count,
				},
				Results: books,
			}
			return bq, nil
		},
	},
}

var BookInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "book_input",
	Fields: graphql.InputObjectConfigFieldMap{
		"series": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"summary": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"story_arc": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"authors": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.String),
		},
		"web": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"genres": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(graphql.String),
		},
		"alternate_series": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"reading_direction": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"type": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"file": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"title": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"volume": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
		"community_rating": &graphql.InputObjectFieldConfig{
			Type: graphql.Float,
		},
		"chapter": &graphql.InputObjectFieldConfig{
			Type: graphql.Float,
		},
		"date_released": &graphql.InputObjectFieldConfig{
			Type: graphql.DateTime,
		},
		"current_page": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
		"last_page_read": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
		"rating": &graphql.InputObjectFieldConfig{
			Type: graphql.Float,
		},
		"read": &graphql.InputObjectFieldConfig{
			Type: graphql.Boolean,
		},
		"pages": &graphql.InputObjectFieldConfig{
			Type: graphql.NewList(PageInput),
		},
	},
})

var PageInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "page_input",
	Fields: graphql.InputObjectConfigFieldMap{
		"url": &graphql.InputObjectFieldConfig{
			Type: graphql.String,
		},
		"number": &graphql.InputObjectFieldConfig{
			Type: graphql.Int,
		},
	},
})

var BookMutations = graphql.Fields{
	"book": &graphql.Field{
		Type: BookType,
		Args: graphql.FieldConfigArgument{
			"id": &graphql.ArgumentConfig{
				Type: graphql.ID,
			},
			"book": &graphql.ArgumentConfig{
				Type: graphql.NewNonNull(BookInput),
			},
		},
		Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			var err error
			c := gql.Ctx(p)
			book := &model.BookUserBook{}
			fmt.Printf("%#v\n", p.Args["book"])
			id, old := p.Args["id"]
			if old {
				// run an sql update
				return nil, fmt.Errorf("make the update adam")
			} else {
				book.ID, err = uuid.NewRandom()
				if err != nil {
					return nil, err
				}
				id = book.ID
				//run sql insert with new id
				return nil, fmt.Errorf("make the insert adam")
			}

			err = database.DB.Get(book, `SELECT * FROM "book_user_book" where  user_id=? and id=? limit 1;`, c.User.ID, id)
			if err == sql.ErrNoRows {
				return nil, nil
			} else if err != nil {
				return nil, err
			}

			return book, nil
		},
	},
}
