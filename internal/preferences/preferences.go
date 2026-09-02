package preferences

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/peterintech/briefed/internal/database"
)

func Replace(ctx context.Context, conn *sql.DB, db *database.Queries, userID uuid.UUID, categoryIDs, feedIDs []uuid.UUID) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queries := db.WithTx(tx)
	if err := queries.DeleteUserCategories(ctx, userID); err != nil {
		return err
	}
	currentFeeds, err := queries.GetFeedsForUser(ctx, userID)
	if err != nil {
		return err
	}
	for _, feed := range currentFeeds {
		if err := queries.DeleteFeedFollow(ctx, database.DeleteFeedFollowParams{
			UserID: userID, FeedID: feed.ID,
		}); err != nil {
			return err
		}
	}
	for _, categoryID := range categoryIDs {
		if err := queries.CreateUserCategory(ctx, database.CreateUserCategoryParams{
			UserID: userID, CategoryID: categoryID,
		}); err != nil {
			return err
		}
	}
	for _, feedID := range feedIDs {
		if err := queries.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
			UserID: userID, FeedID: feedID,
		}); err != nil {
			return err
		}
	}
	return tx.Commit()
}
