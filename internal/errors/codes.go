package errors

const (
	CodeValidationError = "validation_error"
	CodeInternalServer  = "internal_server_error"

	// Auth
	CodeUnauthorized     = "unauthorized"
	CodeInvalidToken     = "invalid_token"
	CodeMissingAuth      = "missing_auth_header"
	CodeInvalidAuth      = "invalid_auth_header"
	CodeForbidden        = "forbidden"
	CodeAccessDenied     = "access_denied"
	CodeMerchantRequired = "merchant_required"
	CodeVerifyMerchant   = "verify_merchant_failed"

	// User / auth domain
	CodeInvalidCredentials = "invalid_credentials"
	CodeUserExists         = "user_exists"
	CodeUserNotFound       = "user_not_found"
	CodeFetchUser          = "fetch_user_failed"
	CodeCreateUser         = "create_user_failed"
	CodeCheckUser          = "check_user_failed"
	CodeHashPassword       = "hash_password_failed"
	CodeGenerateToken      = "generate_token_failed"
	CodeUpdatePassword     = "update_password_failed"

	// Category
	CodeCategoryExists   = "category_exists"
	CodeCategoryNotFound = "category_not_found"
	CodeCreateCategory   = "create_category_failed"
	CodeFetchCategory    = "fetch_category_failed"
	CodeFetchCategories  = "fetch_categories_failed"
	CodeUpdateCategory   = "update_category_failed"
	CodeDeleteCategory   = "delete_category_failed"

	// Product
	CodeProductExists    = "product_exists"
	CodeProductNotFound  = "product_not_found"
	CodeCreateProduct    = "create_product_failed"
	CodeFetchProduct     = "fetch_product_failed"
	CodeFetchProducts    = "fetch_products_failed"
	CodeUpdateProduct    = "update_product_failed"
	CodeDeleteProduct    = "delete_product_failed"
	CodeUploadImages     = "upload_product_images_failed"
	CodeFileTooLarge     = "file_too_large"
	CodeInvalidImageType = "invalid_image_type"

	// Cart
	CodeCartEmpty         = "cart_empty"
	CodeFetchCart         = "fetch_cart_failed"
	CodeCreateCart        = "create_cart_failed"
	CodeAddCartItem       = "add_cart_item_failed"
	CodeUpdateCartItem    = "update_cart_item_failed"
	CodeUpdateCart        = "update_cart_failed"
	CodeRemoveCartItem    = "remove_cart_item_failed"
	CodeClearCart         = "clear_cart_failed"
	CodeCartItemNotFound  = "cart_item_not_found"
	CodeMultipleMerchants = "multiple_merchants_not_supported"
	CodeInsufficientStock = "insufficient_stock"
	CodeInvalidQuantity   = "invalid_quantity"

	// Order
	CodeOrderExists        = "order_exists"
	CodeOrderNotFound      = "order_not_found"
	CodeCreateOrder        = "create_order_failed"
	CodeFetchOrder         = "fetch_order_failed"
	CodeFetchOrders        = "fetch_orders_failed"
	CodeUpdateOrder        = "update_order_failed"
	CodeCancelOrder        = "cancel_order_failed"
	CodeOrderAlreadyClosed = "order_already_cancelled"
	CodeInvalidOrderStatus = "invalid_order_status"

	// Payment
	CodePaymentNotFound         = "payment_not_found"
	CodeFetchPayment            = "fetch_payment_failed"
	CodeCreatePayment           = "create_payment_failed"
	CodeUpdatePayment           = "update_payment_failed"
	CodeGenerateReference       = "generate_reference_failed"
	CodeInvalidWebhookSignature = "invalid_webhook_signature"
	CodeInvalidWebhookEvent     = "invalid_webhook_event"
	CodeWebhookAmountMismatch   = "webhook_amount_mismatch"
	CodeInvalidWebhookStatus    = "invalid_webhook_status"

	// Merchant
	CodeMerchantExists   = "merchant_exists"
	CodeMerchantNotFound = "merchant_not_found"
	CodeCreateMerchant   = "create_merchant_failed"
	CodeFetchMerchant    = "fetch_merchant_failed"
	CodeUpdateMerchant   = "update_merchant_failed"
	CodeFetchDashboard   = "fetch_dashboard_failed"
)
