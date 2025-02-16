package transaction_usecase

import (
	"context"
	"github.com/smthjapanese/avito-merch/internal/entity"
)

type TransactionUC struct {
	userRepo UserRepository
	txRepo   Repository
	dbTx     DBTransactor
}

func NewTransactionUC(
	userRepo UserRepository,
	txRepo Repository,
	dbTx DBTransactor,
) *TransactionUC {
	return &TransactionUC{
		userRepo: userRepo,
		txRepo:   txRepo,
		dbTx:     dbTx,
	}
}

func (uc *TransactionUC) CreateTransfer(ctx context.Context, fromUserID, toUserID int64, amount int64) error {
	// Validation errors return as is
	if amount <= 0 {
		return entity.ErrNegativeAmount
	}

	// Technical errors wrapped in ErrTransactionFailed
	err := uc.dbTx.WithinTransaction(ctx, func(ctx context.Context) error {
		fromUser, err := uc.userRepo.GetByID(ctx, fromUserID)
		if err != nil {
			return err
		}
		if fromUser == nil {
			return entity.ErrUserNotFound
		}

		toUser, err := uc.userRepo.GetByID(ctx, toUserID)
		if err != nil {
			return err
		}
		if toUser == nil {
			return entity.ErrUserNotFound
		}

		// Business validation returns specific errors
		if fromUser.Coins < amount {
			return entity.ErrInsufficientFunds
		}

		// Update balances
		fromUser.Coins -= amount
		toUser.Coins += amount

		if err := uc.userRepo.Update(ctx, fromUser); err != nil {
			return err
		}
		if err := uc.userRepo.Update(ctx, toUser); err != nil {
			return err
		}

		tx := entity.Transaction{
			FromUserID: fromUserID,
			ToUserID:   toUserID,
			Amount:     amount,
			Type:       entity.TransactionTypeTransfer,
		}

		if err := uc.txRepo.Create(ctx, tx); err != nil {
			return err
		}

		return nil
	})

	// Wrap technical errors, but pass through validation errors
	if err != nil {
		switch err {
		case entity.ErrInsufficientFunds, entity.ErrNegativeAmount:
			return err
		default:
			return entity.ErrTransactionFailed
		}
	}

	return nil
}

func (uc *TransactionUC) GetUserHistory(ctx context.Context, userID int64) (*TransactionHistory, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, entity.ErrUserNotFound
	}

	transactions, err := uc.txRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	received := make([]TransactionInfo, 0)
	sent := make([]TransactionInfo, 0)

	for _, tx := range transactions {
		var info TransactionInfo
		var otherUserName string

		if tx.ToUserID == userID {
			// Received transaction
			fromUser, err := uc.userRepo.GetByID(ctx, tx.FromUserID)
			if err != nil {
				continue // Skip if user not found
			}
			otherUserName = fromUser.Username
			info = TransactionInfo{
				ID:        tx.ID,
				User:      otherUserName,
				Amount:    tx.Amount,
				Type:      tx.Type,
				CreatedAt: tx.CreatedAt,
			}
			received = append(received, info)
		} else {
			// Sent transaction
			toUser, err := uc.userRepo.GetByID(ctx, tx.ToUserID)
			if err != nil {
				continue // Skip if user not found
			}
			otherUserName = toUser.Username
			info = TransactionInfo{
				ID:        tx.ID,
				User:      otherUserName,
				Amount:    tx.Amount,
				Type:      tx.Type,
				CreatedAt: tx.CreatedAt,
			}
			sent = append(sent, info)
		}
	}

	return &TransactionHistory{
		Received: received,
		Sent:     sent,
	}, nil
}
