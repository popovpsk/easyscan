package benchmarks

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgproto3/v2"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"

	"github.com/popovpsk/easyscan"
)

type User struct {
	ID               int           `db:"id"`
	Name             string        `db:"name"`
	Age              int           `db:"age"`
	Email            string        `db:"email"`
	Address          string        `db:"address"`
	PhoneNumber      string        `db:"phone_number"`
	Gender           string        `db:"gender"`
	Nationality      string        `db:"nationality"`
	Education        string        `db:"education"`
	Profession       string        `db:"profession"`
	Interests        string        `db:"interests"`
	Income           float64       `db:"income"`
	CreditScore      int           `db:"credit_score"`
	CreatedAt        time.Time     `db:"created_at"`
	UpdatedAt        time.Time     `db:"updated_at"`
	IsActive         bool          `db:"is_active"`
	IsVerified       bool          `db:"is_verified"`
	LastLogin        *time.Time    `db:"last_login"`
	Latitude         *float64      `db:"latitude"`
	Longitude        *float64      `db:"longitude"`
	Country          string        `db:"country"`
	City             string        `db:"city"`
	PostalCode       string        `db:"postal_code"`
	ProfilePic       string        `db:"profile_pic"`
	Website          string        `db:"website"`
	SocialMedia      string        `db:"social_media"`
	SubscriptionPlan string        `db:"subscription_plan"`
	FollowersCount   sql.NullInt64 `db:"followers_count"`
	FollowingCount   sql.NullInt64 `db:"following_count"`
}

type Product struct {
	ID          int             `db:"id"`
	Name        string          `db:"name"`
	Description string          `db:"description"`
	Price       float64         `db:"price"`
	Category    string          `db:"category"`
	Brand       sql.NullString  `db:"brand"`
	Rating      sql.NullFloat64 `db:"rating"`
	CreatedAt   time.Time       `db:"created_at"`
	UpdatedAt   *time.Time      `db:"updated_at"`
}

type queryType int

const (
	usersQuery queryType = iota
	productsQuery
)

type dbQueryStub struct {
	rows      int
	queryType queryType
}

func (q dbQueryStub) Query(_ context.Context, _ string, _ ...interface{}) (pgx.Rows, error) {
	return &rowsImpl{count: q.rows, queryType: q.queryType}, nil
}

type rowsImpl struct {
	current,
	count int
	queryType queryType
}

func (*rowsImpl) Close() {
}

func (*rowsImpl) Err() error {
	return nil
}

func (r *rowsImpl) CommandTag() pgconn.CommandTag {
	panic("implement me")
}

var userFd = []pgproto3.FieldDescription{
	{Name: []byte("id"), DataTypeOID: pgtype.Int8OID},
	{Name: []byte("name"), DataTypeOID: pgtype.TextOID},
	{Name: []byte("age"), DataTypeOID: pgtype.Int4OID},
	{Name: []byte("email"), DataTypeOID: pgtype.TextOID},
	{Name: []byte("address"), DataTypeOID: pgtype.TextOID},
	{Name: []byte("phone_number"), DataTypeOID: pgtype.TextOID},
	{Name: []byte("gender"), DataTypeOID: pgtype.TextOID},
	{Name: []byte("nationality"), DataTypeOID: pgtype.TextOID},
	{Name: []byte("education"), DataTypeOID: pgtype.TextOID},
	{Name: []byte("profession"), DataTypeOID: pgtype.TextOID},
	{Name: []byte("interests"), DataTypeOID: pgtype.TextOID},
	{Name: []byte("income"), DataTypeOID: pgtype.Float8OID},
	{Name: []byte("credit_score"), DataTypeOID: pgtype.Int4OID},
	{Name: []byte("created_at"), DataTypeOID: pgtype.TimestampOID},
	{Name: []byte("updated_at"), DataTypeOID: pgtype.TimestampOID},
	{Name: []byte("is_active"), DataTypeOID: pgtype.BoolOID},
	{Name: []byte("is_verified"), DataTypeOID: pgtype.BoolOID},
	{Name: []byte("last_login"), DataTypeOID: pgtype.TimestamptzOID},
	{Name: []byte("latitude"), DataTypeOID: pgtype.Float8OID},
	{Name: []byte("longitude"), DataTypeOID: pgtype.Float8OID},
	{Name: []byte("country"), DataTypeOID: pgtype.TextOID},
	{Name: []byte("city"), DataTypeOID: pgtype.TextOID},
	{Name: []byte("postal_code"), DataTypeOID: pgtype.TextOID},
	{Name: []byte("profile_pic"), DataTypeOID: pgtype.TextOID},
	{Name: []byte("website"), DataTypeOID: pgtype.TextOID},
	{Name: []byte("social_media"), DataTypeOID: pgtype.TextOID},
	{Name: []byte("subscription_plan"), DataTypeOID: pgtype.TextOID},
	{Name: []byte("followers_count"), DataTypeOID: pgtype.Int8OID},
	{Name: []byte("following_count"), DataTypeOID: pgtype.Int8OID},
}

var productFd = []pgproto3.FieldDescription{
	{Name: []byte("id"), DataTypeOID: pgtype.Int8OID},
	{Name: []byte("name"), DataTypeOID: pgtype.TextOID},
	{Name: []byte("description"), DataTypeOID: pgtype.TextOID},
	{Name: []byte("price"), DataTypeOID: pgtype.Float8OID},
	{Name: []byte("category"), DataTypeOID: pgtype.TextOID},
	{Name: []byte("brand"), DataTypeOID: pgtype.TextOID},
	{Name: []byte("rating"), DataTypeOID: pgtype.Float8OID},
	{Name: []byte("created_at"), DataTypeOID: pgtype.TimestampOID},
	{Name: []byte("updated_at"), DataTypeOID: pgtype.TimestampOID},
}

func (r *rowsImpl) FieldDescriptions() []pgproto3.FieldDescription {
	switch r.queryType {
	case usersQuery:
		return userFd
	case productsQuery:
		return productFd
	default:
		panic("unreachable")
	}
}

func (r *rowsImpl) Next() bool {
	result := r.current < r.count
	r.current++
	return result
}

func (r *rowsImpl) Scan(dest ...interface{}) error {
	*(dest[0]).(*int) = r.current
	return nil
}

func (*rowsImpl) Values() ([]interface{}, error) {
	panic("implement me")
}

func (*rowsImpl) RawValues() [][]byte {
	panic("implement me")
}

func BenchmarkGet_PurePgx(b *testing.B) {
	ctx := context.Background()
	const rows = 1

	for i := 0; i < b.N; i++ {
		func() {
			user := User{}
			r, err := (&dbQueryStub{queryType: usersQuery, rows: rows}).Query(ctx, "select * from users where name=$1", "test")
			if err != nil {
				panic(err)
			}
			defer r.Close()

			err = r.Scan(&user.ID, &user.Name, &user.Age, &user.Email, &user.Address, &user.PhoneNumber, &user.Gender, &user.Nationality, &user.Education, &user.Profession, &user.Interests, &user.Income, &user.CreditScore, &user.CreatedAt, &user.UpdatedAt, &user.IsActive, &user.IsVerified, &user.LastLogin, &user.Latitude, &user.Longitude, &user.Country, &user.City, &user.PostalCode, &user.ProfilePic, &user.Website, &user.SocialMedia, &user.SubscriptionPlan, &user.FollowersCount, &user.FollowingCount)
			if err != nil {
				panic(err)
			}
		}()
	}
}

func BenchmarkGet_Easyscan(b *testing.B) {
	ctx := context.Background()
	const rows = 1

	for i := 0; i < b.N; i++ {
		user := User{}
		err := easyscan.Get(ctx, &dbQueryStub{queryType: usersQuery, rows: rows}, &user, "select * from users where name=$1", "test")
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkGet_GeorgysavaScany(b *testing.B) {
	ctx := context.Background()
	const rows = 1

	for i := 0; i < b.N; i++ {
		user := User{}
		err := pgxscan.Get(ctx, &dbQueryStub{queryType: usersQuery, rows: rows}, &user, "select * from users where name=$1", "test")
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkSelect20RowsToValues_PurePgx(b *testing.B) {
	ctx := context.Background()
	const rows = 20

	for i := 0; i < b.N; i++ {
		func() {
			users := make([]User, 0)
			r, err := (&dbQueryStub{queryType: usersQuery, rows: rows}).Query(ctx, "select * from users")
			if err != nil {
				panic(err)
			}
			defer r.Close()

			for r.Next() {
				var user User
				err = r.Scan(&user.ID, &user.Name, &user.Age, &user.Email, &user.Address, &user.PhoneNumber, &user.Gender, &user.Nationality, &user.Education, &user.Profession, &user.Interests, &user.Income, &user.CreditScore, &user.CreatedAt, &user.UpdatedAt, &user.IsActive, &user.IsVerified, &user.LastLogin, &user.Latitude, &user.Longitude, &user.Country, &user.City, &user.PostalCode, &user.ProfilePic, &user.Website, &user.SocialMedia, &user.SubscriptionPlan, &user.FollowersCount, &user.FollowingCount)
				if err != nil {
					panic(err)
				}
				users = append(users, user)
			}
		}()
	}
}

func BenchmarkSelect20RowsToValues_Easyscan(b *testing.B) {
	ctx := context.Background()
	const rows = 20

	for i := 0; i < b.N; i++ {
		user := make([]User, 0, 20)
		err := easyscan.Select(ctx, &dbQueryStub{queryType: usersQuery, rows: rows}, &user, "select * from users")
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkSelect20RowsToValues_GeorgysavaScany(b *testing.B) {
	ctx := context.Background()
	const rows = 20

	for i := 0; i < b.N; i++ {
		user := make([]User, 0)
		err := pgxscan.Select(ctx, &dbQueryStub{queryType: usersQuery, rows: rows}, &user, "select * from users")
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkSelect20RowsToPointers_PurePgx(b *testing.B) {
	ctx := context.Background()
	const rows = 20

	for i := 0; i < b.N; i++ {
		func() {
			users := make([]*User, 0)
			r, err := (&dbQueryStub{queryType: usersQuery, rows: rows}).Query(ctx, "select * from users")
			if err != nil {
				panic(err)
			}
			defer r.Close()

			for r.Next() {
				user := &User{}
				err = r.Scan(&user.ID, &user.Name, &user.Age, &user.Email, &user.Address, &user.PhoneNumber, &user.Gender, &user.Nationality, &user.Education, &user.Profession, &user.Interests, &user.Income, &user.CreditScore, &user.CreatedAt, &user.UpdatedAt, &user.IsActive, &user.IsVerified, &user.LastLogin, &user.Latitude, &user.Longitude, &user.Country, &user.City, &user.PostalCode, &user.ProfilePic, &user.Website, &user.SocialMedia, &user.SubscriptionPlan, &user.FollowersCount, &user.FollowingCount)
				if err != nil {
					panic(err)
				}
				users = append(users, user)
			}
		}()
	}
}

func BenchmarkSelect20RowsToPointers_Easyscan(b *testing.B) {
	ctx := context.Background()
	const rows = 20

	for i := 0; i < b.N; i++ {
		user := make([]*User, 0)
		err := easyscan.Select(ctx, &dbQueryStub{queryType: usersQuery, rows: rows}, &user, "select * from users")
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkSelect20RowsToPointers_GeorgysavaScany(b *testing.B) {
	ctx := context.Background()
	const rows = 20

	for i := 0; i < b.N; i++ {
		user := make([]*User, 0)
		err := pgxscan.Select(ctx, &dbQueryStub{queryType: usersQuery, rows: rows}, &user, "select * from users")
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkSelect1000RowsWithoutPrealloc_PurePgx(b *testing.B) {
	ctx := context.Background()
	const rows = 1000

	for i := 0; i < b.N; i++ {
		func() {
			users := make([]User, 0)
			r, err := (&dbQueryStub{queryType: usersQuery, rows: rows}).Query(ctx, "select * from users")
			if err != nil {
				panic(err)
			}
			defer r.Close()

			user := User{}
			args := []interface{}{&user.ID, &user.Name, &user.Age, &user.Email, &user.Address, &user.PhoneNumber, &user.Gender, &user.Nationality, &user.Education, &user.Profession, &user.Interests, &user.Income, &user.CreditScore, &user.CreatedAt, &user.UpdatedAt, &user.IsActive, &user.IsVerified, &user.LastLogin, &user.Latitude, &user.Longitude, &user.Country, &user.City, &user.PostalCode, &user.ProfilePic, &user.Website, &user.SocialMedia, &user.SubscriptionPlan, &user.FollowersCount, &user.FollowingCount}
			for r.Next() {
				err = r.Scan(args...)
				if err != nil {
					panic(err)
				}
				users = append(users, user)
			}
		}()
	}
}

func BenchmarkSelect1000RowsWithout_PreallocEasyscan(b *testing.B) {
	ctx := context.Background()
	const rows = 1000

	for i := 0; i < b.N; i++ {
		user := make([]User, 0)
		err := easyscan.Select(ctx, &dbQueryStub{queryType: usersQuery, rows: rows}, &user, "select * from users")
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkSelect1000RowsWithoutPrealloc_GeorgysavaScany(b *testing.B) {
	ctx := context.Background()
	const rows = 1000

	for i := 0; i < b.N; i++ {
		user := make([]User, 0)
		err := pgxscan.Select(ctx, &dbQueryStub{queryType: usersQuery, rows: rows}, &user, "select * from users")
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkSelect1000RowsWithPrealloc_PurePgx(b *testing.B) {
	ctx := context.Background()
	const rows = 1000

	for i := 0; i < b.N; i++ {
		func() {
			users := make([]User, 0, rows)
			r, err := (&dbQueryStub{queryType: usersQuery, rows: rows}).Query(ctx, "select * from users")
			if err != nil {
				panic(err)
			}
			defer r.Close()

			user := User{}
			args := []interface{}{&user.ID, &user.Name, &user.Age, &user.Email, &user.Address, &user.PhoneNumber, &user.Gender, &user.Nationality, &user.Education, &user.Profession, &user.Interests, &user.Income, &user.CreditScore, &user.CreatedAt, &user.UpdatedAt, &user.IsActive, &user.IsVerified, &user.LastLogin, &user.Latitude, &user.Longitude, &user.Country, &user.City, &user.PostalCode, &user.ProfilePic, &user.Website, &user.SocialMedia, &user.SubscriptionPlan, &user.FollowersCount, &user.FollowingCount}
			for r.Next() {
				err = r.Scan(args...)
				if err != nil {
					panic(err)
				}
				users = append(users, user)
			}
		}()
	}
}

func BenchmarkSelect1000RowsWithPrealloc_Easyscan(b *testing.B) {
	ctx := context.Background()
	const rows = 1000

	for i := 0; i < b.N; i++ {
		users := make([]User, 0, rows)
		err := easyscan.Select(ctx, &dbQueryStub{queryType: usersQuery, rows: rows}, &users, "select * from users")
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkSelect1000RowsWithPrealloc_GeorgysavaScany(b *testing.B) {
	ctx := context.Background()
	const rows = 1000

	for i := 0; i < b.N; i++ {
		user := make([]User, 0, rows)
		err := pgxscan.Select(ctx, &dbQueryStub{queryType: usersQuery, rows: rows}, &user, "select * from users")
		if err != nil {
			panic(err)
		}
	}
}
