package usecase

import (
	"context"
	"github.com/smthjapanese/avito-merch/internal/entity"
	"time"
)

type MerchUseCase struct {
	merchRepo    MerchRepository
	userRepo     UserRepository
	invRepo      InventoryRepository
	txRepo       TransactionRepository
	dbTransactor DBTransactor
}

func NewMerchUseCase(
	merchRepo MerchRepository,
	userRepo UserRepository,
	invRepo InventoryRepository,
	txRepo TransactionRepository,
	dbTransactor DBTransactor,
) MerchUseCase {
	return MerchUseCase{
		merchRepo:    merchRepo,
		userRepo:     userRepo,
		invRepo:      invRepo,
		txRepo:       txRepo,
		dbTransactor: dbTransactor,
	}
}

func (uc *MerchUseCase) ListAvailable(ctx context.Context) ([]MerchItemDTO, error) {
	items, err := uc.merchRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]MerchItemDTO, len(items))
	for i, item := range items {
		result[i] = MerchItemDTO{
			ID:    item.ID,
			Name:  item.Name,
			Price: item.Price,
		}
	}
	return result, nil
}

func (uc *MerchUseCase) BuyItem(ctx context.Context, userID int64, itemName string) error {
	return uc.dbTransactor.WithinTransaction(ctx, func(ctx context.Context) error {
		item, err := uc.merchRepo.GetByName(ctx, itemName)
		if err != nil {
			return entity.ErrMerchNotFound
		}

		user, err := uc.userRepo.GetByID(ctx, userID)
		if err != nil {
			return entity.ErrUserNotFound
		}

		if user.Coins < item.Price {
			return entity.ErrInsufficientFunds
		}

		user.Coins -= item.Price
		if err := uc.userRepo.Update(ctx, user); err != nil {
			return err
		}

		inventory, err := uc.invRepo.GetByUserID(ctx, userID)
		var found bool
		for i := range inventory {
			if inventory[i].ItemID == item.ID {
				inventory[i].Quantity++
				if err := uc.invRepo.Update(ctx, inventory[i]); err != nil {
					return err
				}
				found = true
				break
			}
		}

		if !found {
			newInventory := entity.UserInventory{
				UserID:      userID,
				ItemID:      item.ID,
				Quantity:    1,
				PurchasedAt: time.Now(),
			}
			if err := uc.invRepo.Create(ctx, newInventory); err != nil {
				return err
			}
		}

		transaction := entity.Transaction{
			FromUserID: userID,
			ToUserID:   userID, // self-transaction for purchase
			Amount:     item.Price,
			Type:       entity.TransactionTypePurchase,
			ItemID:     &item.ID,
			CreatedAt:  time.Now(),
		}

		if err := uc.txRepo.Create(ctx, transaction); err != nil {
			return err
		}

		return nil
	})
}
