package models

import "github.com/wared2003/freekiosk-hub/internal/repositories"

type TabletDisplay struct {
	repositories.Tablet
	LastReport *repositories.TabletReport
	Groups     []repositories.Group
}
