# Frontend Component Library

## Principles
- Reusable, composable, accessible.
- Server Components by default; client components for interactivity.

## Base & UI Components
- Base: MUI primitives (Button, TextField, Table)
- UI Wrappers: AppButton, AppCard, AppTable to unify styling and variants.

## Feature Components
- WalletCard: Summary with net worth and 24h change.
- AssetHoldings: Paginated asset table with sorting/filtering.
- TransactionHistory: Virtualized list/table with filters.
- FinancialChart: TradingView-based chart with dynamic ranges.

## Props & Usage
- WalletCard: { address, nickname, netWorthUsd, changePct24h }
- AssetHoldings: { items, page, pageSize, onPageChange }
- TransactionHistory: { items, onFilterChange }
- FinancialChart: { data: Array<{ time, value }> }

## Design System
- Spacing scale (4px grid), color tokens, typography ramp.
- Dark/light themes; high-contrast mode options.
- Accessibility: WCAG 2.1 AA; keyboard navigation, focus states, aria-labels.

## Documentation
- Examples in Storybook (future); code samples in /examples.
