package repository

import (
  "loopi-api/internal/domain"
)

type WorkConfigRepository interface {
  GetActiveConfig() domain.WorkConfig
}
