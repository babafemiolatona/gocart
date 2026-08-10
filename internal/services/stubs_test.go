package services

import (
	"gocart/internal/dto"
	"gocart/internal/models"
	"gocart/internal/query"
	"gocart/internal/repositories"
	"mime/multipart"
)

// stubAuthRepo implements repositories.AuthRepository for tests.
type stubAuthRepo struct {
	createFn          func(user *models.User) error
	existsFn          func(email string) (bool, error)
	getByEmailFn      func(email string) (*models.User, error)
	getByIdentifierFn func(identifier string) (*models.User, error)
	getByIDFn         func(id uint) (*models.User, error)
	updatePasswordFn  func(id uint, hashedPassword string) error
}

func (s *stubAuthRepo) Create(u *models.User) error {
	if s.createFn != nil {
		return s.createFn(u)
	}
	return nil
}

func (s *stubAuthRepo) ExistsByEmail(email string) (bool, error) {
	if s.existsFn != nil {
		return s.existsFn(email)
	}
	return false, nil
}

func (s *stubAuthRepo) GetByEmail(email string) (*models.User, error) {
	if s.getByEmailFn != nil {
		return s.getByEmailFn(email)
	}
	return nil, repositories.ErrRecordNotFound
}

func (s *stubAuthRepo) GetByEmailOrUsername(identifier string) (*models.User, error) {
	if s.getByIdentifierFn != nil {
		return s.getByIdentifierFn(identifier)
	}
	return nil, repositories.ErrRecordNotFound
}

func (s *stubAuthRepo) GetByID(id uint) (*models.User, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(id)
	}
	return nil, repositories.ErrRecordNotFound
}

func (s *stubAuthRepo) UpdatePassword(id uint, hashedPassword string) error {
	if s.updatePasswordFn != nil {
		return s.updatePasswordFn(id, hashedPassword)
	}
	return nil
}

// stubCartRepo implements repositories.CartRepository for tests.
type stubCartRepo struct {
	createFn       func(cart *models.Cart) error
	getWithItemsFn func(userID uint) (*models.Cart, error)
	addItemFn      func(item *models.CartItem) error
	updateItemFn   func(item *models.CartItem) error
	removeItemFn   func(cartItemID uint) error
	updateTotalFn  func(cartID uint, total int64, itemCount int) error
	clearCartFn    func(cartID uint) error
}

func (s *stubCartRepo) Create(c *models.Cart) error {
	if s.createFn != nil {
		return s.createFn(c)
	}
	return nil
}

func (s *stubCartRepo) GetWithItems(userID uint) (*models.Cart, error) {
	if s.getWithItemsFn != nil {
		return s.getWithItemsFn(userID)
	}
	return nil, repositories.ErrRecordNotFound
}

func (s *stubCartRepo) AddItem(i *models.CartItem) error {
	if s.addItemFn != nil {
		return s.addItemFn(i)
	}
	return nil
}

func (s *stubCartRepo) UpdateItem(i *models.CartItem) error {
	if s.updateItemFn != nil {
		return s.updateItemFn(i)
	}
	return nil
}

func (s *stubCartRepo) RemoveItem(id uint) error {
	if s.removeItemFn != nil {
		return s.removeItemFn(id)
	}
	return nil
}

func (s *stubCartRepo) UpdateCartTotal(cartID uint, total int64, itemCount int) error {
	if s.updateTotalFn != nil {
		return s.updateTotalFn(cartID, total, itemCount)
	}
	return nil
}

func (s *stubCartRepo) ClearCart(cartID uint) error {
	if s.clearCartFn != nil {
		return s.clearCartFn(cartID)
	}
	return nil
}

// stubCategoryRepo implements repositories.CategoryRepository for tests.
type stubCategoryRepo struct {
	createFn  func(c *models.Category) error
	getByIDFn func(id uint) (*models.Category, error)
	getAllFn  func() ([]models.Category, error)
	updateFn  func(id uint, values map[string]interface{}) error
	deleteFn  func(id uint) error
}

func (s *stubCategoryRepo) Create(c *models.Category) error {
	if s.createFn != nil {
		return s.createFn(c)
	}
	return nil
}

func (s *stubCategoryRepo) GetByID(id uint) (*models.Category, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(id)
	}
	return nil, repositories.ErrRecordNotFound
}

func (s *stubCategoryRepo) GetAll() ([]models.Category, error) {
	if s.getAllFn != nil {
		return s.getAllFn()
	}
	return nil, nil
}

func (s *stubCategoryRepo) Update(id uint, values map[string]interface{}) error {
	if s.updateFn != nil {
		return s.updateFn(id, values)
	}
	return nil
}

func (s *stubCategoryRepo) Delete(id uint) error {
	if s.deleteFn != nil {
		return s.deleteFn(id)
	}
	return nil
}

// stubMerchantRepo implements repositories.MerchantRepository for tests.
type stubMerchantRepo struct {
	createFn    func(m *models.Merchant) error
	getByIDFn   func(id uint) (*models.Merchant, error)
	getByUserFn func(userID uint) (*models.Merchant, error)
	updateFn    func(m *models.Merchant) error
}

func (s *stubMerchantRepo) Create(m *models.Merchant) error {
	if s.createFn != nil {
		return s.createFn(m)
	}
	return nil
}

func (s *stubMerchantRepo) GetByID(id uint) (*models.Merchant, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(id)
	}
	return nil, repositories.ErrRecordNotFound
}

func (s *stubMerchantRepo) GetByUserID(userID uint) (*models.Merchant, error) {
	if s.getByUserFn != nil {
		return s.getByUserFn(userID)
	}
	return nil, repositories.ErrRecordNotFound
}

func (s *stubMerchantRepo) Update(m *models.Merchant) error {
	if s.updateFn != nil {
		return s.updateFn(m)
	}
	return nil
}

// stubOrderRepo implements repositories.OrderRepository for tests.
type stubOrderRepo struct {
	createFn          func(order *models.Order) error
	getByIDFn         func(id uint) (*models.Order, error)
	getByIdemFn       func(userID uint, key string) (*models.Order, error)
	getByUserFn       func(userID uint, p *dto.PaginationQuery) ([]models.Order, int64, error)
	updateStatusFn    func(orderID uint, status models.OrderStatus) error
	transitionFn      func(orderID uint, from, to models.OrderStatus) (bool, error)
	getByMerchantFn   func(merchantID uint, p *dto.PaginationQuery) ([]models.Order, int64, error)
	getMerchantByIDFn func(merchantID uint, orderID uint) (*models.Order, error)
	countByMerchantFn func(merchantID uint) (int64, error)
	countByStatusFn   func(merchantID uint, status models.OrderStatus) (int64, error)
	sumRevenueFn      func(merchantID uint) (int64, error)
	getRecentFn       func(merchantID uint, limit int) ([]models.Order, error)
}

func (s *stubOrderRepo) CreateOrder(o *models.Order) error {
	if s.createFn != nil {
		return s.createFn(o)
	}
	return nil
}

func (s *stubOrderRepo) GetOrderByID(id uint) (*models.Order, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(id)
	}
	return nil, repositories.ErrRecordNotFound
}

func (s *stubOrderRepo) GetByUserIDAndIdempotencyKey(userID uint, key string) (*models.Order, error) {
	if s.getByIdemFn != nil {
		return s.getByIdemFn(userID, key)
	}
	return nil, repositories.ErrRecordNotFound
}

func (s *stubOrderRepo) GetOrdersByUserID(userID uint, p *dto.PaginationQuery) ([]models.Order, int64, error) {
	if s.getByUserFn != nil {
		return s.getByUserFn(userID, p)
	}
	return nil, 0, nil
}

func (s *stubOrderRepo) UpdateOrderStatus(orderID uint, status models.OrderStatus) error {
	if s.updateStatusFn != nil {
		return s.updateStatusFn(orderID, status)
	}
	return nil
}

func (s *stubOrderRepo) TransitionOrderStatus(orderID uint, from, to models.OrderStatus) (bool, error) {
	if s.transitionFn != nil {
		return s.transitionFn(orderID, from, to)
	}
	return false, nil
}

func (s *stubOrderRepo) GetOrdersByMerchantID(merchantID uint, p *dto.PaginationQuery) ([]models.Order, int64, error) {
	if s.getByMerchantFn != nil {
		return s.getByMerchantFn(merchantID, p)
	}
	return nil, 0, nil
}

func (s *stubOrderRepo) GetMerchantOrderByID(merchantID uint, orderID uint) (*models.Order, error) {
	if s.getMerchantByIDFn != nil {
		return s.getMerchantByIDFn(merchantID, orderID)
	}
	return nil, repositories.ErrRecordNotFound
}

func (s *stubOrderRepo) CountByMerchant(merchantID uint) (int64, error) {
	if s.countByMerchantFn != nil {
		return s.countByMerchantFn(merchantID)
	}
	return 0, nil
}

func (s *stubOrderRepo) CountByMerchantAndStatus(merchantID uint, status models.OrderStatus) (int64, error) {
	if s.countByStatusFn != nil {
		return s.countByStatusFn(merchantID, status)
	}
	return 0, nil
}

func (s *stubOrderRepo) SumRevenueByMerchant(merchantID uint) (int64, error) {
	if s.sumRevenueFn != nil {
		return s.sumRevenueFn(merchantID)
	}
	return 0, nil
}

func (s *stubOrderRepo) GetRecentOrdersByMerchant(merchantID uint, limit int) ([]models.Order, error) {
	if s.getRecentFn != nil {
		return s.getRecentFn(merchantID, limit)
	}
	return nil, nil
}

// stubPaymentRepo implements repositories.PaymentRepository for tests.
type stubPaymentRepo struct {
	createFn         func(p *models.Payment) error
	getByReferenceFn func(reference string) (*models.Payment, error)
	getByOrderIDFn   func(orderID uint) (*models.Payment, error)
	transitionFn     func(reference string, from, to models.PaymentStatus) (bool, error)
}

func (s *stubPaymentRepo) Create(p *models.Payment) error {
	if s.createFn != nil {
		return s.createFn(p)
	}
	return nil
}

func (s *stubPaymentRepo) GetByReference(reference string) (*models.Payment, error) {
	if s.getByReferenceFn != nil {
		return s.getByReferenceFn(reference)
	}
	return nil, repositories.ErrRecordNotFound
}

func (s *stubPaymentRepo) GetByOrderID(orderID uint) (*models.Payment, error) {
	if s.getByOrderIDFn != nil {
		return s.getByOrderIDFn(orderID)
	}
	return nil, repositories.ErrRecordNotFound
}

func (s *stubPaymentRepo) TransitionStatus(reference string, from, to models.PaymentStatus) (bool, error) {
	if s.transitionFn != nil {
		return s.transitionFn(reference, from, to)
	}
	return false, nil
}

// stubProductRepo implements repositories.ProductRepository for tests.
type stubProductRepo struct {
	createFn          func(p *models.Product) error
	getByIDFn         func(id uint) (*models.Product, error)
	getAllFn          func(query *dto.PaginationQuery, filters *query.ProductFilters) ([]models.Product, int64, error)
	updateFn          func(id uint, values map[string]interface{}) error
	deleteFn          func(id uint) error
	incrementFn       func(id uint, qty int) error
	decrementFn       func(id uint, qty int) error
	countByMerchantFn func(merchantID uint) (int64, error)
	countLowStockFn   func(merchantID uint, threshold int) (int64, error)
}

func (s *stubProductRepo) Create(p *models.Product) error {
	if s.createFn != nil {
		return s.createFn(p)
	}
	return nil
}

func (s *stubProductRepo) GetByID(id uint) (*models.Product, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(id)
	}
	return nil, repositories.ErrRecordNotFound
}

func (s *stubProductRepo) GetAll(query *dto.PaginationQuery, filters *query.ProductFilters) ([]models.Product, int64, error) {
	if s.getAllFn != nil {
		return s.getAllFn(query, filters)
	}
	return nil, 0, nil
}

func (s *stubProductRepo) Update(id uint, values map[string]interface{}) error {
	if s.updateFn != nil {
		return s.updateFn(id, values)
	}
	return nil
}

func (s *stubProductRepo) Delete(id uint) error {
	if s.deleteFn != nil {
		return s.deleteFn(id)
	}
	return nil
}

func (s *stubProductRepo) IncrementStock(id uint, qty int) error {
	if s.incrementFn != nil {
		return s.incrementFn(id, qty)
	}
	return nil
}

func (s *stubProductRepo) DecrementStock(id uint, qty int) error {
	if s.decrementFn != nil {
		return s.decrementFn(id, qty)
	}
	return nil
}

func (s *stubProductRepo) CountByMerchant(merchantID uint) (int64, error) {
	if s.countByMerchantFn != nil {
		return s.countByMerchantFn(merchantID)
	}
	return 0, nil
}

func (s *stubProductRepo) CountLowStockByMerchant(merchantID uint, threshold int) (int64, error) {
	if s.countLowStockFn != nil {
		return s.countLowStockFn(merchantID, threshold)
	}
	return 0, nil
}

// stubProductImageRepo implements repositories.ProductImageRepository for tests.
type stubProductImageRepo struct {
	createManyFn func(images []models.ProductImage) error
}

func (s *stubProductImageRepo) CreateMany(images []models.ProductImage) error {
	if s.createManyFn != nil {
		return s.createManyFn(images)
	}
	return nil
}

// stubUserRepo implements repositories.UserRepository for tests.
type stubUserRepo struct {
	getByIDFn func(id uint) (*models.User, error)
}

func (s *stubUserRepo) GetByID(id uint) (*models.User, error) {
	if s.getByIDFn != nil {
		return s.getByIDFn(id)
	}
	return nil, repositories.ErrRecordNotFound
}

// stubStorage implements storage.Storage for tests.
type stubStorage struct {
	uploadFn func(file multipart.File, header *multipart.FileHeader, productID uint) (string, error)
	deleteFn func(objectName string) error
}

func (s *stubStorage) UploadProductImage(file multipart.File, header *multipart.FileHeader, productID uint) (string, error) {
	if s.uploadFn != nil {
		return s.uploadFn(file, header, productID)
	}
	return "", nil
}

func (s *stubStorage) DeleteObject(objectName string) error {
	if s.deleteFn != nil {
		return s.deleteFn(objectName)
	}
	return nil
}

// fakeScope implements repositories.TransactionScope backed by the supplied stubs.
type fakeScope struct {
	auth       repositories.AuthRepository
	cart       repositories.CartRepository
	category   repositories.CategoryRepository
	merchant   repositories.MerchantRepository
	order      repositories.OrderRepository
	payment    repositories.PaymentRepository
	product    repositories.ProductRepository
	productImg repositories.ProductImageRepository
	user       repositories.UserRepository
}

func (s *fakeScope) Auth() repositories.AuthRepository                 { return s.auth }
func (s *fakeScope) User() repositories.UserRepository                 { return s.user }
func (s *fakeScope) Category() repositories.CategoryRepository         { return s.category }
func (s *fakeScope) Cart() repositories.CartRepository                 { return s.cart }
func (s *fakeScope) Merchant() repositories.MerchantRepository         { return s.merchant }
func (s *fakeScope) Order() repositories.OrderRepository               { return s.order }
func (s *fakeScope) Payment() repositories.PaymentRepository           { return s.payment }
func (s *fakeScope) Product() repositories.ProductRepository           { return s.product }
func (s *fakeScope) ProductImage() repositories.ProductImageRepository { return s.productImg }

// fakeTxManager implements repositories.TransactionManager. The transaction
// callback runs immediately with the configured scope.
type fakeTxManager struct {
	scope repositories.TransactionScope
}

func (f *fakeTxManager) WithTransaction(fn func(scope repositories.TransactionScope) error) error {
	return fn(f.scope)
}
