package data

import "database/sql"

type Models struct {
	Users        UserModel
	EmailTokens  EmailTokenModel
	Items        ItemModel
	ItemImages   ItemImageModel
	PickupPoints PickupPointModel
	Rentals      RentalModel
	Reviews      ReviewModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Users:        UserModel{DB: db},
		EmailTokens:  EmailTokenModel{DB: db},
		Items:        ItemModel{DB: db},
		ItemImages:   ItemImageModel{DB: db},
		PickupPoints: PickupPointModel{DB: db},
		Rentals:      RentalModel{DB: db},
		Reviews:      ReviewModel{DB: db},
	}
}
