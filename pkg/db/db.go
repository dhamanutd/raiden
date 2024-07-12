package db

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/sev-2/raiden"
	"github.com/valyala/fasthttp"
)

type Query struct {
	Context      raiden.Context
	model        interface{}
	Columns      []string
	WhereAndList *[]string
	WhereOrList  *[]string
	OrderList    *[]string
	LimitValue   int
	OffsetValue  int
	Err          error
}

type ModelBase struct {
	raiden.ModelBase
}

type Where struct {
	column   string
	operator string
	value    any
}

func (q *Query) Error() error {
	return q.Err
}

func NewQuery(ctx raiden.Context) *Query {
	return &Query{
		Context: ctx,
	}
}

func (q *Query) Model(m interface{}) *Query {
	q.model = m
	return q
}

func GetTable(m interface{}) string {
	t := reflect.TypeOf(m)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	field, found := t.FieldByName("Metadata")
	if !found {
		fmt.Println("Field \"tableName\" is not found")
		return ""
	}

	tableName := field.Tag.Get("tableName")

	return tableName
}

func (m *ModelBase) Execute() (model *ModelBase) {
	return m
}

func (q Query) Get() ([]byte, error) {

	url := q.GetUrl()

	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Prefer"] = "return=representation"

	resp, _, err := PostgrestRequest(q.Context, fasthttp.MethodGet, url, nil, headers)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (q Query) Single() ([]byte, error) {
	url := q.Limit(1).GetUrl()

	headers := make(map[string]string)

	headers["Accept"] = "application/vnd.pgrst.object+json"

	res, _, err := PostgrestRequest(q.Context, "GET", url, nil, headers)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (q Query) GetUrl() string {
	urlQuery := buildQueryURI(q)
	baseUrl := getConfig().SupabasePublicUrl
	url := fmt.Sprintf("%s/rest/v1/%s", baseUrl, urlQuery)
	return url
}

func buildQueryURI(q Query) string {
	var output string

	output = fmt.Sprintf("%s?", GetTable(q.model))

	if len(q.Columns) > 0 {
		columns := strings.Join(q.Columns, ",")
		output += fmt.Sprintf("select=%s", columns)
	} else {
		output += "select=*"
	}

	if q.WhereAndList != nil && len(*q.WhereAndList) > 0 {
		eqList := strings.Join(*q.WhereAndList, "&")
		output += fmt.Sprintf("&%s", eqList)
	}

	if q.WhereOrList != nil && len(*q.WhereOrList) > 0 {
		list := strings.Join(*q.WhereOrList, ",")
		output += fmt.Sprintf("&or=(%s)", list)
	}

	if q.OrderList != nil && len(*q.OrderList) > 0 {
		orders := strings.Join(*q.OrderList, ",")
		output += fmt.Sprintf("&order=%s", orders)
	}

	if q.LimitValue > 0 {
		output += fmt.Sprintf("&limit=%v", q.LimitValue)
	}

	if q.OffsetValue > 0 {
		output += fmt.Sprintf("&offset=%v", q.OffsetValue)
	}

	return output
}
