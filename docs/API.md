# Overview

**SecureBin** is a server-rendered application using [HTMX](https://four.htmx.org/) for dynamic interactions. There is no JSON API and all routes return HTML. Routes are split into two categories:

- **Pages**: return a full HTML document
- **Actions**: return an HTMX fragment to be swapped into an existing page

Authentication is cookie-based. Protected routes check for valid session cookie and redirect to `/login` if absent.

# Routes

## Pages

### `GET /`

> Paste feed.

**Responses:**

| Status | Description           |
| ------ | --------------------- |
| `200`  | Renders `index.html`  |
| `500`  | Internal server error |

### `GET /{username}`

> Public account view.

**Parameters:**

| Parameter  | Type   | Description         |
| ---------- | ------ | ------------------- |
| `username` | string | User's display name |

**Responses:**

| Status | Description            |
| ------ | ---------------------- |
| `200`  | Renders `account.html` |
| `404`  | User not found         |
| `500`  | Internal server error  |

### `GET /p/{id}`

> Full paste view

**Parameters:**

| Parameter | Type   | Description |
| --------- | ------ | ----------- |
| `id`      | string | ID of paste |

**Responses:**

| Status | Description           |
| ------ | --------------------- |
| `200`  | Renders `paste.html`  |
| `404`  | Paste not found       |
| `500`  | Internal server error |

### `GET /p/new`

> Create paste form

**Responses:**

| Status | Description                               |
| ------ | ----------------------------------------- |
| `200`  | Renders `new_paste.html`                  |
| `302`  | Redirect to `/login` if not authenticated |
| `500`  | Internal server error                     |

### `GET /login`

> Login page

**Responses:**

| Status | Description                      |
| ------ | -------------------------------- |
| `200`  | Renders `login.html`             |
| `302`  | Redirect to `/` if authenticated |
| `500`  | Internal server error            |

### `GET /admin`

> Account management (admin only)

**Responses:**

| Status | Description                               |
| ------ | ----------------------------------------- |
| `200`  | Renders `admin.html`                      |
| `302`  | Redirect to `/login` if not authenticated |
| `403`  | User is not admin                         |
| `500`  | Internal server error                     |

### `GET /admin/{username}`

> Manage specific account (admin only)

**Parameters:**

| Parameter  | Type   | Description         |
| ---------- | ------ | ------------------- |
| `username` | string | User's display name |

**Responses:**

| Status | Description                               |
| ------ | ----------------------------------------- |
| `200`  | Renders `account_manager.html`            |
| `302`  | Redirect to `/login` if not authenticated |
| `403`  | User is not admin                         |
| `404`  | `username` not found                      |
| `500`  | Internal server error                     |

## Actions

### `GET /feed?p={n}`

> Paginated paste cards fragment

**Parameters:**

| Parameter | Type | Description |
| --------- | ---- | ----------- |
| `p`       | int  | Page number |

**Responses:**

| Status | Description                  |
| ------ | ---------------------------- |
| `200`  | Returns `feed.html` fragment |
| `204`  | No more pastes to load       |
| `400`  | Request was not HTMX         |
| `500`  | Internal server error        |

### `POST /p`

> Create a new paste

**Request:** `application/x-www-form-urlencoded`

| Field               | Type   | Required | Description                                             |
| ------------------- | ------ | -------- | ------------------------------------------------------- |
| title               | string | ✗        | Paste title                                             |
| body                | string | ✓        | Paste content                                           |
| encrypted_paste_key | string | ✓        | Paste encryption key wrapped with the user's master key |

**Responses:**

| Status | Description                                   |
| ------ | --------------------------------------------- |
| `200`  | Returns `create_paste_callback.html` fragment |
| `400`  | Request was not HTMX                          |
| `401`  | User not authenticated                        |
| `500`  | Internal server error                         |

### `PUT /p/{id}`

> Replace paste ciphertext

**Parameters:**

| Parameter | Type   | Description |
| --------- | ------ | ----------- |
| `id`      | string | ID of paste |

**Request:** `application/x-www-form-urlencoded`

| Field               | Type   | Required | Description                                             |
| ------------------- | ------ | -------- | ------------------------------------------------------- |
| title               | string | ✓        | Paste title                                             |
| body                | string | ✓        | Paste content                                           |
| encrypted_paste_key | string | ✓        | Paste encryption key wrapped with the user's master key |

**Responses:**

| Status | Description                                   |
| ------ | --------------------------------------------- |
| `200`  | Returns `update_paste_callback.html` fragment |
| `400`  | Request was not HTMX                          |
| `401`  | User not authenticated                        |
| `403`  | User not authorized to edit paste             |
| `404`  | Paste not found                               |
| `500`  | Internal server error                         |

### `DELETE /p/{id}`

> Delete a paste

**Parameters:**

| Parameter | Type   | Description |
| --------- | ------ | ----------- |
| `id`      | string | ID of paste |

**Responses:**

| Status | Description                                   |
| ------ | --------------------------------------------- |
| `200`  | Returns `delete_paste_callback.html` fragment |
| `400`  | Request was not HTMX                          |
| `401`  | User not authenticated                        |
| `403`  | User not authorized to delete paste           |
| `404`  | Paste not found                               |
| `500`  | Internal server error                         |

### `POST /login`

> Login form submission

**Request:** `application/x-www-form-urlencoded`

| Field    | Type   | Required | Description              |
| -------- | ------ | -------- | ------------------------ |
| username | string | ✓        | User's username or email |
| password | string | ✓        | User's password          |

**Responses:**

| Status | Description                                                  |
| ------ | ------------------------------------------------------------ |
| `200`  | Success. Redirects to `/` via `HX-Redirect` header           |
| `400`  | Request was not HTMX                                         |
| `422`  | Returns `login_callback.html` fragment with validation error |
| `500`  | Internal server error                                        |

### `POST /logout`

> End session

**Responses:**

| Status | Description                     |
| ------ | ------------------------------- |
| `200`  | Successfully ended user session |
| `401`  | User not authenticated          |
| `500`  | Internal server error           |

### `POST /admin/register`

> Create a new user (admin only)

**Request:** `application/x-www-form-urlencoded`

| Field    | Type   | Required | Description         |
| -------- | ------ | -------- | ------------------- |
| email    | string | ✓        | New user's email    |
| username | string | ✓        | New user's username |
| password | string | ✓        | New user's password |

**Responses:**

| Status | Description                                    |
| ------ | ---------------------------------------------- |
| `200`  | Returns `register_user_callback.html` fragment |
| `400`  | Request was not HTMX                           |
| `401`  | User not authenticated                         |
| `403`  | User not authorized to create a new user       |
| `409`  | User with email already exists                 |
| `500`  | Internal server error                          |

### `POST /admin/reset-password`

> Reset user password (admin only)

**Request:** `application/x-www-form-urlencoded`

| Field        | Type   | Required | Description              |
| ------------ | ------ | -------- | ------------------------ |
| username     | string | ✓        | User's username or email |
| new_password | string | ✓        | User's new password      |

**Responses:**

| Status | Description                                     |
| ------ | ----------------------------------------------- |
| `200`  | Returns `reset_password_callback.html` fragment |
| `400`  | Request was not HTMX                            |
| `401`  | User not authenticated                          |
| `403`  | User not authorized to reset password           |
| `404`  | User not found                                  |
| `500`  | Internal server error                           |

### `POST /admin/reset-email`

> Reset user email (admin only)

**Request:** `application/x-www-form-urlencoded`

| Field     | Type   | Required | Description              |
| --------- | ------ | -------- | ------------------------ |
| username  | string | ✓        | User's username or email |
| new_email | string | ✓        | User's new email         |

**Responses:**

| Status | Description                                  |
| ------ | -------------------------------------------- |
| `200`  | Returns `reset_email_callback.html` fragment |
| `400`  | Request was not HTMX                         |
| `401`  | User not authenticated                       |
| `403`  | User not authorized to reset email           |
| `404`  | User not found                               |
| `500`  | Internal server error                        |

# Middleware

Middleware is applied per-route in `NewRouter`. Route-level middleware uses the `func(http.HandlerFunc) http.HandlerFunc` signature for clean chaining. The `log` middleware wraps the entire mux and uses `func(http.Handler) http.Handler`

```go
func (h *Handler) NewRouter() http.Handler {
    mux := http.NewServeMux()

    // Pages
    mux.HandleFunc("GET /admin", h.auth(h.admin(h.PageAdmin)))
    //...

    // Actions
    mux.HandleFunc("POST /p", h.auth(h.htmx(h.HandleCreatePaste)))
    //...

    return h.log(mux)
}
```

Middleware is applied innermost-first. The rightmost wrapper runs first. Available middleware:

| Name    | Description                                                                  | Signature                                 |
| ------- | ---------------------------------------------------------------------------- | ----------------------------------------- |
| `log`   | Wraps the entire mux. Logs method, path, and status via `slog`               | `func(http.Handler) http.Handler`         |
| `auth`  | Checks session cookie, returns `401` if absent. Sets user on request context | `func(http.HandlerFunc) http.HandlerFunc` |
| `admin` | Checks user has admin role, returns `403` if not. Must be used after `auth`  | `func(http.HandlerFunc) http.HandlerFunc` |
| `htmx`  | Checks `HX-Request` header, returns 400 if absent                            | `func(http.HandlerFunc) http.HandlerFunc` |
