# Gocart

Gocart is a backend **REST API** for an e-commerce platform built with **Go**, **Gin**, and **GORM**. It features **JWT authentication**, **role-based authorization**, **product** and **category management**, **shopping cart** and **checkout workflows**, and **MinIO integration** for product image storage.

## Key Features

### Authentication and Users

- Register a new **customer**.
- Log in with **email** and **password**.
- Receive a signed **JWT access token**.
- Access the authenticated **profile endpoint**.
- Default role assignment is **customer**.

### Catalog Management

- Browse **products** publicly.
- Filter products by **category**, **price range**, **stock status**, and **search term**.
- Sort products by **id**, **name**, **price**, **created_at**, or **stock**.
- Browse **categories** publicly.
- **Admins** can create, update, and delete categories and products.

### Cart and Checkout

- Create and retrieve a **cart** automatically for authenticated users.
- **Add**, **update**, **remove**, and **clear** cart items.
- Enforce **stock checks** while modifying the cart.
- **Checkout** converts the cart into an order and deducts product stock.
- **Cart totals** and **item counts** are recalculated after cart mutations.

### Orders

- List the current user’s **orders**.
- Fetch **order details** by id.
- **Cancel** an order.

### Image Uploads

- Upload one or more **product images** with product create and update requests.
- Store image objects in **MinIO** under a product-scoped path.

