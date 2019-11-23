package goodreads

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

const testApiKey = "test-api-key"

func TestNewClient(t *testing.T) {
	c := NewClient("api-key")
	assert.NotNil(t, c)
	assert.Equal(t, "api-key", c.ApiKey)
	assert.Equal(t, DefaultAPIClient, c.httpClient)
}

func TestClient_AuthorBooks(t *testing.T) {
	c, done := newTestClient(t, decodeTestCase{
		expectURL: fmt.Sprintf("/author/list/12345?key=%s&page=1", testApiKey),
		response:  `<response><author><id>AuthorID</id><name>AuthorName</name></author></response>`,
	})
	defer done()

	a, err := c.AuthorBooks("12345", 1)
	assert.Nil(t, err)
	assert.Equal(t, Author{
		ID:   "AuthorID",
		Name: "AuthorName",
	}, *a)
}

func TestClient_AuthorShow(t *testing.T) {
	c, done := newTestClient(t, decodeTestCase{
		expectURL: fmt.Sprintf("/author/show/12345?key=%s", testApiKey),
		response:  `<response><author><id>AuthorID</id><name>AuthorName</name></author></response>`,
	})
	defer done()

	a, err := c.AuthorShow("12345")
	assert.Nil(t, err)
	assert.Equal(t, Author{
		ID:   "AuthorID",
		Name: "AuthorName",
	}, *a)
}

func TestClient_BookReviewCounts(t *testing.T) {
	isbn := "9781400078776"
	c, done := newTestClient(t, decodeTestCase{
		expectURL: fmt.Sprintf("/book/review_counts.json?isbns=%s&key=%s", isbn, testApiKey),
		response: `{
			"books": [{
				"average_rating": "3.82",
				"id": 15,
				"isbn": "1400078776",
				"isbn13": "9781400078776",
				"ratings_count": 1,
				"reviews_count": 2,
				"text_reviews_count": 3,
				"work_ratings_count": 4,
				"work_reviews_count": 5,
				"work_text_reviews_count": 6
			}]
		}`,
	})
	defer done()

	counts, err := c.BookReviewCounts([]string{isbn})
	assert.Nil(t, err)
	assert.Equal(t, []ReviewCounts{
		{
			ID:                   15,
			ISBN:                 "1400078776",
			ISBN13:               "9781400078776",
			RatingsCount:         1,
			ReviewsCount:         2,
			TextReviewsCount:     3,
			WorkRatingsCount:     4,
			WorkReviewsCount:     5,
			WorkTextReviewsCount: 6,
			AverageRating:        "3.82",
		},
	}, counts)
}

func TestClient_ReviewList(t *testing.T) {
	c, done := newTestClient(t, decodeTestCase{
		expectURL: fmt.Sprintf("/review/list/user-id.xml?key=%s&order=d&page=1&per_page=200&search=search&shelf=read&sort=date_read&v=2", testApiKey),
		response: `<response>
			<reviews>
				<review><id>review1</id><rating>1</rating></review>
				<review><id>review2</id><rating>2</rating></review>
				<review><id>review3</id><rating>3</rating></review>
			</reviews>
		</response>`,
	})
	defer done()

	r, err := c.ReviewList("user-id", "read", "date_read", "search", "d", 1, 200)
	assert.Nil(t, err)
	assert.Equal(t, []Review{
		{ID: "review1", Rating: 1},
		{ID: "review2", Rating: 2},
		{ID: "review3", Rating: 3},
	}, r)
}

func TestClient_SearchBooks(t *testing.T) {
	c, done := newTestClient(t, decodeTestCase{
		expectURL: fmt.Sprintf("/search/index.xml?key=%s&q=hello&page=1&search[field]=all", testApiKey),
		// TODO: take a look at what the response type looks like and add fields
		response: `<response>
			<books>
				<user_book><id>book1</id><name>Book 1</name></user_book>
				<user_book><id>book2</id><name>Book 2</name></user_book>
				<user_book><id>book3</id><name>Book 3</name></user_book>
			</books>
		</response>`,
	})
	defer done()
	books, err := c.SearchBooks("hello", 0, AllFields)
	assert.Nil(t, err)
	assert.Equal(t, []Book{
		{ID: "book1", Title: "Book 1"},
		{ID: "book2", Title: "Book 2"},
		{ID: "book3", Title: "Book 3"},
	}, books)
}

func TestClient_ShelvesList(t *testing.T) {
	c, done := newTestClient(t, decodeTestCase{
		expectURL: fmt.Sprintf("/shelf/list.xml?key=%s&user_id=user-id", testApiKey),
		response: `<response>
			<shelves>
				<user_shelf><id>shelf1</id><name>Shelf 1</name></user_shelf>
				<user_shelf><id>shelf2</id><name>Shelf 2</name></user_shelf>
				<user_shelf><id>shelf3</id><name>Shelf 3</name></user_shelf>
			</shelves>
		</response>`,
	})
	defer done()

	s, err := c.ShelvesList("user-id")
	assert.Nil(t, err)
	assert.Equal(t, []UserShelf{
		{ID: "shelf1", Name: "Shelf 1"},
		{ID: "shelf2", Name: "Shelf 2"},
		{ID: "shelf3", Name: "Shelf 3"},
	}, s)
}

func TestClient_UserShow(t *testing.T) {
	c, done := newTestClient(t, decodeTestCase{
		expectURL: fmt.Sprintf("/user/show/user-id.xml?key=%s", testApiKey),
		response: `<response>
			<user>
				<id>user-id</id>
				<name>User Name</name>
			</user>
		</response>`,
	})
	defer done()

	u, err := c.UserShow("user-id")
	assert.Nil(t, err)
	assert.Equal(t, User{
		ID:   "user-id",
		Name: "User Name",
	}, *u)
}

type decodeTestCase struct {
	expectURL string
	response  string
}

func newTestClient(t *testing.T, tc decodeTestCase) (*Client, func()) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, tc.expectURL, r.URL.String())
		_, _ = w.Write([]byte(tc.response))
	}))

	return &Client{
		ApiKey: testApiKey,
		httpClient: &HTTPClient{
			Client:  http.DefaultClient,
			ApiRoot: s.URL,
			Verbose: true,
		},
	}, s.Close
}
