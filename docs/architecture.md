# Architecture Overview

## Layers

- Presentation: frontend/app, frontend/admin.
- Edge: gateway/bff-go.
- Platform: registry-auth, profile, wallet-ledger, progression-entitlements, notification, leaderboard, catalog, scheduler.
- Product: sharaga, burmalda, barygi, kinoshka, balanda.
- Intelligence: recommendation, search, analytics, interrogation-sim.
- Engines: slot-engine, eta-engine.

## Data flow

Frontend -> Gateway -> Domain services -> Ledger / Progression / Events -> Analytics / Notification / Recommendation.

## Core rule

All unlocks must come from progression-entitlements. All money must come from wallet-ledger.