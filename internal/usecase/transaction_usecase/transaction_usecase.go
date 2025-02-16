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
	if amount <= 0 {
		return entity.ErrNegativeAmount
	}

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

		if fromUser.Coins < amount {
			return entity.ErrInsufficientFunds
		}

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
			fromUser, err := uc.userRepo.GetByID(ctx, tx.FromUserID)
			if err != nil {
				continue
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
			toUser, err := uc.userRepo.GetByID(ctx, tx.ToUserID)
			if err != nil {
				continue
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
