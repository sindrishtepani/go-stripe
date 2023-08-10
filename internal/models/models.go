package models

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// DBModel is the type for database connection values
type DBModel struct {
	DB *sql.DB
}

// Models is the wrapper for all models
type Models struct {
	DB DBModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		DB: DBModel{
			DB: db,
		},
	}
}

type Widget struct {
	Id             int       `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Inventorylevel int       `json:"inventory_level"`
	Price          int       `json:"price"`
	Image          string    `json:"image"`
	IsRecurring    bool      `json:"is_recurring"`
	PlanId         string    `json:"plan_id"`
	CreatedAt      time.Time `json:"-"`
	UpdatedAt      time.Time `json:"-"`
}

type Order struct {
	Id            int         `json:"id"`
	WidgetId      int         `json:"widget_id"`
	CustomerId    int         `json:"customer_id"`
	TransactionId int         `json:"transaction_id"`
	StatusId      int         `json:"status_id"`
	Quantity      int         `json:"quantity"`
	Amount        int         `json:"amount"`
	CreatedAt     time.Time   `json:"-"`
	UpdatedAt     time.Time   `json:"-"`
	Widget        Widget      `json:"widget"`
	Transaction   Transaction `json:"transaction"`
	Customer      Customer    `json:"customer"`
}

type Status struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type TransactionStatus struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Transaction struct {
	Id                  int       `json:"id"`
	Amount              int       `json:"amount"`
	Currency            string    `json:"currency"`
	LastFour            string    `json:"last_four"`
	ExpiryMonth         int       `json:"expiry_month"`
	ExpiryYear          int       `json:"expiry_year"`
	PaymentIntent       string    `json:"payment_intent"`
	PaymentMethod       string    `json:"payment_method"`
	BankReturnCode      string    `json:"bank_return_code"`
	TransactionStatusId int       `json:"transaction_status_id"`
	CreatedAt           time.Time `json:"-"`
	UpdatedAt           time.Time `json:"-"`
}

type User struct {
	Id        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

type Customer struct {
	Id        int       `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (m *DBModel) GetWidget(id int) (Widget, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var widget Widget

	row := m.DB.QueryRowContext(ctx, `
	select 
	id, name, description, inventory_level, price, image, is_recurring, plan_id, created_at, updated_at 
	from widgets 
	where id = ?`, id)

	err := row.Scan(
		&widget.Id,
		&widget.Name,
		&widget.Description,
		&widget.Inventorylevel,
		&widget.Price,
		&widget.Image,
		&widget.IsRecurring,
		&widget.PlanId,
		&widget.CreatedAt,
		&widget.UpdatedAt,
	)

	if err != nil {
		return widget, err
	}

	return widget, nil
}

// InsertTransaction inserts a transaction and returns its id
func (m *DBModel) InsertTransaction(txn Transaction) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
	insert into transactions
		(amount, currency, last_four, bank_return_code, expiry_month, expiry_year, 
			payment_intent, payment_method, transaction_status_id, created_at, updated_at)
		values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := m.DB.ExecContext(ctx, stmt,
		txn.Amount,
		txn.Currency,
		txn.LastFour,
		txn.BankReturnCode,
		txn.ExpiryMonth,
		txn.ExpiryYear,
		txn.PaymentIntent,
		txn.PaymentMethod,
		txn.TransactionStatusId,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// InsertOrder inserts an order and returns its id
func (m *DBModel) InsertOrder(order Order) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
	insert into orders
		(widget_id, transaction_id, status_id, quantity, customer_id, amount, created_at, updated_at)
		values (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := m.DB.ExecContext(ctx, stmt,
		order.WidgetId,
		order.TransactionId,
		order.StatusId,
		order.Quantity,
		order.CustomerId,
		order.Amount,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// InsertCustomer inserts a customer and returns its id
func (m *DBModel) InsertCustomer(customer Customer) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `
	insert into customers
		(first_name, last_name, email, created_at, updated_at)
		values (?, ?, ?, ?, ?)`

	result, err := m.DB.ExecContext(ctx, stmt,
		customer.FirstName,
		customer.LastName,
		customer.Email,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

// GetUserByEmail gets a user by email address
func (m *DBModel) GetUserByEmail(email string) (User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	email = strings.ToLower(email)
	var u User

	row := m.DB.QueryRowContext(ctx, `
	select 
		id, first_name, last_name, email, password, created_at, updated_at
	from
		users
	where email = ?`, email)

	err := row.Scan(
		&u.Id,
		&u.FirstName,
		&u.LastName,
		&u.Email,
		&u.Password,
		&u.CreatedAt,
		&u.UpdatedAt,
	)
	if err != nil {
		return u, err
	}

	return u, nil
}

func (m *DBModel) Authenticate(email, password string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var id int
	var hashedPassword string

	row := m.DB.QueryRowContext(ctx, "select id, password from users where email = ?", email)
	err := row.Scan(
		&id,
		&hashedPassword,
	)

	if err != nil {
		return id, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return 0, errors.New("incorrect password")
	} else if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *DBModel) UpdatePasswordForUser(u User, hash string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `update users set password = ? where id = ?`

	_, err := m.DB.ExecContext(ctx, stmt, hash, u.Id)
	if err != nil {
		return err
	}

	return nil
}

func buildOrdersQuery(idWhereClause, isRecurring, addLimit bool) string {
	var recurringFlag int
	if isRecurring {
		recurringFlag = 1
	} else {
		recurringFlag = 0
	}

	var idQuery string
	var recurringQuery string
	var limitQuery string

	if idWhereClause {
		idQuery = " o.id = ? "
		recurringQuery = " w.is_recurring in (1, 0) "
	} else {
		idQuery = " o.id != 0 "
		recurringQuery = " w.is_recurring = " + fmt.Sprintf("%d", recurringFlag)
	}

	if addLimit {
		limitQuery = " limit ? offset ? "
	} else {
		limitQuery = ""
	}

	query := `
	select
		o.id, o.widget_id, o.transaction_id, o.customer_id, 
		o.status_id, o.quantity, o.amount, o.created_at,
		o.updated_at, w.id, w.name, t.id, t.amount, t.currency,
		t.last_four, t.expiry_month, t.expiry_year, t.payment_intent,
		t.bank_return_code, c.id, c.first_name, c.last_name, c.email
		
	from
		orders o
		left join widgets w on (o.widget_id = w.id)
		left join transactions t on (o.transaction_id = t.id)
		left join customers c on (o.customer_id = c.id)
	where ` +
		recurringQuery + `
		and ` + idQuery + `
	order by
		o.created_at desc` +
		limitQuery

	return query
}

func (m DBModel) GetAllOrdersPaginated(pageSize, page int) ([]*Order, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	offset := (page - 1) * pageSize

	var orders []*Order

	query := buildOrdersQuery(false, false, true)

	rows, err := m.DB.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var o Order

		err = rows.Scan(
			&o.Id,
			&o.WidgetId,
			&o.TransactionId,
			&o.CustomerId,
			&o.StatusId,
			&o.Quantity,
			&o.Amount,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.Widget.Id,
			&o.Widget.Name,
			&o.Transaction.Id,
			&o.Transaction.Amount,
			&o.Transaction.Currency,
			&o.Transaction.LastFour,
			&o.Transaction.ExpiryMonth,
			&o.Transaction.ExpiryYear,
			&o.Transaction.PaymentIntent,
			&o.Transaction.BankReturnCode,
			&o.Customer.Id,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
		)
		if err != nil {
			return nil, 0, 0, err
		}
		orders = append(orders, &o)
	}

	query = `select count(o.id)
			from orders o
			left join widgets w on (o.widget_id = w.id)
			where
			w.is_recurring = 0`

	var totalRecords int
	countRow := m.DB.QueryRowContext(ctx, query)
	err = countRow.Scan(&totalRecords)
	if err != nil {
		return nil, 0, 0, err
	}

	lastPage := totalRecords / pageSize

	return orders, lastPage, totalRecords, nil
}

func (m DBModel) GetAllSubscriptionsPaginated(pageSize, page int) ([]*Order, int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	offset := (page - 1) * pageSize

	var orders []*Order

	query := buildOrdersQuery(false, true, true)

	rows, err := m.DB.QueryContext(ctx, query, pageSize, offset)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var o Order

		err = rows.Scan(
			&o.Id,
			&o.WidgetId,
			&o.TransactionId,
			&o.CustomerId,
			&o.StatusId,
			&o.Quantity,
			&o.Amount,
			&o.CreatedAt,
			&o.UpdatedAt,
			&o.Widget.Id,
			&o.Widget.Name,
			&o.Transaction.Id,
			&o.Transaction.Amount,
			&o.Transaction.Currency,
			&o.Transaction.LastFour,
			&o.Transaction.ExpiryMonth,
			&o.Transaction.ExpiryYear,
			&o.Transaction.PaymentIntent,
			&o.Transaction.BankReturnCode,
			&o.Customer.Id,
			&o.Customer.FirstName,
			&o.Customer.LastName,
			&o.Customer.Email,
		)
		if err != nil {
			return nil, 0, 0, err
		}
		orders = append(orders, &o)
	}

	query = `select count(o.id)
			from orders o
			left join widgets w on (o.widget_id = w.id)
			where
			w.is_recurring = 1`

	var totalRecords int
	countRow := m.DB.QueryRowContext(ctx, query)
	err = countRow.Scan(&totalRecords)
	if err != nil {
		return nil, 0, 0, err
	}

	lastPage := totalRecords / pageSize

	return orders, lastPage, totalRecords, nil
}

func (m DBModel) GetOrderById(id int) (Order, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var o Order

	query := buildOrdersQuery(true, false, false)

	row := m.DB.QueryRowContext(ctx, query, id)

	err := row.Scan(
		&o.Id,
		&o.WidgetId,
		&o.TransactionId,
		&o.CustomerId,
		&o.StatusId,
		&o.Quantity,
		&o.Amount,
		&o.CreatedAt,
		&o.UpdatedAt,
		&o.Widget.Id,
		&o.Widget.Name,
		&o.Transaction.Id,
		&o.Transaction.Amount,
		&o.Transaction.Currency,
		&o.Transaction.LastFour,
		&o.Transaction.ExpiryMonth,
		&o.Transaction.ExpiryYear,
		&o.Transaction.PaymentIntent,
		&o.Transaction.BankReturnCode,
		&o.Customer.Id,
		&o.Customer.FirstName,
		&o.Customer.LastName,
		&o.Customer.Email,
	)
	if err != nil {
		return o, err
	}

	return o, nil

}

func (m *DBModel) UpdateOrderStatus(id, statusId int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	stmt := `update orders set status_id = ? where id = ?`

	_, err := m.DB.ExecContext(ctx, stmt, statusId, id)
	if err != nil {
		return err
	}

	return nil
}
