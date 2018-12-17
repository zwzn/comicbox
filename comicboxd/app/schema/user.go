package schema

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/zwzn/comicbox/comicboxd/app/controller"

	"github.com/Masterminds/squirrel"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jmoiron/sqlx"
	"github.com/zwzn/comicbox/comicboxd/app/database"
	"github.com/zwzn/comicbox/comicboxd/app/model"
)

func (q *query) Me(ctx context.Context) (*UserResolver, error) {
	c := q.Ctx(ctx)
	user := c.User
	if user == nil {
		return nil, fmt.Errorf("user not set")
	}
	return &UserResolver{u: *user}, nil
}

type userInput struct {
	Name     *string `db:"name"`
	Username *string `db:"username"`
	Password *string `db:"password"`
}

type userArgs struct {
	ID graphql.ID
}

func (q *query) User(ctx context.Context, args userArgs) (*UserResolver, error) {
	// c := q.Ctx(ctx)
	user := model.User{}
	qSQL, qArgs, err := squirrel.
		Select("*").
		From("user").
		Where(squirrel.Eq{"id": args.ID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	err = database.Get(&user, qSQL, qArgs...)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	return &UserResolver{u: user}, nil
}

type updateUserArgs struct {
	ID   graphql.ID
	Data userInput
}

func (q *query) UpdateUser(ctx context.Context, args updateUserArgs) (*UserResolver, error) {
	// c := q.Ctx(ctx)
	if args.Data.Password != nil {
		pass, err := controller.HashPassword(*args.Data.Password)
		if err != nil {
			return nil, err
		}
		args.Data.Password = &pass
	}
	m := toStruct(args.Data)
	if len(m) == 0 {
		return nil, nil
	}

	err := database.Tx(ctx, func(tx *sqlx.Tx) error {
		query := squirrel.Update("user").Where("id = ?", args.ID)

		query = update(m, query)

		_, err := query.RunWith(tx).Exec()
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return q.User(ctx, userArgs{ID: args.ID})
}

type UserResolver struct {
	u model.User
}

func (r *UserResolver) ID() graphql.ID {
	return graphql.ID(r.u.ID.String())
}
func (r *UserResolver) CreatedAt() graphql.Time {
	return graphql.Time{Time: r.u.CreatedAt}
}
func (r *UserResolver) UpdatedAt() graphql.Time {
	return graphql.Time{Time: r.u.UpdatedAt}
}
func (r *UserResolver) Name() string {
	return r.u.Name
}
func (r *UserResolver) Username() string {
	return r.u.Username
}
