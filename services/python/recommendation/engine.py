from __future__ import annotations

import random
from collections import Counter
from typing import Any


class RecommendationEngine:
    """Подбор позиций каталога с учётом уровня, категории и «популярности»."""

    def __init__(self, items: list[dict[str, Any]]) -> None:
        self._items: list[dict[str, Any]] = [dict(x) for x in items]
        cats = Counter(str(x.get("category", "")) for x in self._items)
        total = sum(cats.values()) or 1
        self._cat_share: dict[str, float] = {c: cats[c] / total for c in cats}

    def all_items(self) -> list[dict[str, Any]]:
        return [dict(x) for x in self._items]

    def recommend(
        self,
        user_id: str,
        level: int,
        category: str | None = None,
        limit: int = 10,
    ) -> list[dict[str, Any]]:
        _ = user_id
        rng = random.Random(hash(user_id) & 0xFFFFFFFF)

        candidates: list[dict[str, Any]] = []
        for raw in self._items:
            if int(raw.get("min_level", 0)) > level:
                continue
            if category is not None and str(raw.get("category")) != category:
                continue
            candidates.append(dict(raw))

        scored: list[dict[str, Any]] = []
        for item in candidates:
            min_lv = int(item.get("min_level", 0))
            cat = str(item.get("category", ""))
            pop = self._cat_share.get(cat, 0.0)

            base = 1.0
            reason = "Рекомендуем попробовать"

            if min_lv == level:
                base += 5.0
                reason = "Новинка для вашего уровня"
            elif pop >= max(self._cat_share.values(), default=0.0) * 0.85:
                base += 1.5
                reason = "Популярное"

            pop_boost = 1.0 + 0.6 * pop
            jitter = rng.uniform(0.0, 1.25)
            score = base * pop_boost + jitter

            item["score"] = round(score, 4)
            item["reason"] = reason
            scored.append(item)

        scored.sort(key=lambda x: float(x.get("score", 0.0)), reverse=True)
        return scored[: max(0, limit)]
