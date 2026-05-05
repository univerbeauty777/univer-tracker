package woocommerce

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

// Order is a minimal projection of a WooCommerce order — only fields we care about.
type Order struct {
	ID               int64       `json:"id"`
	Status           string      `json:"status"`
	DateCreatedGMT   TimeRFC3339 `json:"date_created_gmt"`
	DatePaidGMT      TimeRFC3339 `json:"date_paid_gmt"`
	DateCompletedGMT TimeRFC3339 `json:"date_completed_gmt"`
	Total            string      `json:"total"`
	Billing          Address     `json:"billing"`
	Shipping         Address     `json:"shipping"`
	LineItems        []LineItem  `json:"line_items"`
	ShippingLines    []Shipping  `json:"shipping_lines"`
	MetaData         []Meta      `json:"meta_data"`
}

type Address struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	City      string `json:"city"`
	State     string `json:"state"`
	Postcode  string `json:"postcode"`
}

type LineItem struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Total    string `json:"total"`
}

type Shipping struct {
	ID          int64  `json:"id"`
	MethodTitle string `json:"method_title"`
	MethodID    string `json:"method_id"`
	MetaData    []Meta `json:"meta_data"`
}

type Meta struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

// TimeRFC3339 is a thin wrapper to handle WC's empty-string timestamps.
type TimeRFC3339 struct {
	time.Time
}

func (t *TimeRFC3339) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" || s == `""` {
		return nil
	}
	parsed, err := time.Parse(`"`+time.RFC3339+`"`, s)
	if err == nil {
		t.Time = parsed
		return nil
	}
	// Fallback: WC sends "2026-04-15T10:25:48" without timezone.
	parsed, err = time.Parse(`"2006-01-02T15:04:05"`, s)
	if err == nil {
		t.Time = parsed
	}
	return err
}

// ListOrdersParams configures the orders listing query.
type ListOrdersParams struct {
	Status   []string
	After    time.Time
	PerPage  int
	Page     int
	Modified time.Time
}

// ListOrders fetches a page of orders matching the filters.
func (c *Client) ListOrders(ctx context.Context, params ListOrdersParams) ([]Order, error) {
	q := url.Values{}
	if len(params.Status) > 0 {
		q.Set("status", joinComma(params.Status))
	}
	if !params.After.IsZero() {
		q.Set("after", params.After.UTC().Format(time.RFC3339))
	}
	if !params.Modified.IsZero() {
		q.Set("modified_after", params.Modified.UTC().Format(time.RFC3339))
	}
	per := params.PerPage
	if per <= 0 {
		per = 50
	}
	q.Set("per_page", strconv.Itoa(per))
	if params.Page > 0 {
		q.Set("page", strconv.Itoa(params.Page))
	}

	var orders []Order
	if err := c.do(ctx, "GET", "orders", q, nil, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

// GetOrder fetches a single order by id.
func (c *Client) GetOrder(ctx context.Context, id int64) (*Order, error) {
	var o Order
	if err := c.do(ctx, "GET", "orders/"+strconv.FormatInt(id, 10), nil, nil, &o); err != nil {
		return nil, err
	}
	return &o, nil
}

// UpdateOrderStatus moves an order to a new WooCommerce status.
func (c *Client) UpdateOrderStatus(ctx context.Context, id int64, status string) error {
	body := map[string]any{"status": status}
	return c.do(ctx, "PUT", "orders/"+strconv.FormatInt(id, 10), nil, body, nil)
}

// AddOrderNote attaches a note to an order. customerVisible determines whether
// the note is sent to the customer by email.
func (c *Client) AddOrderNote(ctx context.Context, orderID int64, note string, customerVisible bool) error {
	body := map[string]any{
		"note":          note,
		"customer_note": customerVisible,
	}
	return c.do(ctx, "POST", "orders/"+strconv.FormatInt(orderID, 10)+"/notes", nil, body, nil)
}

func joinComma(s []string) string {
	out := ""
	for i, v := range s {
		if i > 0 {
			out += ","
		}
		out += v
	}
	return out
}
