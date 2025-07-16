package dbo

import (
	"context"
	"fmt"
)

func (q *Queries) UpdatePageNum(ctx context.Context, id int64, pageNum int64) error {
	return q.Transaction(ctx, func(sub *Queries) error {
		if err := sub.movePages(ctx, pageNum); err != nil {
			return fmt.Errorf("move pages to have slot: %w", err)
		}
		if err := sub.setPageNum(ctx, setPageNumParams{
			Num: pageNum,
			ID:  id,
		}); err != nil {
			return fmt.Errorf("set desired page num: %w", err)
		}
		if err := sub.compressPagesSequence(ctx); err != nil {
			return fmt.Errorf("compress pages sequence: %w", err)
		}
		return nil
	})
}

func (q *Queries) DeletePage(ctx context.Context, id int64) error {
	return q.Transaction(ctx, func(sub *Queries) error {
		if err := sub.deletePage(ctx, id); err != nil {
			return fmt.Errorf("delete page: %w", err)
		}
		if err := sub.compressPagesSequence(ctx); err != nil {
			return fmt.Errorf("compress pages sequence: %w", err)
		}
		return nil
	})
}

func (p *ListBlueprintPagesRow) IsIncluded() bool {
	return p != nil && p.Inline != nil
}

func (p *ListBlueprintPagesRow) IsSummary() bool {
	return p.IsIncluded() && !(*p.Inline)
}

func (p *ListBlueprintPagesRow) IsFull() bool {

	return p.IsIncluded() && (*p.Inline)
}
