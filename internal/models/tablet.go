package models

import "freekiosk-hub/internal/repositories"

type TabletDisplay struct {
	repositories.Tablet
	LastReport *repositories.TabletReport
	Groups     []repositories.Group
}
