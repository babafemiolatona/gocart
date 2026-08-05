package mapper

import (
	"testing"
	"time"

	"gocart/internal/models"
)

func TestUnitToMinorUnits(t *testing.T) {
	tests := []struct {
		name  string
		input float64
		want  int64
	}{
		{"whole dollars", 10, 1000},
		{"cents", 19.99, 1999},
		{"rounding up", 0.005, 1},
		{"zero", 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UnitToMinorUnits(tt.input); got != tt.want {
				t.Errorf("UnitToMinorUnits(%v) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestMinorUnitsToUnit(t *testing.T) {
	tests := []struct {
		name  string
		input int64
		want  float64
	}{
		{"whole dollars", 1000, 10},
		{"cents", 1999, 19.99},
		{"zero", 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MinorUnitsToUnit(tt.input); got != tt.want {
				t.Errorf("MinorUnitsToUnit(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestMinorUnitsRoundTrip(t *testing.T) {
	for _, v := range []float64{0.5, 9.99, 123.45, 999.99} {
		if got := MinorUnitsToUnit(UnitToMinorUnits(v)); got != v {
			t.Errorf("round trip failed for %v: got %v", v, got)
		}
	}
}

func TestToCategoryResponse(t *testing.T) {
	created := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	category := &models.Category{
		ID:          7,
		Name:        "Electronics",
		Description: "Gadgets",
		Slug:        "electronics",
		CreatedAt:   created,
		UpdatedAt:   created,
	}

	resp := ToCategoryResponse(category)

	if resp.ID != 7 || resp.Name != "Electronics" || resp.Description != "Gadgets" || resp.Slug != "electronics" {
		t.Errorf("unexpected category response: %+v", resp)
	}
	if !resp.CreatedAt.Equal(created) || !resp.UpdatedAt.Equal(created) {
		t.Errorf("timestamps not copied: %+v", resp)
	}
}

func TestToProductResponse(t *testing.T) {
	product := &models.Product{
		ID:          3,
		Name:        "Laptop",
		Description: "Fast",
		Price:       149999,
		Stock:       5,
		CategoryID:  2,
		MerchantID:  1,
		Slug:        "laptop",
		Sku:         "LAP-1",
		Images: []models.ProductImage{
			{ID: 1, ImageURL: "/img/laptop.jpg", IsPrimary: true},
		},
	}

	resp := ToProductResponse(product)

	if resp.ID != 3 || resp.Name != "Laptop" || resp.Price != 1499.99 || resp.Stock != 5 {
		t.Errorf("unexpected product response: %+v", resp)
	}
	if resp.Sku != "LAP-1" || resp.Slug != "laptop" || resp.MerchantID != 1 || resp.CategoryID != 2 {
		t.Errorf("fields not copied: %+v", resp)
	}
	if len(resp.Images) != 1 || resp.Images[0].ImageURL != "/img/laptop.jpg" || !resp.Images[0].IsPrimary {
		t.Errorf("images not copied: %+v", resp.Images)
	}
}

func TestToProductResponseNil(t *testing.T) {
	if got := ToProductResponse(nil); got != nil {
		t.Errorf("expected nil response for nil product, got %+v", got)
	}
}

func TestToCartResponse(t *testing.T) {
	cart := &models.Cart{
		ID:        9,
		Total:     2998,
		ItemCount: 2,
		Items: []models.CartItem{
			{
				ID:       1,
				Quantity: 2,
				Price:    1499,
				Product:  models.Product{ID: 5, Name: "Laptop", Price: 1499},
			},
		},
	}

	resp := ToCartResponse(cart)

	if resp.ID != 9 || resp.Total != 29.98 || resp.ItemCount != 2 {
		t.Errorf("unexpected cart response: %+v", resp)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Items))
	}
	item := resp.Items[0]
	if item.Quantity != 2 || item.Price != 14.99 || item.Subtotal != 29.98 {
		t.Errorf("unexpected cart item: %+v", item)
	}
	if item.Product.ID != 5 || item.Product.Name != "Laptop" || item.Product.Price != 14.99 {
		t.Errorf("unexpected cart product: %+v", item.Product)
	}
}

func TestToCartResponsePrimaryImageFallback(t *testing.T) {
	t.Run("uses primary image", func(t *testing.T) {
		product := models.Product{
			Images: []models.ProductImage{
				{ImageURL: "/img/secondary.jpg"},
				{ImageURL: "/img/primary.jpg", IsPrimary: true},
			},
		}
		if got := ToCartProductResponse(product).ImageURL; got != "/img/primary.jpg" {
			t.Errorf("expected primary image, got %q", got)
		}
	})

	t.Run("falls back to first image", func(t *testing.T) {
		product := models.Product{
			Images: []models.ProductImage{
				{ImageURL: "/img/first.jpg"},
			},
		}
		if got := ToCartProductResponse(product).ImageURL; got != "/img/first.jpg" {
			t.Errorf("expected first image fallback, got %q", got)
		}
	})

	t.Run("empty when no images", func(t *testing.T) {
		if got := ToCartProductResponse(models.Product{}).ImageURL; got != "" {
			t.Errorf("expected empty image URL, got %q", got)
		}
	})
}

func TestToOrderDetailsResponse(t *testing.T) {
	order := &models.Order{
		ID:              4,
		Status:          models.OrderStatusConfirmed,
		Total:           1999,
		ShippingAddress: "1 Main St",
		Items: []models.OrderItem{
			{ID: 1, ProductID: 2, ProductName: "Widget", Quantity: 1, Price: 1999},
		},
	}

	resp := ToOrderDetailsResponse(order)

	if resp.ID != 4 || resp.Status != "confirmed" || resp.Total != 19.99 {
		t.Errorf("unexpected order response: %+v", resp)
	}
	if resp.ShippingAddress != "1 Main St" {
		t.Errorf("shipping address not copied: %+v", resp)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(resp.Items))
	}
	item := resp.Items[0]
	if item.ProductName != "Widget" || item.Quantity != 1 || item.Price != 19.99 {
		t.Errorf("unexpected order item: %+v", item)
	}
}

func TestToPaymentResponse(t *testing.T) {
	payment := &models.Payment{
		ID:        6,
		OrderID:   4,
		Reference: "PAY_ABC",
		Amount:    1999,
		Status:    models.PaymentStatusPending,
		Provider:  models.PaymentProviderMock,
	}

	resp := ToPaymentResponse(payment)

	if resp.ID != 6 || resp.OrderID != 4 || resp.Reference != "PAY_ABC" {
		t.Errorf("unexpected payment response: %+v", resp)
	}
	if resp.Amount != 19.99 || resp.Status != "pending" || resp.Provider != "mock" {
		t.Errorf("payment fields not copied: %+v", resp)
	}
}

func TestToUserResponse(t *testing.T) {
	user := &models.User{
		ID:        2,
		Username:  "chris",
		Email:     "chris@example.com",
		FirstName: "Chris",
		LastName:  "Taylor",
		Role:      models.RoleCustomer,
	}

	resp := ToUserResponse(user)

	if resp.ID != 2 || resp.Username != "chris" || resp.Email != "chris@example.com" {
		t.Errorf("unexpected user response: %+v", resp)
	}
	if resp.Role != models.RoleCustomer || resp.FirstName != "Chris" || resp.LastName != "Taylor" {
		t.Errorf("user fields not copied: %+v", resp)
	}
}

func TestToMerchantResponse(t *testing.T) {
	merchant := &models.Merchant{
		ID:           5,
		BusinessName: "Acme",
		Description:  "Sells stuff",
		Phone:        "555-0100",
		LogoURL:      "/logo.png",
		IsVerified:   true,
	}

	resp := ToMerchantResponse(merchant)

	if resp.ID != 5 || resp.BusinessName != "Acme" || resp.Description != "Sells stuff" {
		t.Errorf("unexpected merchant response: %+v", resp)
	}
	if !resp.IsVerified || resp.Phone != "555-0100" || resp.LogoURL != "/logo.png" {
		t.Errorf("merchant fields not copied: %+v", resp)
	}
}

func TestToMerchantRecentOrderResponse(t *testing.T) {
	order := models.Order{
		ID:     1,
		User:   models.User{Email: "buyer@example.com"},
		Status: models.OrderStatusShipped,
		Total:  2500,
		Items:  []models.OrderItem{{}, {}, {}},
	}

	resp := ToMerchantRecentOrderResponse(order)

	if resp.ID != 1 || resp.Customer != "buyer@example.com" || resp.Status != models.OrderStatusShipped {
		t.Errorf("unexpected recent order response: %+v", resp)
	}
	if resp.Total != 25 || resp.ItemCount != 3 {
		t.Errorf("recent order fields not copied: %+v", resp)
	}
}
