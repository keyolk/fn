package models

import (
	"errors"
	"net/http"
	"time"

	"github.com/fnproject/fn/api/id"
	"github.com/go-openapi/strfmt"
)

//go:generate jsonenums -type=TriggerType
type TriggerType int

const (
	Unknown TriggerType = iota
	HTTP
)

type Trigger struct {
	ID          string          `json:"id" db:"id"`
	Name        string          `json:"name" db:"name"`
	AppID       string          `json:"app_id" db:"app_id"`
	FnID        string          `json:"fn_id" db:"fn_id"`
	CreatedAt   strfmt.DateTime `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt   strfmt.DateTime `json:"updated_at,omitempty" db:"updated_at"`
	Type        TriggerType     `json:"type" db:"type"`
	Source      string          `json:"source" db:"source"`
	Annotations Annotations     `json:"annotations,omitempty" db:"annotations"`
}

func (t *Trigger) SetDefaults() {
	if time.Time(t.CreatedAt).IsZero() {
		t.CreatedAt = strfmt.DateTime(time.Now())
	}
	if time.Time(t.UpdatedAt).IsZero() {
		t.UpdatedAt = strfmt.DateTime(time.Now())
	}
	if t.ID == "" {
		t.ID = id.New().String()
	}
}

func (t1 *Trigger) Equals(t2 *Trigger) bool {
	eq := true
	eq = eq && t1.ID == t2.ID
	eq = eq && t1.Name == t2.Name
	eq = eq && t1.AppID == t2.AppID
	eq = eq && t1.FnID == t2.FnID

	eq = eq && t1.Type == t2.Type
	eq = eq && t1.Source == t2.Source
	eq = eq && t1.Annotations.Equals(t2.Annotations)

	// NOTE: datastore tests are not very fun to write with timestamp checks,
	// and these are not values the user may set so we kind of don't care.
	//eq = eq && time.Time(t1.CreatedAt).Equal(time.Time(t2.CreatedAt))
	//eq = eq && time.Time(t1.UpdatedAt).Equal(time.Time(t2.UpdatedAt))
	return eq
}

var (
	ErrTriggerMissingName = err{
		code:  http.StatusBadRequest,
		error: errors.New("Missing Trigger Name")}
	ErrTriggerMissingAppID = err{
		code:  http.StatusBadRequest,
		error: errors.New("Missing Trigger AppID")}
	ErrTriggerMissingFnID = err{
		code:  http.StatusBadRequest,
		error: errors.New("Missing Trigger FnID")}
	ErrTriggerTypeUnknown = err{
		code:  http.StatusBadRequest,
		error: errors.New("Trigger Type Unknown")}
	ErrTriggerMissingSource = err{
		code:  http.StatusBadRequest,
		error: errors.New("Missing Trigger Source")}
	ErrTriggerNotFound = err{
		code:  http.StatusNotFound,
		error: errors.New("Trigger not found")}
	ErrTriggerAlreadyExists = err{
		code:  http.StatusConflict,
		error: errors.New("Trigger already exists")}
	ErrDatastoreEmptyTrigger = err{
		code:  http.StatusBadRequest,
		error: errors.New("Trigger empty")}
	// move to Fn when merged
	ErrDatastoreFnNotFound = err{
		code:  http.StatusBadRequest,
		error: errors.New("Trigger empty")}
)

func (t *Trigger) Validate() error {
	if t.Name == "" {
		return ErrTriggerMissingName
	}

	if t.AppID == "" {
		return ErrTriggerMissingAppID
	}

	if t.FnID == "" {
		return ErrTriggerMissingFnID
	}

	if t.Type == Unknown {
		return ErrTriggerTypeUnknown
	}

	if t.Source == "" {
		return ErrTriggerMissingSource
	}

	t.Annotations.Validate()

	return nil
}

func (t *Trigger) Clone() *Trigger {
	clone := new(Trigger)
	*clone = *t // shallow copy
	//annotations? Not copied in explicitly others...
	return clone
}

func (t *Trigger) Update(src *Trigger) *Trigger {
	return src
}
