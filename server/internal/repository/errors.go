package repository

import "errors"

var ErrNotFound = errors.New("not found")

var ErrInvalidGoalType = errors.New("field not valid for this goal's type")

var ErrDuplicateMessage = errors.New("duplicate whatsapp message")

var ErrCodeTaken = errors.New("link code already exists")
