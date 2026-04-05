# Domain Model

## Main entities

- User
- Profile
- WalletAccount
- WalletEntry
- ProgressionState
- EntitlementSnapshot
- Ban
- ChallengeRun
- Spin
- DebtState
- Order
- FilmPurchase
- Subscription

## Core invariants

- WalletEntry is append-only.
- ProgressionState is the source of truth for unlocks.
- Ban state is derived from debt and rule engine.
- Product services read snapshot data, they do not invent it.