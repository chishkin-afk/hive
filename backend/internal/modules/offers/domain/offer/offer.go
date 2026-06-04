package offer

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidAuthorID    = errors.New("author's id is invalid")
	ErrInvalidTitle       = errors.New("len of title must be more than 3 and less than 128")
	ErrInvalidDescription = errors.New("len of description must be more than 3 and less than 512")
)

type Offer struct {
	id           uuid.UUID
	authorID     uuid.UUID
	title        string
	description  string // can be empty
	slotDuration SlotDuration
	slotInterval SlotInterval
	isActive     IsActive
	createdAt    time.Time
	updatedAt    time.Time
}

func New(
	authorID uuid.UUID,
	title, description string,
	slotDuration SlotDuration,
	slotInterval SlotInterval,
) (*Offer, error) {
	if authorID == uuid.Nil {
		return nil, ErrInvalidAuthorID
	}

	title = strings.TrimSpace(title)
	if !isFit(title, 3, 128) {
		return nil, ErrInvalidTitle
	}

	description = strings.TrimSpace(description)
	if description != "" && !isFit(description, 3, 512) {
		return nil, ErrInvalidDescription
	}

	if err := slotDuration.Validate(); err != nil {
		return nil, err
	}

	if err := slotInterval.Validate(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &Offer{
		id:           uuid.New(),
		authorID:     authorID,
		title:        title,
		description:  description,
		slotDuration: slotDuration,
		slotInterval: slotInterval,
		isActive:     true,
		createdAt:    now,
		updatedAt:    now,
	}, nil
}

func From(
	id uuid.UUID,
	authorID uuid.UUID,
	title, description string,
	slotDuration SlotDuration,
	slotInterval SlotInterval,
	isActive IsActive,
	createdAt, updatedAt time.Time,
) *Offer {
	return &Offer{
		id:           id,
		authorID:     authorID,
		title:        title,
		description:  description,
		slotDuration: slotDuration,
		slotInterval: slotInterval,
		isActive:     isActive,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

func (o *Offer) ChangeTitle(newTitle string) error {
	newTitle = strings.TrimSpace(newTitle)
	if !isFit(newTitle, 3, 128) {
		return ErrInvalidTitle
	}

	o.title = newTitle
	o.updatedAt = time.Now().UTC()

	return nil
}

func (o *Offer) ChangeDescription(newDescription string) error {
	newDescription = strings.TrimSpace(newDescription)
	if newDescription != "" && !isFit(newDescription, 3, 512) {
		return ErrInvalidDescription
	}

	o.description = newDescription
	o.updatedAt = time.Now().UTC()

	return nil
}

func (o *Offer) ChangeSlotDuration(newSlotDuration SlotDuration) error {
	if err := newSlotDuration.Validate(); err != nil {
		return err
	}

	o.slotDuration = newSlotDuration
	o.updatedAt = time.Now().UTC()

	return nil
}

func (o *Offer) ChangeSlotInterval(newSlotInterval SlotInterval) error {
	if err := newSlotInterval.Validate(); err != nil {
		return err
	}

	o.slotInterval = newSlotInterval
	o.updatedAt = time.Now().UTC()

	return nil
}

func (o *Offer) Enable() error {
	newIsActive, err := o.isActive.Enable()
	if err != nil {
		return err
	}

	o.isActive = newIsActive
	o.updatedAt = time.Now().UTC()

	return nil
}

func (o *Offer) Disable() error {
	newIsActive, err := o.isActive.Disable()
	if err != nil {
		return err
	}

	o.isActive = newIsActive
	o.updatedAt = time.Now().UTC()

	return nil
}

func (o *Offer) ID() uuid.UUID {
	return o.id
}

func (o *Offer) AuthorID() uuid.UUID {
	return o.authorID
}

func (o *Offer) Title() string {
	return o.title
}

func (o *Offer) Description() string {
	return o.description
}

func (o *Offer) SlotDuration() SlotDuration {
	return o.slotDuration
}

func (o *Offer) SlotInterval() SlotInterval {
	return o.slotInterval
}

func (o *Offer) IsActive() IsActive {
	return o.isActive
}

func (o *Offer) CreatedAt() time.Time {
	return o.createdAt
}

func (o *Offer) UpdatedAt() time.Time {
	return o.updatedAt
}

func isFit(str string, min, max int) bool {
	len := len([]rune(str))
	return len > min && len <= max
}
