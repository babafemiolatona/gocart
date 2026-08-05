package repositories

import "gorm.io/gorm"

// TransactionScope exposes repositories bound to a single connection. The
// concrete UnitOfWork binds them to one database transaction inside
// TransactionManager.WithTransaction.
type TransactionScope interface {
	Auth() AuthRepository
	User() UserRepository
	Category() CategoryRepository
	Cart() CartRepository
	Merchant() MerchantRepository
	Order() OrderRepository
	Payment() PaymentRepository
	Product() ProductRepository
	ProductImage() ProductImageRepository
}

// TransactionManager runs a unit of work. The concrete UnitOfWork executes it
// inside a single database transaction; an interface allows services to be
// unit-tested with a fake scope.
type TransactionManager interface {
	WithTransaction(fn func(scope TransactionScope) error) error
}

// UnitOfWork groups repositories bound to a single connection. Calling
// WithTransaction scopes them to one database transaction, so services never
// handle a *gorm.DB directly.
type UnitOfWork struct {
	db *gorm.DB
}

func NewUnitOfWork(db *gorm.DB) *UnitOfWork {
	return &UnitOfWork{db: db}
}

func (u *UnitOfWork) WithTransaction(fn func(scope TransactionScope) error) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		return fn(&UnitOfWork{db: tx})
	})
}

func (u *UnitOfWork) Auth() AuthRepository {
	return &authRepository{db: u.db}
}

func (u *UnitOfWork) User() UserRepository {
	return &userRepository{db: u.db}
}

func (u *UnitOfWork) Category() CategoryRepository {
	return &categoryRepository{db: u.db}
}

func (u *UnitOfWork) Cart() CartRepository {
	return &cartRepository{db: u.db}
}

func (u *UnitOfWork) Merchant() MerchantRepository {
	return &merchantRepository{db: u.db}
}

func (u *UnitOfWork) Order() OrderRepository {
	return &orderRepository{db: u.db}
}

func (u *UnitOfWork) Payment() PaymentRepository {
	return &paymentRepository{db: u.db}
}

func (u *UnitOfWork) Product() ProductRepository {
	return &productRepository{db: u.db}
}

func (u *UnitOfWork) ProductImage() ProductImageRepository {
	return &productImageRepository{db: u.db}
}
