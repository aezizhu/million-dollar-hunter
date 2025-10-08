# PRD: Web-based On-chain Cryptocurrency Dashboard (Million $ Hunter)

---

## Overview

Million $ Hunter is a secure, responsive, single-user web dashboard for real-time monitoring and analysis of cryptocurrency wallets and token flows. Designed for personal use by the owner/operator, the app prioritizes rapid code translation, explicit modularity, and simple technical mapping for AGI coding agents.

---

## Authentication & Session Management

### UI Elements

* **Login Page:** `/login`

  * Username input field (HTML `<input type="text">`)

  * Password input field (HTML `<input type="password">`)

  * Login button (HTML `<button>`)

  * Error message area (HTML `<div>` for inline error banner)

* **Logout Button:** Shown in header/nav bar or profile menu, persistent on all authenticated routes

### Technical Specifications

* **Static Credentials (MVP):**

  * `username = 'admin'`

  * `password = 'Zz@123456789'`

* **Hardcoded Authentication Logic:**

  * On backend: validate provided credentials against hardcoded values.

  * Passwords handled via secure POST; never exposed in client JS.

  * No registration, reset, or forgotten password flows in MVP.

* **Session Management:**

  * Session created on successful login; JWT or secure server-managed session.

  * Session expiry: 30 minutes idle timeout (auto-logout).

  * On session timeout/invalid session, user is redirected to `/login` with inline error.

  * All authenticated endpoints must verify valid session.

* **Logout Flow:**

  * Logout button triggers session/server-side token invalidation.

  * Redirect to `/login` after success.

* **Login Error Handling:**

  * On invalid username/password, display inline error: `"Incorrect username or password."`

  * Account lockout after 5 consecutive failed attempts (show banner: `"Too many failed logins. Please wait 5 minutes."`)

  * All errors and lock warning shown only as web UI banners (no emails or push).

* **Initial Security:**

  * No third-party auth, biometric, or multifactor for MVP.

---

## Front-End / Back-End Modularity

* **Front-End:** Pure SPA, modular React or Vue components, mapped directly to route/component names.

* **Back-End:** RESTful API/GraphQL with named endpoints, all business logic/auth, API fetch, and alert conditions implemented server-side.

* **Front-End never stores or processes authentication credentials locally except JWT/session token post-login.**

* **Back-End exposes only authenticated APIs for dashboard, wallet data, alerts, and download/export.**

---

## Core Views, Routes, and Components

All routes require valid authentication except for `/login`.

---

## Explicit Functional and UI Requirements

### 1\. Dashboard Analytics and Data Flow

**Data Required (via backend API, polled every 5 seconds):**

* Wallet/Address Balances

* Capital Inflow, Outflow, Net Flow (per-address & total)

* Current token prices and interval price changes (1h/24h)

* Large Transaction list (in/out, configurable threshold)

* Top-50 Holder grouping and activity (per token)

* Real-time data update status

**Backend Endpoints:**

* `POST /api/auth/login` – Auth login

* `POST /api/auth/logout` – Auth logout

* `GET /api/wallets` – List tracked wallets

* `POST /api/wallets` – Add new wallet

* `DELETE /api/wallets/:address` – Remove wallet

* `GET /api/analytics/summary` – Current net worth, inflow, outflow, net flow

* `GET /api/analytics/price/:token/:chain` – Latest & interval price

* `GET /api/analytics/large-tx` – List of large transactions

* `GET /api/analytics/holders/:token/:chain` – Top-50 holders and grouping

***All endpoints require session token in*** `Authorization` ***header.***

**Token Query Input Specification:**

* Support for BSC (Binance Smart Chain) and Solana.

* Input: Contract address (HTML `<input type="text">` with chain selection dropdown).

* On submit, backend validates format and fetches via respective blockchain APIs.

* Error states:

  * Invalid address/contract: Show inline error banner: `"Invalid contract address."`

  * Unsupported chain: `"Selected network not supported."`

  * No API data: `"No data available for this token."`

### 2\. UI Elements/Sections

Dashboard Page (`/dashboard`)

* **Header:** Username label, logout button (persistent)

* **Global Alert Banner:** `<div role="alert">` for system status/errors

* **Token Search Form:** (HTML `<form>`)

  * Input field for contract address

  * Chain/network dropdown (BSC, Solana)

  * Submit button

* **Summary Cards:** For

  * Total portfolio value

  * Capital inflow (last 24h)

  * Capital outflow (last 24h)

  * Net flow (last 24h)

  * Current token price and interval change (1h/24h)

* **Tabs:** (HTML `<nav>`)

  * Overview (default)

  * Capital Flows

  * Token Holders

  * Large Transactions

* **Data Tables:** (HTML `<table>`, one per tab)

  * For each: columns explicitly named (see table below)

  * All tables mobile-scrollable, with export CSV button

* **Export/Download:** Button (HTML `<button>`) next to each data table for CSV export

* **Last Updated Timestamp:** Relative timer, reload button

**Table Example: Large Transactions**

Wallet Manager (`/wallets`)

* **Add Wallet Form**

  * Input for address

  * Network selector

  * Confirm/Add button

  * Inline validation (valid address, not duplicate)

  * Success/error alerts via UI banner

* **Wallet List Table**

  * Address, nickname/tag, network, current balance, manage actions (remove, rename/tag)

* **Remove Wallet**

  * Remove button per row with confirmation (modal/dialog)

  * Success/error message in banner

Alert Setup (`/alerts`)

* **Alert Rule Creator**

  * Condition builder: amount, token, flow direction, price change percent

  * Rule list table with enable/disable and delete

  * Inline test alert button

* **Web UI Alerts**

  * Only web alerts at MVP (live banner at top or toast notification)

  * No email, SMS, or push

  * Alert states: match (triggered event), error (rule configuration), status (test ok)

  * All alerts auto-dismiss after 10 seconds or can be closed manually

Transactions (`/transactions`)

* **Filterable/Sortable Table**

  * Filters: date range, wallet, token, flow

  * Columns: date, wallet, token, type, amount, fee, TX hash (link)

  * Pagination or progressive load

* **Export CSV Button**

* **Empty State:** "No transactions found" banner

Settings (`/settings`)

* **Theme Toggle:** Light/Dark mode

* **Default Currency Selector**

* **Sign Out From All Devices**

* **Security:**

  * Display current session info (IP, browser, last login)

  * "Logout Everywhere" button

  * Banners for session/logout errors

---

## Data Handling & Polling

* All main dashboard data is fetched via authenticated API every 5 seconds (adjustable).

* On API fetch error (timeout >5s):

  * System displays global alert banner: `"Data refresh failed, retrying..."`.

  * After 3 failed attempts, show `"Connection lost. Please check your network or try again later."`

  * All loading and error states clearly visible in UI.

* Empty/No Data:

  * Explicit banner in each section, e.g., "No activity in the last 24 hours."

* Backend must de-duplicate requests from same session to avoid overload.

* Real-time WebSocket/live-update support optional; initial implementation via polling.

---

## Demo User Workflow Examples

**As the owner:**

1. I log in at `/login` using `admin` and `Zz@123456789`.

2. On `/dashboard`, I use the chain/token search to query the latest status of a new contract address (BSC or Solana).

3. The dashboard displays my tracked wallets, capital flows, and price changes with up-to-date cards and tables.

4. In `/wallets`, I add a new address, tag it, or remove an address as needed; validation and success/errors display inline.

5. In `/alerts`, I add an inflow/outflow alert; triggered events appear as a banner/toast in the UI, with a test alert for confirmation.

6. On `/transactions`, I filter by date/token to review recent large transfers and export results to CSV.

7. In `/settings`, I change the app to dark mode and log out everywhere for security.

All errors and feedback (auth, API, data states) appear only via banners or toast notifications in the web UI.

---

## MVP and Technical Scope Summary

* **Single user only:** No registration, roles, or invite logic.

* **No direct trading, no private key capture, no crypto custody.**

* **All persistent state managed server-side and fetched via explicit authenticated API calls.**

* **All navigation, UI elements, and states defined by exact route/component names above.**

* **Code structure and components should support stepwise AGI generation (Devin, Claude, Codex) without ambiguity.**

* **Mobile-responsive:** All UI via uniform CSS grid/flexbox, touch targets >= 44px, font size/min contrast via WCAG 2.1 AA.

---

## MVP Success & Acceptance Criteria

* Login/logout flow works as specified with hardcoded credentials and session timeouts.

* All dashboard data displays with explicit error/empty/loading UI and N=5s polling.

* Token query by contract address supports both BSC and Solana contracts, with clearly defined error handling.

* All alert/event logic is visible only in the web UI (banner/toast), following defined conditions.

* All tables, search forms, tabs, export/download, and theme toggle components are visibly present and function in mobile and desktop sizing.

* Backend and frontend code is explicitly separated, modular, with endpoints and route/component mapping per this document.

---

## Final Notes for Developer

* All technical requirements must be directly and unambiguously mappable to code artifacts.

* No business/user stories, only technical owner workflow examples.

* No external notifications or non-web delivery methods for alerts in MVP.

* Maintain clean, modular, and testable code organization for AGI-assisted rapid iteration and deployment.