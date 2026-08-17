package repository

import "errors"

var ErrNotFound = errors.New("not found")

var ErrDuplicateMessage = errors.New("duplicate whatsapp message")

var ErrCodeTaken = errors.New("link code already exists")
