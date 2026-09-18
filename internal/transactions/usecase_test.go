package transactions

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/BernardBerenes/SupplyHub-API/internal/stores"
)

type fakeRepo struct {
	transactions   []Transaction
	total          int64
	pending        bool
	deletedID      string
	revenueBuckets []RevenueBucket
	earliestPaid   *time.Time
}

func (r *fakeRepo) Create(ctx context.Context, transaction *Transaction) error {
	r.transactions = append(r.transactions, *transaction)
	return nil
}

func (r *fakeRepo) FindByID(ctx context.Context, id string) (*Transaction, error) {
	for _, t := range r.transactions {
		if t.ID == id {
			return &t, nil
		}
	}
	return nil, nil
}

func (r *fakeRepo) FindPaginated(ctx context.Context, filter PaginateFilter, limit, offset int) ([]Transaction, int64, error) {
	return r.transactions, r.total, nil
}

func (r *fakeRepo) ExistsPendingByStoreAndDate(ctx context.Context, storeID int64, date time.Time) (bool, error) {
	return r.pending, nil
}

func (r *fakeRepo) FindAllPending(ctx context.Context) ([]Transaction, error) {
	var pending []Transaction
	for _, t := range r.transactions {
		if t.DeliveryStatus == DELIVERY_STATUS_PENDING {
			pending = append(pending, t)
		}
	}
	return pending, nil
}

func (r *fakeRepo) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	for i, t := range r.transactions {
		if t.ID != id {
			continue
		}
		if store, ok := updates["store"].(StoreSnapshot); ok {
			r.transactions[i].Store = store
		}
		if status, ok := updates["payment_status"].(string); ok {
			r.transactions[i].PaymentStatus = status
		}
		if status, ok := updates["delivery_status"].(string); ok {
			r.transactions[i].DeliveryStatus = status
		}
		if date, ok := updates["date"].(time.Time); ok {
			r.transactions[i].Date = date
		}
	}
	return nil
}

func (r *fakeRepo) SoftDelete(ctx context.Context, id string) (int64, error) {
	if id == r.deletedID {
		return 0, nil
	}
	return 1, nil
}

func (r *fakeRepo) SumRevenueByBucket(ctx context.Context, filter RevenueFilter) ([]RevenueBucket, error) {
	return r.revenueBuckets, nil
}

func (r *fakeRepo) FindEarliestPaidDate(ctx context.Context) (*time.Time, error) {
	return r.earliestPaid, nil
}

type fakeStoreLookup struct {
	stores map[int64]stores.Store
}

func (r *fakeStoreLookup) FindByID(ctx context.Context, id int64) (*stores.Store, error) {
	if s, ok := r.stores[id]; ok {
		return &s, nil
	}
	return nil, nil
}

func (r *fakeStoreLookup) FindByIDIncludingDeleted(ctx context.Context, id int64) (*stores.Store, error) {
	return r.FindByID(ctx, id)
}

func TestCreate_Success(t *testing.T) {
	repo := &fakeRepo{}
	storeRepo := &fakeStoreLookup{stores: map[int64]stores.Store{1: {ID: 1, Name: "Toko Surya"}}}
	uc := NewUseCase(repo, storeRepo)

	date, _ := time.Parse(DateFormat, "2026-09-03")
	if err := uc.Create(context.Background(), CreateInput{StoreID: 1, Date: date}); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if len(repo.transactions) != 1 {
		t.Fatalf("expected transaction persisted, got %+v", repo.transactions)
	}

	created := repo.transactions[0]
	if created.PaymentStatus != PAYMENT_STATUS_UNPAID || created.DeliveryStatus != DELIVERY_STATUS_PENDING {
		t.Fatalf("expected default statuses, got %+v", created)
	}
	if created.Store != (StoreSnapshot{ID: 1, Name: "Toko Surya"}) {
		t.Fatalf("expected store snapshot, got %+v", created.Store)
	}
}

func TestCreate_StoreNotFound(t *testing.T) {
	repo := &fakeRepo{}
	storeRepo := &fakeStoreLookup{stores: map[int64]stores.Store{}}
	uc := NewUseCase(repo, storeRepo)

	date, _ := time.Parse(DateFormat, "2026-09-03")
	err := uc.Create(context.Background(), CreateInput{StoreID: 99, Date: date})
	if !errors.Is(err, ErrStoreNotFound) {
		t.Fatalf("expected ErrStoreNotFound, got %v", err)
	}
}

func TestCreate_RejectsDuplicatePending(t *testing.T) {
	repo := &fakeRepo{pending: true}
	storeRepo := &fakeStoreLookup{stores: map[int64]stores.Store{1: {ID: 1, Name: "Toko Surya"}}}
	uc := NewUseCase(repo, storeRepo)

	date, _ := time.Parse(DateFormat, "2026-09-03")
	err := uc.Create(context.Background(), CreateInput{StoreID: 1, Date: date})
	if !errors.Is(err, ErrDuplicatePending) {
		t.Fatalf("expected ErrDuplicatePending, got %v", err)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	repo := &fakeRepo{}
	uc := NewUseCase(repo, &fakeStoreLookup{})

	status := PAYMENT_STATUS_PAID
	err := uc.Update(context.Background(), "missing", UpdateInput{PaymentStatus: &status})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdate_RejectsStoreChangeWhenNotPending(t *testing.T) {
	repo := &fakeRepo{transactions: []Transaction{
		{ID: "t1", DeliveryStatus: DELIVERY_STATUS_ON_DELIVERY},
	}}
	storeRepo := &fakeStoreLookup{stores: map[int64]stores.Store{2: {ID: 2, Name: "Toko Baru"}}}
	uc := NewUseCase(repo, storeRepo)

	storeID := int64(2)
	err := uc.Update(context.Background(), "t1", UpdateInput{StoreID: &storeID})
	if !errors.Is(err, ErrStoreUpdateNotPending) {
		t.Fatalf("expected ErrStoreUpdateNotPending, got %v", err)
	}
}

func TestUpdate_AllowsStoreChangeWhenPending(t *testing.T) {
	repo := &fakeRepo{transactions: []Transaction{
		{ID: "t1", DeliveryStatus: DELIVERY_STATUS_PENDING, Store: StoreSnapshot{ID: 1, Name: "Toko Surya"}},
	}}
	storeRepo := &fakeStoreLookup{stores: map[int64]stores.Store{2: {ID: 2, Name: "Toko Baru"}}}
	uc := NewUseCase(repo, storeRepo)

	storeID := int64(2)
	if err := uc.Update(context.Background(), "t1", UpdateInput{StoreID: &storeID}); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if repo.transactions[0].Store != (StoreSnapshot{ID: 2, Name: "Toko Baru"}) {
		t.Fatalf("expected store snapshot updated, got %+v", repo.transactions[0].Store)
	}
}

func TestDelete_NotFound(t *testing.T) {
	repo := &fakeRepo{deletedID: "missing"}
	uc := NewUseCase(repo, &fakeStoreLookup{})

	err := uc.Delete(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestSyncStoreNames_UpdatesOnlyPendingWithChangedName(t *testing.T) {
	repo := &fakeRepo{transactions: []Transaction{
		{ID: "t1", DeliveryStatus: DELIVERY_STATUS_PENDING, Store: StoreSnapshot{ID: 1, Name: "Toko Surya"}},
		{ID: "t2", DeliveryStatus: DELIVERY_STATUS_ON_DELIVERY, Store: StoreSnapshot{ID: 1, Name: "Toko Surya"}},
	}}
	storeRepo := &fakeStoreLookup{stores: map[int64]stores.Store{1: {ID: 1, Name: "Toko Surya Baru"}}}
	uc := NewUseCase(repo, storeRepo)

	if err := uc.SyncStoreNames(context.Background()); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if repo.transactions[0].Store.Name != "Toko Surya Baru" {
		t.Fatalf("expected pending transaction synced, got %+v", repo.transactions[0])
	}
	if repo.transactions[1].Store.Name != "Toko Surya" {
		t.Fatalf("expected non-pending transaction untouched, got %+v", repo.transactions[1])
	}
}

func TestRevenue_FillsEmptyMonthBuckets(t *testing.T) {
	repo := &fakeRepo{
		revenueBuckets: []RevenueBucket{
			{Period: "2026-07", Revenue: 5000},
			{Period: "2026-09", Revenue: 3000},
		},
	}
	uc := NewUseCase(repo, &fakeStoreLookup{})

	res, err := uc.Revenue(context.Background(), RevenueRequest{
		DateFrom: "2026-07-01",
		DateTo:   "2026-09-30",
		GroupBy:  GROUP_BY_MONTH,
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if len(res.Points) != 3 {
		t.Fatalf("expected 3 month points, got %d: %+v", len(res.Points), res.Points)
	}

	want := []RevenuePoint{
		{Period: "2026-07", Revenue: 5000},
		{Period: "2026-08", Revenue: 0},
		{Period: "2026-09", Revenue: 3000},
	}
	for i, p := range want {
		if res.Points[i] != p {
			t.Fatalf("point %d mismatch: want %+v, got %+v", i, p, res.Points[i])
		}
	}

	if res.TotalRevenue != 8000 {
		t.Fatalf("expected total 8000, got %d", res.TotalRevenue)
	}
}

func TestRevenue_FillsEmptyDayBuckets(t *testing.T) {
	repo := &fakeRepo{
		revenueBuckets: []RevenueBucket{
			{Period: "2026-09-03", Revenue: 1000},
		},
	}
	uc := NewUseCase(repo, &fakeStoreLookup{})

	res, err := uc.Revenue(context.Background(), RevenueRequest{
		DateFrom: "2026-09-01",
		DateTo:   "2026-09-05",
		GroupBy:  GROUP_BY_DAY,
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if len(res.Points) != 5 {
		t.Fatalf("expected 5 day points, got %d: %+v", len(res.Points), res.Points)
	}
	if res.Points[0].Period != "2026-09-01" || res.Points[4].Period != "2026-09-05" {
		t.Fatalf("unexpected day range: %+v", res.Points)
	}
	if res.Points[2].Revenue != 1000 || res.TotalRevenue != 1000 {
		t.Fatalf("unexpected revenue mapping: %+v total=%d", res.Points, res.TotalRevenue)
	}
}

func TestRevenue_WeekBucketFormat(t *testing.T) {
	repo := &fakeRepo{
		revenueBuckets: []RevenueBucket{
			{Period: "2026-W38", Revenue: 2500},
		},
	}
	uc := NewUseCase(repo, &fakeStoreLookup{})

	res, err := uc.Revenue(context.Background(), RevenueRequest{
		DateFrom: "2026-09-14",
		DateTo:   "2026-09-20",
		GroupBy:  GROUP_BY_WEEK,
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if len(res.Points) != 1 || res.Points[0].Period != "2026-W38" {
		t.Fatalf("expected single 2026-W38 bucket, got %+v", res.Points)
	}
	if res.Points[0].Revenue != 2500 {
		t.Fatalf("expected week revenue 2500, got %d", res.Points[0].Revenue)
	}
}

func TestAutoGroupBy(t *testing.T) {
	cases := []struct {
		from, to string
		want     string
	}{
		{"2026-09-18", "2026-09-18", GROUP_BY_TOTAL},
		{"2026-08-19", "2026-09-18", GROUP_BY_DAY},
		{"2025-09-18", "2026-09-18", GROUP_BY_MONTH},
	}
	for _, c := range cases {
		from, _ := time.Parse(DateFormat, c.from)
		to, _ := time.Parse(DateFormat, c.to)
		if got := autoGroupBy(from, to); got != c.want {
			t.Fatalf("autoGroupBy(%s,%s) = %s, want %s", c.from, c.to, got, c.want)
		}
	}
}

func TestRevenue_PeriodAllUsesEarliestPaidDate(t *testing.T) {
	earliest, _ := time.Parse(DateFormat, "2026-01-15")
	repo := &fakeRepo{
		earliestPaid:   &earliest,
		revenueBuckets: []RevenueBucket{{Period: "2026-01", Revenue: 4000}},
	}
	uc := NewUseCase(repo, &fakeStoreLookup{})

	res, err := uc.Revenue(context.Background(), RevenueRequest{Period: PERIOD_ALL})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if res.DateFrom != "2026-01-15" {
		t.Fatalf("expected date_from from earliest paid date, got %s", res.DateFrom)
	}
	if res.GroupBy != GROUP_BY_MONTH {
		t.Fatalf("expected month grouping for all-time span, got %s", res.GroupBy)
	}
}
