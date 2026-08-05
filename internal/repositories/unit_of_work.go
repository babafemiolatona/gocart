package repositories

import "gorm.io/gorm"

// UnitOfWork groups repositories bound to a single connection. Calling
// WithTransaction scopes them to one database transaction, so services never
// handle a *gorm.DB directly.
type UnitOfWork struct {
	db *gorm.DB
}

func NewUnitOfWork(db *gorm.DB) *UnitOfWork {
	return &UnitOfWork{db: db}
}

func (u *UnitOfWork) WithTransaction(fn func(uow *UnitOfWork) error) error {
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
